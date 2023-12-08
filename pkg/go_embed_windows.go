package pkg

import "embed"

//go:embed "gin-server/templates/*.tmpl"
var TemplateFs embed.FS

//go:embed "ffgo/ffmpeg/static/ffmpeg.exe"
var FFBin embed.FS
