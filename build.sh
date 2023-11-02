#!/bin/zsh

echo "编译 macOS"

go build -ldflags="-s -w" && mv gortsp dist

echo "编译 Windows"

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" && mv gortsp.exe dist