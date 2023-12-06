package go_rtsp_udp

import (
	"errors"
	"fmt"
	"gortsp/pkg"
	"gortsp/pkg/go-rtsp"
	"gortsp/pkg/go-rtsp/sdp"
	"gortsp/utils"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type OnCloseCallback func()
type OnRtpBufCallback func(buf []byte)
type OnSampleCallback func(sample rtsp.RtspSample)

type RtspUdpPlaySession struct {
	udpport    uint16
	videoFile  *os.File
	audioFile  *os.File
	tsFile     *os.File
	timeout    int
	once       sync.Once
	die        chan struct{}
	c          net.Conn
	lastError  error
	sesss      map[string]*UdpPairSession
	remoteAddr string

	//-------
	rtpBufCallback  OnRtpBufCallback
	sampleCallback  OnSampleCallback
	onCloseCallback OnCloseCallback
}

func NewRtspUdpPlaySession(c net.Conn) *RtspUdpPlaySession {
	udpPort, err := utils.GetFreePort("udp")
	if err != nil {
		pkg.Logger.Error("Get free udp port", "err", err.Error())
		panic(err)
	}
	return &RtspUdpPlaySession{udpport: uint16(udpPort), die: make(chan struct{}), c: c, sesss: make(map[string]*UdpPairSession)}
}

func (cli *RtspUdpPlaySession) WithCallback(onRtpBufCallback OnRtpBufCallback, onSampleCallback OnSampleCallback, onCloseCallback OnCloseCallback) *RtspUdpPlaySession {
	cli.rtpBufCallback = onRtpBufCallback
	cli.sampleCallback = onSampleCallback
	cli.onCloseCallback = onCloseCallback
	return cli
}

func (cli *RtspUdpPlaySession) HandleOption(client *rtsp.RtspClient, res rtsp.RtspResponse, public []string) error {
	pkg.Logger.Debug("[HandleOption]", "public", public)
	return nil
}

func (cli *RtspUdpPlaySession) HandleSetup(client *rtsp.RtspClient, res rtsp.RtspResponse, track *rtsp.RtspTrack, tracks map[string]*rtsp.RtspTrack, sessionId string, timeout int) error {
	pkg.Logger.Debug("[HandleSetup]", "SessionId", sessionId, "timeout", timeout)
	if res.StatusCode == rtsp.Unsupported_Transport {
		return errors.New("unsupport udp transport")
	}
	if res.StatusCode == rtsp.Internal_Server_Error {
		return errors.New("internal server error")
	}
	cli.sesss[track.TrackName] = makeUdpPairSession(track.GetTransport().Client_ports[0], track.GetTransport().Client_ports[1], cli.c.RemoteAddr().String(), track.GetTransport().Server_ports[0], track.GetTransport().Server_ports[1])
	track.OnPacket(func(b []byte, isRtcp bool) (err error) {
		if isRtcp {
			_, err = cli.sesss[track.TrackName].rtcpSess.Write(b)
		}
		return
	})

	// 读取rtp数据
	go func() {
		buf := make([]byte, 1500)
		pkg.Logger.Debug("[Read RTP Data]")
		for {
			select {
			case <-cli.die:
				pkg.Logger.Warn("udp connection closed, exit rtp goroutine")
				return
			default:
				r, err := cli.sesss[track.TrackName].rtpSess.Read(buf)
				if err != nil {
					pkg.Logger.Error(fmt.Sprintf("[Read RTP Data], error: %s", err.Error()))
					break
				}

				if cli.rtpBufCallback != nil {
					cli.rtpBufCallback(buf[:r])
				}

				err = track.Input(buf[:r], false)
				if err != nil {
					pkg.Logger.Error(fmt.Sprintf("[Read RTP Data], Track Input error: %s", err.Error()))
					break
				}
			}
		}
	}()

	// 读取rtcp数据
	go func() {
		buf := make([]byte, 1500)
		for {
			select {
			case <-cli.die:
				pkg.Logger.Warn("udp connection closed, exit rtcp goroutine")
				return
			default:
				r, err := cli.sesss[track.TrackName].rtcpSess.Read(buf)
				if err != nil {
					pkg.Logger.Error(err.Error())
					break
				}

				err = track.Input(buf[:r], true)
				if err != nil {
					pkg.Logger.Error(err.Error())
					break
				}
			}
		}
	}()

	cli.timeout = timeout
	return nil
}

func (cli *RtspUdpPlaySession) HandleDescribe(client *rtsp.RtspClient, res rtsp.RtspResponse, sdpx *sdp.Sdp, tracks map[string]*rtsp.RtspTrack) error {
	pkg.Logger.Info("[HandleDescribe]", "StatusCode", res.StatusCode, "Reason", res.Reason)

	if res.StatusCode != 200 {
		panic(res.Reason)
	}

	var sps []byte
	var pps []byte

	for _, media := range sdpx.Medias {
		tmp := strings.ToLower(media.EncodeName)
		switch tmp {
		case "h264":
			fmtpHandle := sdp.NewH264FmtpParam()
			fmtpHandle.Load(media.Attrs["fmtp"])
			sps, pps = fmtpHandle.GetSpsPps()
			pkg.Logger.Debug("[HandleDescribe]", "sps", sps, "pps", pps)
		}
	}

	for k, t := range tracks {
		if t == nil {
			continue
		}

		pkg.Logger.Info(fmt.Sprintf("Got [%s] Track", k))
		//transport := rtsp.NewRtspTransport(rtsp.WithEnableUdp(), rtsp.WithClientUdpPort(cli.udpport, cli.udpport+1))
		transport := rtsp.NewRtspTransport(rtsp.WithEnableUdp(), rtsp.WithClientUdpPort(cli.udpport, cli.udpport+1), rtsp.WithMode("")) // rtsp.MODE_PLAY
		t.SetTransport(transport)
		t.OpenTrack()
		cli.udpport += 2
		if t.Codec.Cid == rtsp.RTSP_CODEC_H264 {
			if cli.videoFile == nil {
				//cli.videoFile, _ = os.OpenFile("test_data/h264_file/video_full.h264", os.O_CREATE|os.O_RDWR, 0666)
				//cli.videoFile.Write(sps)
				//cli.videoFile.Write(pps)
			}

			t.OnSample(func(sample rtsp.RtspSample) {
				//slog.Debug("【5】[OnSample] Got H264", "frameLen", len(sample.Sample), "timestamp", sample.Timestamp)

				if cli.sampleCallback != nil {
					cli.sampleCallback(sample)
				}

				//err := videoTrack.WriteSample(mediaSample)
				//if err != nil {
				//	log.Println("WriteSample ERROR", err)
				//}
			})
		} else if t.Codec.Cid == rtsp.RTSP_CODEC_AAC {
			if cli.audioFile == nil {
				//cli.audioFile, _ = os.OpenFile("test_data/audio.aac", os.O_CREATE|os.O_RDWR, 0666)
			}
			t.OnSample(func(sample rtsp.RtspSample) {
				//pkg.Logger.Debug("【5】[OnSample] Got AAC", "frameLen", len(sample.Sample), "timestamp", sample.Timestamp)
				//cli.audioFile.Write(sample.Sample)
				//if cli.sampleCallback != nil {
				//	cli.sampleCallback(sample)
				//}
			})
		} else if t.Codec.Cid == rtsp.RTSP_CODEC_TS {
			if cli.tsFile == nil {
				//cli.tsFile, _ = os.OpenFile("test_data/mp2t.ts", os.O_CREATE|os.O_RDWR, 0666)
			}
			t.OnSample(func(sample rtsp.RtspSample) {
				pkg.Logger.Debug("【5】[OnSample] Got TS", "frameLen", len(sample.Sample), "timestamp", sample.Timestamp)
				//cli.tsFile.Write(sample.Sample)
			})
		} else if t.Codec.Cid == rtsp.RTSP_CODEC_H265 {
			if cli.videoFile == nil {
				//cli.videoFile, _ = os.OpenFile("video_full.h265", os.O_CREATE|os.O_RDWR, 0666)
				//cli.videoFile.Write(sps)
				//cli.videoFile.Write(pps)
			}

			t.OnSample(func(sample rtsp.RtspSample) {
				//slog.Debug("【5】[OnSample] Got H265", "frameLen", len(sample.Sample), "timestamp", sample.Timestamp)
				//cli.videoFile.Write(sample.Sample)

				if cli.sampleCallback != nil {
					//cli.sampleCallback(sample)
				}
			})
		}
	}
	return nil
}

func (cli *RtspUdpPlaySession) HandlePlay(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
	pkg.Logger.Info("[HandlePlay]", "res", res.Reason)
	if res.StatusCode != 200 {
		pkg.Logger.Error("play failed ", res.StatusCode, res.Reason)
		return nil
	}
	go func() {
		//rtsp keepalive
		to := time.NewTicker(time.Duration(cli.timeout/2) * time.Second)
		defer to.Stop()
		for {
			select {
			case <-to.C:
				pkg.Logger.Info("[Send KeepAlive]...")
				_ = client.KeepAlive(rtsp.OPTIONS)
			case <-cli.die:
				return
			}
		}
	}()
	return nil
}

func (cli *RtspUdpPlaySession) HandleAnnounce(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
	pkg.Logger.Info("[HandleAnnounce]")
	return nil
}

func (cli *RtspUdpPlaySession) HandlePause(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
	pkg.Logger.Info("[HandlePause]")
	return nil
}

func (cli *RtspUdpPlaySession) HandleTeardown(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
	pkg.Logger.Info("[HandleTeardown]")
	return nil
}

func (cli *RtspUdpPlaySession) HandleGetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
	pkg.Logger.Info("[HandleGetParameter]")
	return nil
}

func (cli *RtspUdpPlaySession) HandleSetParameter(client *rtsp.RtspClient, res rtsp.RtspResponse) error {
	pkg.Logger.Info("[HandleSetParameter]")
	return nil
}

func (cli *RtspUdpPlaySession) HandleRedirect(client *rtsp.RtspClient, req rtsp.RtspRequest, location string, timeRange *rtsp.RangeTime) error {
	pkg.Logger.Info("[HandleRedirect]")
	return nil
}

func (cli *RtspUdpPlaySession) HandleRecord(client *rtsp.RtspClient, res rtsp.RtspResponse, timeRange *rtsp.RangeTime, info *rtsp.RtpInfo) error {
	pkg.Logger.Info("[HandleRecord]")
	return nil
}

func (cli *RtspUdpPlaySession) HandleRequest(client *rtsp.RtspClient, req rtsp.RtspRequest) error {
	pkg.Logger.Info("[HandleRequest]")
	return nil
}

func (cli *RtspUdpPlaySession) Destory() {
	pkg.Logger.Info("[Destory]")
	cli.once.Do(func() {
		if cli.onCloseCallback != nil {
			cli.onCloseCallback()
		}
		if cli.videoFile != nil {
			_ = cli.videoFile.Close()
		}
		if cli.audioFile != nil {
			_ = cli.audioFile.Close()
		}
		if cli.tsFile != nil {
			_ = cli.tsFile.Close()
		}

		// close all udp conn
		for _, v := range cli.sesss {
			if v != nil {
				if v.rtpSess != nil {
					_ = v.rtpSess.Close()
				}
				if v.rtcpSess != nil {
					_ = v.rtcpSess.Close()
				}
			}
		}
		_ = cli.c.Close()
		close(cli.die)
	})
}

func (cli *RtspUdpPlaySession) sendInLoop(sendChan chan []byte) {
	for {
		select {
		case b := <-sendChan:
			_, err := cli.c.Write(b)
			if err != nil {
				cli.Destory()
				cli.lastError = err
				pkg.Logger.Debug("[SendInLoop] quit send in loop")
				return
			}

		case <-cli.die:
			pkg.Logger.Debug("[SendInLoop] quit send in loop")
			return
		}
	}
}
