package rtp_forwarder

import (
	"context"
	"fmt"
	"gortsp/pkg"
	"gortsp/utils"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
)

type RtpForwarderOnAudioBuf func(buf []byte)
type RtpForwarderOnVideoBuf func(buf []byte)

type RtpForwarder struct {
	rtspAddress string

	audioUdpConn *net.UDPConn
	videoUdpConn *net.UDPConn

	onAudioBuf RtpForwarderOnAudioBuf
	onVideoBuf RtpForwarderOnVideoBuf

	onClose func()

	die chan struct{}
}

type RtpForwarderOptions struct {
	EnableVideo      bool // 是否开启视频
	EnableAudio      bool // 是否开启音频
	VideoTranscoding bool // 是否需要转码视频
	AudioTranscoding bool // 是否需要转码音频
}

var ff_bin_path string

func init() {
	// 处理ffmpeg
	dir, err := os.UserConfigDir()
	if err != nil {
		panic("获取用户配置目录失败!")
	}

	binPath := path.Join(dir, "gortsp", "bin")
	if _, err := os.Stat(binPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(binPath, 0777)
		}
	}

	binExt := ""
	if runtime.GOOS == "windows" {
		binExt = ".exe"
	}

	ff_bin_path = path.Join(binPath, fmt.Sprintf("goff%s", binExt))
	pkg.Logger.Info("ff-bin", "location", ff_bin_path)
	if _, err := os.Stat(ff_bin_path); err != nil {
		if os.IsNotExist(err) {

			f, err := pkg.FFBin.Open(fmt.Sprintf("ffgo/ffmpeg/static/ffmpeg%s", binExt))
			if err != nil {
				panic(err)
			}
			defer f.Close()

			ff, err := os.OpenFile(ff_bin_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				panic(err)
			}

			if _, err := io.Copy(ff, f); err != nil {
				panic(err)
			}
		}
	}
}

func NewRtpForwarder(addr string, onAudioBuf RtpForwarderOnAudioBuf, onVideoBuf RtpForwarderOnVideoBuf, onClose func()) *RtpForwarder {
	return &RtpForwarder{
		rtspAddress: addr,
		onAudioBuf:  onAudioBuf,
		onVideoBuf:  onVideoBuf,
		onClose:     onClose,
		die:         make(chan struct{}),
	}
}

func (f *RtpForwarder) Start(opts RtpForwarderOptions) error {
	pkg.Logger.Info("RtpForwarder", "开启转码服务")
	aPort, err := utils.GetFreePort("udp")
	vPort, err := utils.GetFreePort("udp")
	if err != nil {
		pkg.Logger.Error("没有可用udp端口", "err", err.Error())
		return err
	}

	// 创建udp监听
	pkg.Logger.Info("RtpForwarder", "创建udp监听", "audio-port", aPort, "video-port", vPort)
	if opts.EnableAudio {
		f.audioUdpConn, err = f._createUdpSessionConn(aPort, f.onAudioBuf)
	}

	if opts.EnableVideo {
		f.videoUdpConn, err = f._createUdpSessionConn(vPort, f.onVideoBuf)
	}

	if err != nil {
		pkg.Logger.Error("UDP", "创建udp服务失败", "err", err.Error())
		return err
	}

	//cmd := fmt.Sprintf("ffmpeg -hide_banner -i %s -c:v libx264 -an -f rtp rtp://127.0.0.1:%d -vn -c:a libopus -b:a 32k -page_duration 20000 -f rtp rtp://127.0.0.1:%d")
	//
	var ff_cmd_args = make([]string, 0)
	ff_cmd_args = append(ff_cmd_args, "-hide_banner")
	ff_cmd_args = append(ff_cmd_args, "-i")
	ff_cmd_args = append(ff_cmd_args, f.rtspAddress) // 设置地址
	ff_cmd_args = append(ff_cmd_args, "-tune")
	ff_cmd_args = append(ff_cmd_args, "zerolatency") // 零缓冲
	ff_cmd_args = append(ff_cmd_args, "-preset")     // 极速转码
	ff_cmd_args = append(ff_cmd_args, "ultrafast")   // 极速转码

	//ff_cmd_args = append(ff_cmd_args, "-threads")
	//ff_cmd_args = append(ff_cmd_args, "5")

	//ff_cmd_args = append(ff_cmd_args, "-fflags")
	//ff_cmd_args = append(ff_cmd_args, "nobuffer")

	// 通过添加 -fflags +genpts 选项到您的FFmpeg命令中，可以强制生成时间戳。这有助于确保首帧能够及时显示。
	//ff_cmd_args = append(ff_cmd_args, "-fflags")
	//ff_cmd_args = append(ff_cmd_args, "+igndts")

	//ff_cmd_args = append(ff_cmd_args, "-vsync")
	//ff_cmd_args = append(ff_cmd_args, " 0")

	// 添加 -analyzeduration 和 -probesize 选项可以增加解码器的缓冲区大小，以便更快地获取首帧
	//ff_cmd_args = append(ff_cmd_args, "-analyzeduration")
	//ff_cmd_args = append(ff_cmd_args, "10M")
	//ff_cmd_args = append(ff_cmd_args, "-probesize")
	//ff_cmd_args = append(ff_cmd_args, "10M")

	if opts.EnableVideo {
		if opts.VideoTranscoding { // 如果需要转码, 则标准输出为 h264
			ff_cmd_args = append(ff_cmd_args, "-c:v")
			ff_cmd_args = append(ff_cmd_args, "libx264")
		} else {
			ff_cmd_args = append(ff_cmd_args, "-c:v")
			ff_cmd_args = append(ff_cmd_args, "copy")
		}
		ff_cmd_args = append(ff_cmd_args, "-an") // 忽略音频

		//ff_cmd_args = append(ff_cmd_args, "-b:v")
		//ff_cmd_args = append(ff_cmd_args, "1M")

		ff_cmd_args = append(ff_cmd_args, "-f")
		ff_cmd_args = append(ff_cmd_args, "rtp")
		//ff_cmd_args = append(ff_cmd_args, "-max_delay")
		//ff_cmd_args = append(ff_cmd_args, "0")
		//ff_cmd_args = append(ff_cmd_args, "-application")
		//ff_cmd_args = append(ff_cmd_args, "lowdelay")
		ff_cmd_args = append(ff_cmd_args, fmt.Sprintf("rtp://127.0.0.1:%d", vPort)) // 转发视频
	}

	if opts.EnableAudio {
		if opts.AudioTranscoding { // 如果需要转码, 则标准输出为 opus
			ff_cmd_args = append(ff_cmd_args, "-c:a")
			ff_cmd_args = append(ff_cmd_args, "libopus")
			//ff_cmd_args = append(ff_cmd_args, "-reorder_queue_size")
			//ff_cmd_args = append(ff_cmd_args, "1024")
			ff_cmd_args = append(ff_cmd_args, "-b:a") // 音频码率
			ff_cmd_args = append(ff_cmd_args, " 128k")
		} else {
			ff_cmd_args = append(ff_cmd_args, "-c:a")
			ff_cmd_args = append(ff_cmd_args, "copy")
		}
		ff_cmd_args = append(ff_cmd_args, "-vn") // 忽略视频
		ff_cmd_args = append(ff_cmd_args, "-f")  // 忽略视频
		ff_cmd_args = append(ff_cmd_args, "rtp") // 忽略视频
		//ff_cmd_args = append(ff_cmd_args, "-max_delay")
		//ff_cmd_args = append(ff_cmd_args, "0")
		//ff_cmd_args = append(ff_cmd_args, "-application")
		//ff_cmd_args = append(ff_cmd_args, "lowdelay")
		ff_cmd_args = append(ff_cmd_args, fmt.Sprintf("rtp://127.0.0.1:%d", aPort)) // 转发音频
	}

	ctx, cancel := context.WithCancel(context.Background())
	ff := exec.CommandContext(ctx, ff_bin_path, ff_cmd_args...)

	pkg.Logger.Info("ffmpeg", "cmd", ff.String())

	go func() {
		select {
		case <-f.die:
			cancel()
			return
		}
	}()

	log := pkg.NewLogWriter()
	ff.Stdout = log
	ff.Stderr = log

	if err = ff.Start(); err != nil {
		pkg.Logger.Error("FF", "error", err.Error())
		return err
	}

	if err = ff.Wait(); err != nil {
		pkg.Logger.Error("FF", "error", err.Error())
		return nil
	}

	return nil
}

func (f *RtpForwarder) Stop() {
	close(f.die)

	if f.audioUdpConn != nil {
		_ = f.audioUdpConn.Close()
	}

	if f.videoUdpConn != nil {
		_ = f.videoUdpConn.Close()
	}
}

func (f *RtpForwarder) _createUdpSessionConn(port int, callback func(buf []byte)) (*net.UDPConn, error) {
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
			case <-f.die:
				pkg.Logger.Warn("udp connection closed, exit rtp goroutine")
				return
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
