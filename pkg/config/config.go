package config

import (
	go_rtsp_udp "gortsp/pkg/go-rtsp-udp"
	rtp_forwarder "gortsp/pkg/rtp-forwarder"
	webrtc_server "gortsp/pkg/webrtc-server"
	"sync"
)

var Config = &ConfigST{
	rtspServers:   make(map[string]*go_rtsp_udp.RtspUdpServer),
	webrtcServers: make(map[string]*webrtc_server.WebRTCServer),
	rtpForwarders: make(map[string]*rtp_forwarder.RtpForwarder),
}

type ConfigST struct {
	mutex         sync.RWMutex
	rtspServers   map[string]*go_rtsp_udp.RtspUdpServer
	webrtcServers map[string]*webrtc_server.WebRTCServer
	rtpForwarders map[string]*rtp_forwarder.RtpForwarder
}

// ------->> rtsp

func (config *ConfigST) RtspAdd(suuid string, server *go_rtsp_udp.RtspUdpServer) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	config.rtspServers[suuid] = server
}

func (config *ConfigST) RtspGet(suuid string) *go_rtsp_udp.RtspUdpServer {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	return config.rtspServers[suuid]
}

func (config *ConfigST) RtspRm(suuid string) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	delete(config.rtspServers, suuid)
}

func (config *ConfigST) RtspCloseAndRm(suuid string) {
	if o := config.RtspGet(suuid); o != nil {
		o.Close()
		config.RtspRm(suuid)
	}
}

// ------->> rtpForwarder
func (config *ConfigST) RtpForwarderAdd(suuid string, server *rtp_forwarder.RtpForwarder) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	config.rtpForwarders[suuid] = server
}

func (config *ConfigST) RtpForwarderGet(suuid string) *rtp_forwarder.RtpForwarder {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	return config.rtpForwarders[suuid]
}

func (config *ConfigST) RtpForwarderRm(suuid string) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	delete(config.rtspServers, suuid)
}

func (config *ConfigST) RtpForwarderCloseAndRm(suuid string) {
	if o := config.RtpForwarderGet(suuid); o != nil {
		o.Stop()
		config.RtpForwarderRm(suuid)
	}
}

// ------->> webrtc

func (config *ConfigST) WebrtcSerAdd(suuid string, server *webrtc_server.WebRTCServer) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	config.webrtcServers[suuid] = server
}

func (config *ConfigST) WebrtcSerGet(suuid string) *webrtc_server.WebRTCServer {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	return config.webrtcServers[suuid]
}

func (config *ConfigST) WebrtcSerRm(suuid string) {
	config.mutex.Lock()
	defer config.mutex.Unlock()
	delete(config.webrtcServers, suuid)
}

func (config *ConfigST) WebrtcCloseAndRm(suuid string) {
	if o := config.WebrtcSerGet(suuid); o != nil {
		_ = o.Close()
		config.WebrtcSerRm(suuid)
	}
}
