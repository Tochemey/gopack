package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	zaplogfmt "github.com/sykesm/zap-logfmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWithReqID(t *testing.T) {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "time"
	config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	buf := &bytes.Buffer{}
	zl := zap.New(zapcore.NewCore(
		zaplogfmt.NewEncoder(config),
		zapcore.AddSync(buf),
		zapcore.InfoLevel,
	)).Sugar()

	tests := []struct {
		name           string
		reqIDExtractor RequestIDExtractor
	}{
		{
			name: "ReqIDandTracing",
			reqIDExtractor: func(ctx context.Context) string {
				return "test"
			},
		},
		{
			name: "ReqIDOnly",
			reqIDExtractor: func(ctx context.Context) string {
				return "test"
			},
		},
	}

	for _, tt := range tests {
		buf.Reset()

		t.Run(tt.name, func(t *testing.T) {
			ll := &loggerImpl{
				opts: &loggerOpts{
					reqIDExtractor: tt.reqIDExtractor,
				},
				log: zl,
			}

			ctx := context.Background()
			ll.WithCtx(ctx).Info("test")

			res := buf.String()

			if tt.reqIDExtractor != nil {
				if !strings.Contains(res, "req_id=") {
					t.Fatal("req_id field not found")
				}
			}
		})
	}
}
