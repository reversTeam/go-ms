package core

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func Trace(ctx context.Context, name string, action string) (context.Context, trace.Span) {
	propagator := otel.GetTextMapPropagator()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		propagator.Inject(ctx, MetadataReaderWriter{md})
		ctx = propagator.Extract(ctx, MetadataReaderWriter{md})
	}

	tracer := otel.Tracer(name)
	ctx, span := tracer.Start(ctx, action, trace.WithSpanKind(trace.SpanKindServer))

	return ctx, span
}
