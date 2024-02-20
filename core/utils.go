package core

import (
	"context"
	"crypto/rand"
	"math/big"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func Trace(ctxParent context.Context, name string, action string) (context.Context, trace.Span) {
	propagator := otel.GetTextMapPropagator()
	ctx := ctxParent

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		propagator.Inject(ctxParent, MetadataReaderWriter{md})
	}
	ctx = propagator.Extract(ctxParent, MetadataReaderWriter{md})

	tracer := otel.Tracer(name)
	ctx, span := tracer.Start(ctx, action, trace.WithSpanKind(trace.SpanKindServer))

	return ctx, span
}

func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[randInt.Int64()]
	}
	return string(b), nil
}
