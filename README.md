# RTSP OVER UDP TO WebRTC

本项目无`cgo` `ffmpeg` 等依赖, 完全使用golang实现.

使用的第三方依赖包

- gin: 提供http接口服务
- gomedia: 提供rtsp、rtp解析
- pion: 提供webrtc实现
- adapterjs: 提供webrtc兼容

## 一、演示

1. 启动工程
2. 浏览器访问: http://127.0.0.1:17890

## 二、构建

### 1. 使用脚本构建

```shell
./build.sh
```

产物会生成到 `./dist/` 目录中
