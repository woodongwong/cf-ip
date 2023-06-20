package tcp

import (
	"net"
	"strconv"
	"time"
)

func IsTCPPortOpen(host string, port uint16, timeout time.Duration) bool {

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))), timeout)

	if err != nil {
		return false
	}

	defer conn.Close()

	if tcpCon, ok := conn.(*net.TCPConn); ok {
		tcpCon.SetLinger(0)
	}

	return true
}
