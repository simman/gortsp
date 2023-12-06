package main

import "C"
import (
	"fmt"
	"gortsp/ffgo"
)

func RunConvert() {
	fmt.Println("=========>>>>>>>RunConvert")
	fmt.Println(ffgo.X265_to_H264())
}
