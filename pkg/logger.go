package pkg

import (
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"time"
)

var LogToFile = false

var Logger *slog.Logger

func init() {
	var w *os.File
	if LogToFile {
		w, _ = os.OpenFile("logs.log", os.O_CREATE|os.O_RDWR, 0666)
	} else {
		w = os.Stderr
	}

	Logger = slog.New(tint.NewHandler(w, &tint.Options{
		AddSource:  true,
		Level:      slog.LevelDebug,
		TimeFormat: time.RFC3339,
		NoColor:    LogToFile,
	}))
}
