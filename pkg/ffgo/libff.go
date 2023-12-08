package ffgo

//#cgo pkg-config: libavformat libavcodec libavutil
//#include <stdio.h>
//#include "ffgo.h"
import "C"
import "log"

var client C.ff_rtsp_client

func X265_to_H264() {
	//ret := C.x265_to_h264()
	//return int(ret)

	C.ff_version()

	client = C.ff_rtsp_client{}

	ret := C.prepare_rtsp_client(&client, C.CString("rtsp://192.168.5.222:8554/live/test110/test110"))
	log.Println("ret: %d", ret)
	log.Println("url: %s", C.GoString(client.url))

	log.Println("========================== 退出程序 =========================")
	select {}
	//return 0
}
