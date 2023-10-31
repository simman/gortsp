package utils

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func Bytes2Bits(data []byte) []int {
	dst := make([]int, 0)
	for _, v := range data {
		for i := 0; i < 8; i++ {
			move := uint(7 - i)
			dst = append(dst, int((v>>move)&1))
		}
	}
	//fmt.Println(len(dst))
	return dst
}

func Bytes2HexString(data []byte) string {
	hexStr := hex.EncodeToString(data)

	var tmp string
	for i, s := range strings.Split(hexStr, "") {
		if i%2 == 0 {
			tmp += fmt.Sprintf("%s", s)
		} else {
			tmp += fmt.Sprintf("%s ", s)
		}
	}
	return tmp
}
