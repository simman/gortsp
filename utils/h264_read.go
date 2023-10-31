package utils

import (
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
	"os"
)

func H264Reader(path string) (*h264reader.H264Reader, error) {
	file, h264Err := os.Open(path)
	if h264Err != nil {
		return nil, h264Err
	}

	h264, h264Err := h264reader.NewReader(file)
	if h264Err != nil {
		return nil, h264Err
	}

	return h264, nil
}
