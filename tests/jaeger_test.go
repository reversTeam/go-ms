package tests

import (
	"context"
	"testing"

	"github.com/reversTeam/go-ms/core"
	"github.com/stretchr/testify/assert"
)

func TestNewJeagerWithGrpcMode(t *testing.T) {
	ctx := context.Background()
	config := core.JaegerConfig{
		Host:     "localhost",
		Port:     4317,
		Mode:     "grpc",
		Unsecure: true,
	}

	jaeger := core.NewJeager(ctx, "testService", config)

	assert.NotNil(t, jaeger.Exporter)
	assert.NotNil(t, jaeger.TraceProvider)
	assert.NotNil(t, jaeger.GracefullFunc)
}

func TestNewJeagerWithHttpMode(t *testing.T) {
	ctx := context.Background()
	config := core.JaegerConfig{
		Host:     "localhost",
		Port:     4318,
		Mode:     "http",
		Unsecure: true,
	}

	jaeger := core.NewJeager(ctx, "testService", config)

	assert.NotNil(t, jaeger.Exporter)
	assert.NotNil(t, jaeger.TraceProvider)
	assert.NotNil(t, jaeger.GracefullFunc)
}

func TestNewJeagerGracefulShutdown(t *testing.T) {
	ctx := context.Background()
	config := core.JaegerConfig{
		Host:     "localhost",
		Port:     4317,
		Mode:     "grpc",
		Unsecure: true,
	}

	jaeger := core.NewJeager(ctx, "testService", config)

	// This test checks if graceful shutdown does not produce any error, but it's hard to assert without a real exporter
	assert.NotPanics(t, func() {
		jaeger.GracefullFunc()
	})
}
