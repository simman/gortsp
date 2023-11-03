#!/bin/zsh

echo "1. 生成桥接的 Napi C/C++ 代码"
gonacli generate --config ./goaddon.json

echo "2. 编译静态库"
gonacli build

echo "3. 安装 Nodejs 相关依赖"
gonacli install

echo "4. 编译 Nodejs Adddon"
gonacli make