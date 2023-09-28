package grpc

import (
	"github.com/easyparkgroup/go-svc-kit/pkg/logger"
	"github.com/tochemey/gopack/logger/internal/logutil"
	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

const (
	// pkg grpc transport logger logs at verbosity level 2
	verbosityLevel = 2
)

type GrpcLogOption func(*grpcLogOpts)

func WithLogLevel(level string) GrpcLogOption {
	return func(opt *grpcLogOpts) {
		opt.level = level
	}
}

func WithLogEncoding(encoding string) GrpcLogOption {
	return func(opt *grpcLogOpts) {
		opt.encoding = encoding
	}
}

type grpcLogger struct {
	l *zap.SugaredLogger
}

type grpcLogOpts struct {
	encoding string
	level    string
}

func New(opts ...GrpcLogOption) grpclog.LoggerV2 {
	lo := &grpcLogOpts{}

	for _, opt := range opts {
		opt(lo)
	}

	ll := logutil.CreateLogger(lo.level, lo.encoding)
	return &grpcLogger{
		l: ll.Sugar(),
	}
}

// NewGrpcLogger creates new instance of grpc LoggerV2
func NewGrpcLogger(baseLog logger.Logger, level string) grpclog.LoggerV2 {
	logcore, ok := baseLog.CoreLog().(*zap.Logger)
	if !ok || logcore == nil {
		panic("please initialize main logger")
	}
	lvl := logutil.ParseLevel(level).Level()
	l := logcore.WithOptions(zap.IncreaseLevel(lvl)).Sugar()
	return &grpcLogger{
		l: l,
	}
}

func (g *grpcLogger) Info(args ...interface{}) {
	g.l.Info(args...)
}

func (g *grpcLogger) Infoln(args ...interface{}) {
	g.l.Info(args...)
}

func (g *grpcLogger) Infof(format string, args ...interface{}) {
	g.l.Infof(format, args...)
}

func (g *grpcLogger) Warning(args ...interface{}) {
	g.l.Warn(args...)
}

func (g *grpcLogger) Warningln(args ...interface{}) {
	g.l.Warn(args...)
}

func (g *grpcLogger) Warningf(format string, args ...interface{}) {
	g.l.Warnf(format, args...)
}

func (g *grpcLogger) Error(args ...interface{}) {
	g.l.Error(args...)
}

func (g *grpcLogger) Errorln(args ...interface{}) {
	g.l.Error(args...)
}

func (g *grpcLogger) Errorf(format string, args ...interface{}) {
	g.l.Errorf(format, args...)
}

func (g *grpcLogger) Fatal(args ...interface{}) {
	g.l.Fatal(args...)
}

func (g *grpcLogger) Fatalln(args ...interface{}) {
	g.l.Fatal(args...)
}

func (g *grpcLogger) Fatalf(format string, args ...interface{}) {
	g.l.Fatalf(format, args...)
}

func (g *grpcLogger) V(l int) bool {
	return l <= verbosityLevel
}
