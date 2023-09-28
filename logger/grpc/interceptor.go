package grpc

import (
	"context"

	"github.com/easyparkgroup/go-svc-kit/pkg/idgen"
	"github.com/easyparkgroup/go-svc-kit/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type requestIDKey struct{}

var reqIDKey = requestIDKey{}

func LoggingServerInterceptor(baseLog logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		reqCtx := withNewReqID(ctx)

		l := baseLog.WithCtx(reqCtx)
		l.Infow("server call start", "method", info.FullMethod)
		resp, err := handler(reqCtx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				l.Warnw("server call end", "method", info.FullMethod, "status", "NotFound")
			} else {
				l.Errorw("server call error", "method", info.FullMethod)
			}
		} else {
			l.Infow("server call end", "method", info.FullMethod)
		}

		return resp, err
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func LoggingServerStreamingInterceptor(baseLog logger.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		reqCtx := withNewReqID(ss.Context())

		l := baseLog.WithCtx(reqCtx)
		l.Infow("server streaming call start", "method", info.FullMethod)
		ws := &wrappedStream{
			ServerStream: ss,
			ctx:          reqCtx,
		}
		err := handler(srv, ws)
		if err != nil {
			l.Errorw("server streaming call error", "method", info.FullMethod)
			return err
		}

		l.Infow("server streaming call completed", "method", info.FullMethod)
		return nil
	}
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(reqIDKey); v != nil {
		return v.(string)
	}

	return ""
}

func withNewReqID(ctx context.Context) context.Context {
	reqID := idgen.NewULID()
	return context.WithValue(ctx, reqIDKey, reqID)
}
