package core

import (
	"context"

	"google.golang.org/grpc"
)

type BaseMiddleware struct {
	Middleware
}

func (m *BaseMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, newReq, err := m.Apply(ctx, req)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, newReq)
	}
}

func (m *BaseMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, _, err := m.Apply(ss.Context(), nil)
		if err != nil {
			return err
		}
		wrappedStream := wrapServerStream(ss, newCtx)
		return handler(srv, wrappedStream)
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func wrapServerStream(ss grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &wrappedStream{ServerStream: ss, ctx: ctx}
}
