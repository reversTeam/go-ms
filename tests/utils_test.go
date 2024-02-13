package tests

import (
	"context"
	"testing"

	"github.com/reversTeam/go-ms/core"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/metadata"
)

func TestTrace(t *testing.T) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource.NewWithAttributes(
			"test",
			attribute.String("service.name", "testService"),
		)),
	)
	otel.SetTracerProvider(tp)

	ctx := context.Background()
	md := metadata.MD{}
	ctx = metadata.NewIncomingContext(ctx, md)

	ctx, span := core.Trace(ctx, "testTracer", "testAction")

	assert.NotNil(t, ctx, "context should not be nil")
	assert.NotNil(t, span, "span should not be nil")
	assert.Equal(t, span.SpanContext().TraceID().IsValid(), true, "trace ID should be valid")
	_, ok := metadata.FromIncomingContext(ctx)
	assert.True(t, ok, "metadata should be injected into context")

	span.End()
}
