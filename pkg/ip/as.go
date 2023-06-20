package ip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 120,
	Transport: &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: time.Second * 30,
	},
}

type IPService interface {
	GetIPv4CIDRByASN(asn int) ([]*net.IPNet, error)
}

func GetIPv4CIDRByASN(ipService IPService, asn int) ([]*net.IPNet, error) {
	return ipService.GetIPv4CIDRByASN(asn)
}

type IPIPNetService struct{}

func (s *IPIPNetService) GetIPv4CIDRByASN(asn int) ([]*net.IPNet, error) {

	resp, err := httpClient.Get("http://whois.ipip.net/AS" + strconv.Itoa(asn))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`>(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,3})<`)
	regexpRes := re.FindAllSubmatch(body, -1)
	ipNetList := make([]*net.IPNet, 0, len(regexpRes))

	for _, val := range regexpRes {
		if len(val) > 1 {
			_, ipNet, err := net.ParseCIDR(string(val[1]))
			if err == nil {
				isAppend := true
				for index, item := range ipNetList {
					if IsSubnet(ipNet, item) {
						isAppend = false
						break
					}

					if IsSubnet(item, ipNet) {
						ipNetList[index] = ipNet
						isAppend = false
						break
					}
				}

				if isAppend {
					ipNetList = append(ipNetList, ipNet)
				}
			}
		}
	}

	return ipNetList, nil
}

type IPverseASNIPServcie struct{}

func (s *IPverseASNIPServcie) GetIPv4CIDRByASN(asn int) ([]*net.IPNet, error) {
	resp, err := httpClient.Get(fmt.Sprintf("https://raw.githubusercontent.com/ipverse/asn-ip/master/as/%d/ipv4-aggregated.txt", asn))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyArr := regexp.MustCompile(`\r?\n`).Split(string(body), -1)
	ipNetList := make([]*net.IPNet, 0, len(bodyArr))

	for _, val := range bodyArr {
		_, ipNet, err := net.ParseCIDR(val)
		if err == nil {
			ipNetList = append(ipNetList, ipNet)
		}
	}
	return ipNetList, nil
}
