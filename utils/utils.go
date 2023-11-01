package utils

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func Bytes2Bits(data []byte) []int {
	dst := make([]int, 0)
	for _, v := range data {
		for i := 0; i < 8; i++ {
			move := uint(7 - i)
			dst = append(dst, int((v>>move)&1))
		}
	}
	//fmt.Println(len(dst))
	return dst
}

func Bytes2HexString(data []byte) string {
	hexStr := hex.EncodeToString(data)

	var tmp string
	for i, s := range strings.Split(hexStr, "") {
		if i%2 == 0 {
			tmp += fmt.Sprintf("%s", s)
		} else {
			tmp += fmt.Sprintf("%s ", s)
		}
	}
	return tmp
}

// GetFreePort Get an available port.
// [network] tcp/udp
func GetFreePort(network string) (port int, err error) {
	switch network {
	case "tcp":
		if a, err := net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
			var l *net.TCPListener
			if l, err = net.ListenTCP("tcp", a); err == nil {
				defer func(l *net.TCPListener) {
					_ = l.Close()
				}(l)
				return l.Addr().(*net.TCPAddr).Port, nil
			}
		}
	case "udp":
		if a, err := net.ResolveUDPAddr("udp", "localhost:0"); err == nil {
			var l *net.UDPConn
			if l, err = net.ListenUDP("udp", a); err == nil {
				defer func(l *net.UDPConn) {
					_ = l.Close()
				}(l)
				_, port, err := net.SplitHostPort(l.LocalAddr().String())
				if err != nil {
					return 0, err
				}
				portInt, _ := strconv.Atoi(port)
				return portInt, nil
			}
		}
	default:
		return 0, net.UnknownNetworkError(network)
	}

	return
}
