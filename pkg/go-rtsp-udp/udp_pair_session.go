package go_rtsp_udp

import (
	"fmt"
	"gortsp/pkg"
	"net"
	"regexp"
)

type UdpPairSession struct {
	rtpSess  *net.UDPConn
	rtcpSess *net.UDPConn
}

func makeUdpPairSession(localRtpPort, localRtcpPort uint16, remoteAddr string, remoteRtpPort, remoteRtcpPort uint16) *UdpPairSession {
	re := regexp.MustCompile(":[0-9].*")
	remoteAddr = re.ReplaceAllString(remoteAddr, "")

	srcAddr := net.UDPAddr{IP: net.IPv4zero, Port: int(localRtpPort)}
	srcAddr2 := net.UDPAddr{IP: net.IPv4zero, Port: int(localRtcpPort)}
	dstAddr := net.UDPAddr{IP: net.ParseIP(remoteAddr), Port: int(remoteRtpPort)}
	dstAddr2 := net.UDPAddr{IP: net.ParseIP(remoteAddr), Port: int(remoteRtcpPort)}
	pkg.Logger.Info(fmt.Sprintf("[makeUdpPairSession] rtp: src: %s, dst: %s", &srcAddr, &dstAddr))
	pkg.Logger.Info(fmt.Sprintf("[makeUdpPairSession] rtcp: src: %s, dst: %s", &srcAddr2, &dstAddr2))

	//rtpUdpsess, _ := net.DialUDP("udp4", &srcAddr, &dstAddr)
	//rtcpUdpsess, _ := net.DialUDP("udp4", &srcAddr2, &dstAddr2)

	rtpUdpsess, _ := net.ListenUDP("udp4", &srcAddr)
	rtcpUdpsess, _ := net.ListenUDP("udp4", &srcAddr2)

	return &UdpPairSession{
		rtpSess:  rtpUdpsess,
		rtcpSess: rtcpUdpsess,
	}
}
