#!/bin/zsh

echo "编译 macOS"

go build && mv gortsp dist

echo "编译 Windows"

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build && mv gortsp.exe dist