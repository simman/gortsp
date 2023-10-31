package go_rtsp_udp

import (
	"fmt"
	"gortsp/pkg"
	"gortsp/pkg/go-rtsp"
	"net"
	"net/url"
)

type RtspUdpServer struct {
	rtspAddress string
	session     *RtspUdpPlaySession
	rtspClient  *rtsp.RtspClient
	conn        net.Conn
}

func NewRtspUdpServer(rtspAddress string, onRtpBufCallback OnRtpBufCallback, onSampleCallback OnSampleCallback, onCloseCallback OnCloseCallback) (*RtspUdpServer, error) {
	pkg.Logger.Debug(fmt.Sprintf("开始加载: %s", rtspAddress))

	u, err := url.Parse(rtspAddress)
	if err != nil {
		panic(err)
	}
	host := u.Host
	if u.Port() == "" {
		host += ":554"
	}
	pkg.Logger.Debug(fmt.Sprintf("host: %s", host))
	c, err := net.Dial("tcp4", host)
	if err != nil {
		pkg.Logger.Error(err.Error())
		return nil, err
	}

	sess := NewRtspUdpPlaySession(c).WithCallback(onRtpBufCallback, onSampleCallback, onCloseCallback)
	client, _ := rtsp.NewRtspClient(rtspAddress, sess)

	return &RtspUdpServer{
		rtspAddress: rtspAddress,
		session:     sess,
		rtspClient:  client,
		conn:        c,
	}, nil
}

func (s *RtspUdpServer) Start() {
	sc := make(chan []byte, 100)
	go s.session.sendInLoop(sc)

	s.rtspClient.SetOutput(func(b []byte) error {
		if s.session.lastError != nil {
			return s.session.lastError
		}
		sc <- b
		return nil
	})
	if err := s.rtspClient.Start(); err != nil {
		pkg.Logger.Error(err.Error())
	}

	buf := make([]byte, 4096)
	for {
		n, err := s.conn.Read(buf)
		if err != nil {
			pkg.Logger.Error("Read Buffer Error", "err", err.Error())
			break
		}
		if err = s.rtspClient.Input(buf[:n]); err != nil {
			pkg.Logger.Error("Input Buffer Error", "err", err.Error())
			break
		}
	}
	s.session.Destory()
}

func (s *RtspUdpServer) Close() {
	s.session.Destory()
}
