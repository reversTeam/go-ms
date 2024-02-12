package core

import (
	"context"
	"fmt"
	"strings"

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

func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var chainHandler grpc.StreamHandler

		// Définition de la fonction chainHandler qui encapsule le flux initial
		chainHandler = func(srv interface{}, stream grpc.ServerStream) error {
			return handler(srv, stream)
		}

		// Itérer sur les intercepteurs en sens inverse
		for i := len(interceptors) - 1; i >= 0; i-- {
			currentInterceptor := interceptors[i]
			nextHandler := chainHandler // Capture de la référence actuelle à la chainHandler pour chaque itération

			// Définition de la nouvelle chainHandler qui encapsule l'intercepteur actuel et la chainHandler précédente
			chainHandler = func(srv interface{}, stream grpc.ServerStream) error {
				return currentInterceptor(srv, stream, info, nextHandler)
			}
		}

		// Appel du flux de gestion final avec la chainHandler finale
		return chainHandler(srv, stream)
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

func loggingStreamMiddleware(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	ctx, span := Trace(ctx, "grpc", info.FullMethod)
	defer span.End()

	spanContext := span.SpanContext()
	spanID := spanContext.SpanID().String()
	spanIDHeader := metadata.Pairs("request-id", spanID)

	stream.SendHeader(spanIDHeader)

	return handler(srv, &wrappedServerStream{ServerStream: stream, ctx: ctx})
}

// wrappedServerStream is a wrapper around grpc.ServerStream to override the Context method.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// func applyDynamicMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, middlewareFnMap map[string]GoMsMiddlewareFunc) (interface{}, error) {
// 	middlewaresCtx, middlewaresSpan := Trace(ctx, "middlewares", "apply-dynamic-middlewares")

// 	cachedMiddlewares := getCachedMiddlewareByServiceEndpoint(info)
// 	var err error
// 	var res interface{} = req

// 	for i := 0; i < len(cachedMiddlewares); i++ {
// 		m := cachedMiddlewares[i]

// 		mFn, ok := middlewareFnMap[m]
// 		if !ok {
// 			return nil, fmt.Errorf("middleware %s does not exist", m)
// 		}

// 		nextHandler := func(innerCtx context.Context, innerReq interface{}) (interface{}, error) {
// 			return res, err
// 		}

// 		_, span := Trace(middlewaresCtx, "middleware", m)
// 		res, err = mFn(ctx, res, info, nextHandler)
// 		span.End()

// 		if err != nil {
// 			break
// 		}
// 	}
// 	middlewaresSpan.End()

// 	if err == nil {
// 		res, err = handler(ctx, res)
// 	}

// 	return res, err
// }

func applySelectedMiddleware(ctx context.Context, res interface{}, selectedMiddlewares []string, middlewares map[string]Middleware) (context.Context, interface{}, error) {
	var err error

	for _, middleware := range selectedMiddlewares {
		_, middlewareSpan := Trace(ctx, "middleware", middleware)
		ctx, res, err = middlewares[middleware].Apply(ctx, res)
		middlewareSpan.End()
		if err != nil {
			return nil, nil, err
		}
	}

	return ctx, res, nil
}

func applyMiddleware(ctx context.Context, srv interface{}, req interface{}, info interface{}, handler interface{}, middlewares map[string]Middleware, mdConf map[string]map[string][]string) (interface{}, error) {
	var err error = nil
	var res interface{} = req
	middlewaresCtx, middlewaresSpan := Trace(ctx, "middlewares", "apply-middlewares")
	defer middlewaresSpan.End()

	switch h := handler.(type) {
	case grpc.UnaryHandler:
		methodParts := strings.Split(info.(*grpc.UnaryServerInfo).FullMethod, "/")
		service := strings.Split(methodParts[1], ".")[3]
		methodName := methodParts[len(methodParts)-1]

		ctx, res, err = applySelectedMiddleware(middlewaresCtx, res, mdConf[service][methodName], middlewares)
		if err != nil {
			return nil, err
		}
		return h(ctx, res)
	case grpc.StreamHandler:
		methodParts := strings.Split(info.(*grpc.StreamServerInfo).FullMethod, "/")
		service := strings.Split(methodParts[1], ".")[3]
		methodName := methodParts[len(methodParts)-1]

		_, _, err := applySelectedMiddleware(middlewaresCtx, res, mdConf[service][methodName], middlewares)
		if err != nil {
			return nil, err
		}

		ss, ok := req.(grpc.ServerStream)
		if !ok {
			return nil, fmt.Errorf("expected grpc.ServerStream, got %T", req)
		}
		err = h(srv, ss)
		return nil, err
	default:
	}
	return nil, fmt.Errorf("Request type is not implemented")
}
