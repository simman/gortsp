package ffgo

//#cgo pkg-config: libavformat libavcodec libavutil
//#include <stdio.h>
//#include "ffgo.h"
import "C"

//var client C.struct_ff_rtsp_client

func X265_to_H264() int {
	//ret := C.x265_to_h264()
	//return int(ret)
	client := C.struct_ff_rtsp_client{}

	_ = C.prepare_rtsp_client(client, C.CString("rtsp://192.168.5.222:5544/live/test110/test110"))

	return 0
}
