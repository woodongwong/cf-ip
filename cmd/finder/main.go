package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/woodongwong/cf-ip/pkg/ip"
	"github.com/woodongwong/cf-ip/pkg/tcp"
)

var (
	serverName          = ""
	concurrency         = 10
	portCheckTimeout    = time.Millisecond * 500
	cfProxyCheckTimeout = time.Millisecond * 3000
	asn                 = 0
	httpClient          = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Disable automatic navigation
		},
		Timeout: cfProxyCheckTimeout,
		// Refer to: https://coolshell.cn/articles/22263.html
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName: serverName,
			},
			DisableKeepAlives: true,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := net.Dialer{Timeout: cfProxyCheckTimeout}
				conn, err := d.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				tcpConn, ok := conn.(*net.TCPConn)
				if ok {
					tcpConn.SetLinger(0)
					return tcpConn, nil
				}
				return conn, nil
			},
		},
	}
)

func main() {
	parseArgs()

	ipService := &ip.IPverseASNIPServcie{}
	ipNets, err := ip.GetIPv4CIDRByASN(ipService, asn)

	if err != nil {
		panic(err)
	}

	paramChan := make(chan string)
	resultChan := make(chan string)
	cfResultChan := make(chan string)

	for i := 0; i < concurrency; i++ {
		// Check if the port is open
		go func() {
			for {
				addr, ok := <-paramChan
				if !ok {
					break
				}

				if tcp.IsTCPPortOpen(addr, uint16(443), portCheckTimeout) {
					resultChan <- addr
				}
			}
		}()

		// Check if it is a CF reverse proxy
		go func() {
			for {
				addr, ok := <-resultChan
				if !ok {
					break
				}

				if checkCFReverseProxy(addr) {
					cfResultChan <- addr
				}
			}
		}()
	}

	// Get IP address from IP segment
	go func() {
		defer close(paramChan)
		defer close(resultChan)
		defer close(cfResultChan)

		for _, ipNet := range ipNets {
			startIP, endIP := ip.CIDR2IPRange(ipNet)

			stAddr, err := netip.ParseAddr(startIP.String())
			if err != nil {
				panic(err)
			}

			endAddr, err := netip.ParseAddr(endIP.String())
			if err != nil {
				panic(err)
			}

			currentAddr := stAddr
			for {
				paramChan <- currentAddr.String()
				if currentAddr == endAddr {
					break
				}
				currentAddr = currentAddr.Next()
			}
		}

	}()

	// Output result
	for {
		res, ok := <-cfResultChan
		if !ok {
			break
		}
		fmt.Println(res)
	}
}

func parseArgs() {
	var _portCheckTimeout, _cfProxyCheckTimeout int

	flag.StringVar(&serverName, "s", "", "Server name (required)")
	flag.IntVar(&asn, "asn", 0, "ASN (required)")
	flag.IntVar(&concurrency, "n", 10, "Concurrency, default 10")
	flag.IntVar(&_portCheckTimeout, "t_p", 500, "Port check timeout, default 500(ms)")
	flag.IntVar(&_cfProxyCheckTimeout, "t_cf", 3000, "CF proxy check timeout, default 3000(ms)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if serverName == "" || concurrency == 0 || _portCheckTimeout == 0 || asn == 0 || _cfProxyCheckTimeout == 0 {
		flag.Usage()
		os.Exit(1)
	}

	portCheckTimeout = time.Millisecond * time.Duration(_portCheckTimeout)
	cfProxyCheckTimeout = time.Millisecond * time.Duration(_cfProxyCheckTimeout)

	fmt.Printf("serverName: %s\n", serverName)
	fmt.Printf("concurrency: %d\n", concurrency)
	fmt.Printf("portCheckTimeout: %dms\n", _portCheckTimeout)
	fmt.Printf("cfProxyCheckTimeout: %dms\n", _cfProxyCheckTimeout)
	fmt.Printf("asn: %d\n", asn)
}

func checkCFReverseProxy(addr string) bool {
	req, _ := http.NewRequest("GET", "https://"+addr+"/cdn-cgi/trace", nil)

	req.Host = serverName
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:83.0) Gecko/20100101 Firefox/93.0")

	resp, err := httpClient.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return false
	}

	body := make([]byte, 20+len(serverName))
	if _, err := io.ReadFull(resp.Body, body); err != nil {
		return false
	}

	bodyString := string(body)

	return strings.Contains(bodyString, "h="+serverName)
}
