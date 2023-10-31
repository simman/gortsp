package main

import (
	"bytes"
	"fmt"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
	"gortsp/pkg"
	"gortsp/utils"
	"log/slog"
	"os"
	"strings"
)

func parsePayload(payload []byte, index int) {
	str := utils.Bytes2HexString(payload[:5])
	pkg.Logger.Info("[ParsePayload]", "hex", str, "index", index)
	//
	have701 := strings.Contains(str, "00 00 00 01")
	have501 := strings.Contains(str, "00 00 01")

	pkg.Logger.Info("[ParsePayload]", "payload[0]", utils.Bytes2Bits([]byte(string(payload[0]))))
	packType := payload[0] & 0x1f
	pkg.Logger.Info("[ParsePayload]", "packType", packType)

	// 解析
	fuHeader := payload[1:2]
	fuHeaderBits := utils.Bytes2Bits(fuHeader)
	pkg.Logger.Info("[ParsePayload]", "fuHeaderBits", fuHeaderBits)

	var isStart = fuHeaderBits[0] == 1
	var isCenter = fuHeaderBits[0] == 0 && fuHeaderBits[1] == 0
	var isEnd = fuHeaderBits[1] == 1

	slog.Info("FuHeader", "起始包", isStart, "中间包", isCenter, "结束包", isEnd, "have501", have501, "have701", have701)
}

func mainTest() {
	h264PacketCodes := codecs.H264Packet{}
	//h264_data_file, _ := os.OpenFile("h264_data_file.h264", os.O_CREATE|os.O_RDWR, 0666)
	for i := 0; i < 10; i++ {
		//rtpData, err := os.ReadFile(fmt.Sprintf("test_data/s20_rtp_data/s20_%d.bin", i))
		rtpData, err := os.ReadFile(fmt.Sprintf("test_data/mediamtx_%d.rtp", i))
		if err != nil {
			panic(err)
		}

		// 解析 RTP 包
		rtpPacket := &rtp.Packet{}
		if err := rtpPacket.Unmarshal(rtpData); err != nil {
			panic(err)
		}

		// 调试fua header
		if true {
			parsePayload(rtpPacket.Payload, i)
		}

		// 解析h264
		h264Bytes, err := h264PacketCodes.Unmarshal(rtpPacket.Payload)
		if err != nil {
			panic(err)
		}

		if len(h264Bytes) > 0 {
			//h264_data_file.Write(h264_b)
		}
	}
}

func testH264File() {

	//reader1, _ := utils.H264Reader("/Users/longzy/Documents/media/gomedia/test_data/h264_file_01/video_full.h264")
	reader2, _ := utils.H264Reader("/Users/longzy/Documents/media/gomedia/test_data/h264_file_01/video_0.h264")

	//nalInfo1, _ := reader1.NextNAL()
	nalInfo2, _ := reader2.NextNAL()

	fileBytes, _ := os.ReadFile("/Users/longzy/Documents/media/gomedia/test_data/h264_file_01/video_0.h264")
	nalData := nalInfo2.Data

	fileBytes1 := fileBytes[4:]

	r := bytes.NewReader(fileBytes)
	xxxxxx, _ := h264reader.NewReader(r)

	ooo, _ := xxxxxx.NextNAL()
	asdfasdf := ooo.Data

	fmt.Println(nalData)
	fmt.Println(fileBytes1)
	fmt.Println(asdfasdf)
	//nalInfo3, err := reader2.NextNAL()

	//var hasNext = true
	//var count = 0
	//for hasNext {
	//	nalInfo3, nalErr := nal2.NextNAL()
	//	if nalErr != io.EOF {
	//		count++
	//		fmt.Println(nalInfo3.UnitType)
	//	} else {
	//		hasNext = false
	//	}
	//}
	//fmt.Println(count)

	//fmt.Println(nalInfo1.UnitType)
	fmt.Println(nalInfo2.UnitType)
	//fmt.Println(nalInfo3.UnitType)
	//fmt.Println(err)

}
