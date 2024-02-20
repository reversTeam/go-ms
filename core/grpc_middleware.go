package core

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
)

type GoMsResponseWrapper struct {
	Code     int
	Response interface{}
}

// TODO: Useless code
// func chainInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
// 	return func(parentCtx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 		currentCtx := parentCtx
// 		lastCtx := parentCtx
// 		var err error
// 		var res interface{}

// 		for i := range interceptors {
// 			index := len(interceptors) - 1 - i
// 			log.Printf("---- LOOP : %d / %d\n", i, index)

// 			currentInterceptor := interceptors[index]
// 			currentHandler := handler

// 			nextHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
// 				return currentHandler(ctx, req)
// 			}

// 			currentHandler = func(innerCtx context.Context, innerReq interface{}) (interface{}, error) {
// 				return currentInterceptor(innerCtx, innerReq, info, nextHandler)
// 			}

// 			res, err = currentHandler(currentCtx, err)
// 			if err != nil {
// 				break
// 			}
// 		}

// 		if err == nil {
// 			res, err = handler(lastCtx, req)
// 		}

// 		if err != nil {
// 			if httpErr, ok := err.(*HttpError); ok {
// 				md := metadata.Pairs("http-status-code", fmt.Sprintf("%d", httpErr.Code))
// 				grpc.SendHeader(lastCtx, md)
// 			}
// 		}

// 		return res, err
// 	}
// }

// func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
// 	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
// 		var chainHandler grpc.StreamHandler

// 		chainHandler = func(srv interface{}, stream grpc.ServerStream) error {
// 			return handler(srv, stream)
// 		}

// 		for i := len(interceptors) - 1; i >= 0; i-- {
// 			currentInterceptor := interceptors[i]
// 			nextHandler := chainHandler

// 			chainHandler = func(srv interface{}, stream grpc.ServerStream) error {
// 				return currentInterceptor(srv, stream, info, nextHandler)
// 			}
// 		}

// 		return chainHandler(srv, stream)
// 	}
// }

// func loggingMiddleware(parentCtx context.Context, params interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 	middlewareCtx, span := Trace(parentCtx, "grpc", info.FullMethod)
// 	defer span.End()

// 	spanContext := span.SpanContext()
// 	spanID := spanContext.SpanID().String()
// 	spanIDHeader := metadata.Pairs("request-id", spanID)

// 	if err := grpc.SendHeader(middlewareCtx, spanIDHeader); err != nil {
// 		return nil, err
// 	}

// 	return handler(middlewareCtx, params)
// }

// func loggingStreamMiddleware(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
// 	ctx := stream.Context()
// 	ctx, span := Trace(ctx, "grpc", info.FullMethod)
// 	defer span.End()

// 	spanContext := span.SpanContext()
// 	spanID := spanContext.SpanID().String()
// 	spanIDHeader := metadata.Pairs("request-id", spanID)

// 	err := stream.SendHeader(spanIDHeader)
// 	if err != nil {
// 		return err
// 	}

// 	return handler(srv, &wrappedServerStream{ServerStream: stream, Ctx: ctx})
// }

func applySelectedMiddleware(ctx context.Context, res interface{}, selectedMiddlewares []string, middlewares map[string]Middleware) (context.Context, interface{}, error) {
	var err error

	for _, middleware := range selectedMiddlewares {
		mctx, middlewareSpan := Trace(ctx, "middleware", middleware)
		_, res, err = middlewares[middleware].Apply(mctx, res)
		middlewareSpan.End()
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, res, nil
}

func applyMiddleware(ctx context.Context, srv interface{}, req interface{}, info interface{}, handler interface{}, middlewares map[string]Middleware, mdConf map[string]map[string][]string) (interface{}, error) {
	var err error
	var res interface{} = req
	parentCtx := ctx

	switch h := handler.(type) {
	case grpc.UnaryHandler:
		methodParts := strings.Split(info.(*grpc.UnaryServerInfo).FullMethod, "/")
		service := strings.Split(methodParts[1], ".")[3]
		methodName := methodParts[len(methodParts)-1]

		_, res, err = applySelectedMiddleware(parentCtx, res, mdConf[service][methodName], middlewares)
		if err != nil {
			return nil, err
		}

		return res, err
	case grpc.StreamHandler:
		methodParts := strings.Split(info.(*grpc.StreamServerInfo).FullMethod, "/")
		service := strings.Split(methodParts[1], ".")[3]
		methodName := methodParts[len(methodParts)-1]

		_, _, err := applySelectedMiddleware(parentCtx, res, mdConf[service][methodName], middlewares)
		if err != nil {
			return nil, err
		}

		ss, ok := req.(grpc.ServerStream)
		if !ok {
			return nil, fmt.Errorf("expected grpc.ServerStream, got %T", req)
		}
		// middlewaresSpan.End()
		handleCtx, s := Trace(ctx, "gprc", "handler")
		defer s.End()
		wrappedSS := wrapServerStream(ss, handleCtx)

		return nil, h(srv, wrappedSS)
	default:
		// middlewaresSpan.End()
	}
	return nil, fmt.Errorf("Request type is not implemented")
}
