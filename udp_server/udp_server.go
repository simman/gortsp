package udp_server

import (
	"fmt"
	"github.com/pion/rtp"
	"gortsp/pkg"
	"net"
)

type OnRecBuf func(buf []byte)

func CreateUdpSessionConn(port int, callback OnRecBuf) (*net.UDPConn, error) {
	addr := net.UDPAddr{IP: net.IPv4zero, Port: port}
	session, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		return nil, err
	}

	go func() {
		buf := make([]byte, 1500)
		pkg.Logger.Debug("[Read RTP Data]")

		for {
			select {
			default:
				r, err := session.Read(buf)
				if err != nil {
					pkg.Logger.Error(fmt.Sprintf("[Read RTP Data], error: %s", err.Error()))
					break
				}

				if callback != nil {
					callback(buf[:r])
				}
			}
		}
	}()

	return session, nil
}

func RunUdpServer() {
	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -c:v libx264 -preset fast -c:a libopus udp://localhost:18888
	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -c:v libx264 -preset fast -c:a libopus -f rtp 'rtp://127.0.0.1:18888'

	// ffmpeg -f lavfi -i 'sine=frequency=1000' -c:a libopus -b:a 48000 -sample_fmt s16p -ssrc 1 -payload_type 111 -f rtp -max_delay 0 -application lowdelay 'rtp://127.0.0.1:18888'
	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -c:v libx264 -preset fast -f rtp 'rtp://127.0.0.1:18888'

	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -c:v libx264 -f rtsp 'rtsp://192.168.5.222:18000' -preset fast -c:a libopus -f rtsp 'rtsp://192.168.5.222:18001'

	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -c:v libx264 -f rtp 'rtp://192.168.5.222:18000' -preset fast -c:a libopus -f rtp 'rtp://192.168.5.222:18001'

	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -vcodec copy -an -f rtp rtp://127.0.0.1:18000 -vn -acodec copy -f rtp rtp://127.0.0.1:18001

	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -vcodec copy -an -f rtp rtp://127.0.0.1:18000 -vn -c:a libopus -f rtp rtp://127.0.0.1:18001

	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -vcodec copy -an -f rtp rtp://127.0.0.1:18000 -vn -c:a libopus -page_duration 20000 -f rtp rtp://127.0.0.1:18001

	// ffmpeg -i rtsp://192.168.5.222:8554/live/test110/test110 -c:v libx264 -an -f rtp rtp://127.0.0.1:18000 -vn -c:a libopus -b:a 32k -page_duration 20000 -f rtp rtp://127.0.0.1:18001

	go func() {
		_, _ = CreateUdpSessionConn(18000, func(buf []byte) {
			// buf[:r]
			//pkg.Logger.Info("收到RTP数据", "video", "len", len(buf))
			packer := rtp.Packet{}
			err := packer.Unmarshal(buf)
			if err != nil {
				pkg.Logger.Error("解析rtp失败.....")
			} else {
				pkg.Logger.Info("rtp info", "payloadType", packer.PayloadType)
			}
		})

		//select {}
	}()

	go func() {
		_, _ = CreateUdpSessionConn(18001, func(buf []byte) {
			// buf[:r]
			//pkg.Logger.Info("收到RTP数据", "audio", "len", len(buf))
			//packer := rtp.Packet{}
			//err := packer.Unmarshal(buf)
			//if err != nil {
			//	pkg.Logger.Error("解析rtp失败.....")
			//} else {
			//	pkg.Logger.Info("rtp info", "payloadType", packer.PayloadType)
			//}
			//packer := rtp.H264UnPacker{}
			//_ = packer.UnPack(buf)
			//packer.OnFrame(func(frame []byte, timestamp uint32, lost bool) {
			//	pkg.Logger.Info("======== 解析成功!!!!")
			//})
		})

		select {}
	}()

	select {}
}
