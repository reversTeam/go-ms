package core

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func loggingMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, MetadataReaderWriter{md})
		tracer := otel.Tracer("grpc")
		_, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
	}

	return handler(ctx, req)
}
