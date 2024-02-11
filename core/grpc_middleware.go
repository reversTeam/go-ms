package core

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func chainInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		for i := len(interceptors) - 1; i >= 0; i-- {
			currentInterceptor := interceptors[i]
			nextHandler := handler
			handler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, info, nextHandler)
			}
		}
		return handler(ctx, req)
	}
}

func loggingMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := Trace(ctx, "grpc", info.FullMethod)
	defer span.End()

	spanContext := span.SpanContext()
	spanID := spanContext.SpanID().String()
	spanIDHeader := metadata.Pairs("request-id", spanID)

	grpc.SendHeader(ctx, spanIDHeader)

	return handler(ctx, req)
}

func applyDynamicMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, middlewareFnMap map[string]GoMsMiddlewareFunc) (interface{}, error) {
	ctx, parentSpan := Trace(ctx, "middlewares", "apply-dynamic-middlewares")
	defer parentSpan.End()

	middlewares := getCachedMiddlewareByServiceEndpoint(info)
	currentHandler := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		m := middlewares[i]

		mFn, ok := middlewareFnMap[m]
		if !ok {
			return nil, fmt.Errorf("middleware %s does not exist", m)
		}

		nextHandler := currentHandler
		currentHandler = func(mdCtx context.Context, currentReq interface{}) (interface{}, error) {
			_, middlewareSpan := Trace(mdCtx, "middleware", m)
			defer middlewareSpan.End()
			return mFn(mdCtx, currentReq, info, nextHandler)
		}
	}

	return currentHandler(ctx, req)
}
