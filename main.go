package main

import (
	"fmt"
	"gortsp/pkg/gin-server"
)

var Version = "1.0.0-2310311419"
var GitSource = "http://192.168.100.4:9088/frontend/gortsp.git"

func main() {
	fmt.Println(fmt.Sprintf(`# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # 
# GoRTSP                                                            #
#                                                                   #
# Version: %s                                         #
# GitSource: %s          #
#                                                                   #
# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
`, Version, GitSource))
	gin_server.RunHttpServer()
}
