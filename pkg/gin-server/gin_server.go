package gin_server

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
	sloggin "github.com/samber/slog-gin"
	"gortsp/pkg"
	"gortsp/pkg/config"
	"gortsp/pkg/go-rtsp"
	go_rtsp_udp "gortsp/pkg/go-rtsp-udp"
	webrtc_server "gortsp/pkg/webrtc-server"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"
)

import (
	"github.com/gin-gonic/gin"
)

//go:embed templates/*.tmpl
var TemplateFs embed.FS

func RunHttpServer() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	templ := template.Must(template.New("").ParseFS(TemplateFs, "templates/*.tmpl"))
	router.SetHTMLTemplate(templ)

	router.Use(sloggin.New(pkg.Logger))
	router.Use(CORSMiddleware())
	router.Use(gin.Recovery())
	router.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	if _, err := os.Stat("./web"); !os.IsNotExist(err) {
		router.LoadHTMLGlob("web/templates/*")
	}
	router.GET("/", HTTPAPIServerIndexHandle)
	router.POST("/play/:uuid", SwitchLocalDescriptionAndPlayHandle)

	//router.StaticFS("/static", http.Dir("web/static"))

	pkg.Logger.Info("Start HTTP Server", "url", fmt.Sprintf("http://127.0.0.1:17890"))
	err := router.Run(":17890")

	if err != nil {
		panic(errors.Join(errors.New("Start HTTP Server error"), err))
	}
}

func HTTPAPIServerIndexHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", nil)
}

func SwitchLocalDescriptionAndPlayHandle(c *gin.Context) {
	sUUID := c.Param("uuid")
	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	jsonMap := make(map[string]interface{})
	_ = json.Unmarshal(bodyAsByteArray, &jsonMap)
	bsd := fmt.Sprintf("%s", jsonMap["bsd"])
	rtspAddress := fmt.Sprintf("%s", jsonMap["rtsp"])

	// 如果存在WebRTC, 则先关闭
	config.Config.WebrtcCloseAndRm(sUUID)

	// 创建并启动 WebRTC
	webRTCServer := webrtc_server.NewWebRTCServer().WithSample().WithSessionDescription(bsd)
	webRTCServer.OnConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected || connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed {
			// 如果存在RTSPServer, 则先关闭|删除
			config.Config.RtspCloseAndRm(sUUID)
		}
	})
	webRTCServer.Start()
	config.Config.WebrtcSerAdd(sUUID, webRTCServer)

	// RTP包回调
	var onRtpBufCallback = func(buf []byte) {
		webRTCServer := config.Config.WebrtcSerGet(sUUID)
		if webRTCServer != nil && webRTCServer.IsStarted && webRTCServer.TrackLocalStaticRTP != nil {
			if _, err := webRTCServer.TrackLocalStaticRTP.Write(buf); err != nil {
				panic(err)
			}
		}
	}

	// 帧回调
	var onSampleCallback = func(sample rtsp.RtspSample) {
		webRTCServer := config.Config.WebrtcSerGet(sUUID)
		if webRTCServer != nil && webRTCServer.IsStarted && webRTCServer.TrackLocalStaticSample != nil {
			r := bytes.NewReader(sample.Sample)
			if sample.Cid == rtsp.RTSP_CODEC_H264 {
				h264Reader, _ := h264reader.NewReader(r)
				nal, _ := h264Reader.NextNAL()
				mediaSample := media.Sample{
					Data: nal.Data,
					//Timestamp: time.Now(),
					Duration: time.Millisecond * 25,
					//PacketTimestamp:    nil,
					//PrevDroppedPackets: 0,
					//Metadata:           nil,
				}
				if err := webRTCServer.TrackLocalStaticSample.WriteSample(mediaSample); err != nil {
					pkg.Logger.Error("WriteSample", "err", err.Error())
				}
			} else if sample.Cid == rtsp.RTSP_CODEC_H265 {
				//h264 := codecs.H264Packet{}
				h265 := codecs.H265Packet{}

				if data, err := h265.Unmarshal(sample.Sample); err != nil {
					pkg.Logger.Error("h265 unmarshal failed", "err", err.Error())
				} else {
					pkg.Logger.Info("h265 unmarshal success...")
					mediaSample := media.Sample{
						Data: data,
						//Timestamp: time.Now(),
						Duration: time.Millisecond * 25,
						//PacketTimestamp:    nil,
						//PrevDroppedPackets: 0,
						//Metadata:           nil,
					}
					if err := webRTCServer.TrackLocalStaticSample.WriteSample(mediaSample); err != nil {
						pkg.Logger.Error("WriteSample", "err", err.Error())
					}
				}
			} else {
				pkg.Logger.Warn("Webrtc不支持当前编码", "CODEC", sample.Cid)
			}
		}
	}

	// 关闭后
	var onCloseCallback = func() {
		config.Config.WebrtcCloseAndRm(sUUID)
	}

	// 如果存在RTSPServer, 则先关闭
	config.Config.RtspCloseAndRm(sUUID)

	// 创建并启动 RTSPServer
	rtspServer, err := go_rtsp_udp.NewRtspUdpServer(rtspAddress, onRtpBufCallback, onSampleCallback, onCloseCallback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	config.Config.RtspAdd(sUUID, rtspServer)
	go rtspServer.Start()

	c.JSON(200, webRTCServer.LocalDescription())
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, x-access-token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
