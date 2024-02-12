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
	middlewaresCtx, middlewaresSpan := Trace(ctx, "middlewares", "apply-dynamic-middlewares")

	middlewares := getCachedMiddlewareByServiceEndpoint(info)
	var err error
	var res interface{} = req

	for i := 0; i < len(middlewares); i++ {
		m := middlewares[i]

		mFn, ok := middlewareFnMap[m]
		if !ok {
			return nil, fmt.Errorf("middleware %s does not exist", m)
		}

		nextHandler := func(innerCtx context.Context, innerReq interface{}) (interface{}, error) {
			return res, err
		}

		_, span := Trace(middlewaresCtx, "middleware", m)
		res, err = mFn(ctx, res, info, nextHandler)
		span.End()

		if err != nil {
			break
		}
	}
	middlewaresSpan.End()

	if err == nil {
		res, err = handler(ctx, res)
	}

	return res, err
}
