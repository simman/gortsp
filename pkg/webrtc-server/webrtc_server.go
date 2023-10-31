package webrtc_server

import (
	"github.com/pion/webrtc/v4"
	"gortsp/pkg"
	"gortsp/pkg/signal"
)

type WebRTCServer struct {
	peer                    *webrtc.PeerConnection
	sessionDesc             webrtc.SessionDescription
	TrackLocalStaticRTP     *webrtc.TrackLocalStaticRTP
	TrackLocalStaticSample  *webrtc.TrackLocalStaticSample
	IsStarted               bool
	onConnectionStateChange func(connectionState webrtc.ICEConnectionState)
}

func NewWebRTCServer() *WebRTCServer {
	// Create a new RTCPeerConnection
	peerConnection, _ := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	return &WebRTCServer{peer: peerConnection, sessionDesc: webrtc.SessionDescription{}}
}

func (rtc *WebRTCServer) WithRTP() *WebRTCServer {
	rtc.TrackLocalStaticRTP, _ = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	return rtc
}

func (rtc *WebRTCServer) WithSample() *WebRTCServer {
	rtc.TrackLocalStaticSample, _ = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	return rtc
}

func (rtc *WebRTCServer) WithSessionDescription(bsd string) *WebRTCServer {
	// Browser base64 Session Description
	signal.Decode(bsd, &rtc.sessionDesc)
	return rtc
}

func (rtc *WebRTCServer) OnConnectionStateChange(fn func(connectionState webrtc.ICEConnectionState)) {
	rtc.onConnectionStateChange = fn
}

func (rtc *WebRTCServer) LocalDescription() string {
	return signal.Encode(*rtc.peer.LocalDescription())
}

func (rtc *WebRTCServer) Close() error {
	return rtc.peer.Close()
}

func (rtc *WebRTCServer) Start() {
	var track webrtc.TrackLocal
	if rtc.TrackLocalStaticRTP != nil {
		track = rtc.TrackLocalStaticRTP
	}
	if rtc.TrackLocalStaticSample != nil {
		track = rtc.TrackLocalStaticSample
	}

	if track == nil {
		panic("track is nil")
	}

	rtpSender, err := rtc.peer.AddTrack(track)
	if err != nil {
		panic(err)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	rtc.peer.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		pkg.Logger.Info("Connection State has changed", "state", connectionState.String())

		if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := rtc.peer.Close(); closeErr != nil {
				panic(closeErr)
			}
		}

		if connectionState == webrtc.ICEConnectionStateConnected {
			rtc.IsStarted = true
		}

		if rtc.onConnectionStateChange != nil {
			rtc.onConnectionStateChange(connectionState)
		}
	})

	// Set the remote SessionDescription
	if err = rtc.peer.SetRemoteDescription(rtc.sessionDesc); err != nil {
		panic(err)
	}

	// Create answer
	answer, err := rtc.peer.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(rtc.peer)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = rtc.peer.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete
}
