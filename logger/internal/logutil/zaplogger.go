package logutil

import (
	"strings"

	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logfmtEncoding = "logfmt"

func CreateLogger(level, encoding string) *zap.Logger {
	zc := zap.NewProductionConfig()
	zc.Encoding = "json"
	if strings.EqualFold(encoding, logfmtEncoding) {
		zc.Encoding = logfmtEncoding
	}
	zc.DisableCaller = true
	zc.EncoderConfig.TimeKey = "@timestamp"
	zc.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	zc.OutputPaths = []string{"stdout"}
	zc.Level = ParseLevel(level)
	l, _ := zc.Build()
	return l
}

func ParseLevel(level string) zap.AtomicLevel {
	l, err := zapcore.ParseLevel(level)
	if err != nil {
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	return zap.NewAtomicLevelAt(l)
}

func init() {
	zap.RegisterEncoder("logfmt", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		enc := zaplogfmt.NewEncoder(cfg)
		return enc, nil
	})
}
