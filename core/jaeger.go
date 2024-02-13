package core

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Jeager struct {
	Config        JaegerConfig
	Exporter      sdktrace.SpanExporter
	TraceProvider trace.TracerProvider
	GracefullFunc func()
}

func NewJeager(ctx context.Context, name string, config JaegerConfig) *Jeager {
	endpoint := fmt.Sprintf("%s:%d", config.Host, config.Port)
	var exporter sdktrace.SpanExporter
	var err error

	// TODO: check grpc implementation
	if config.Mode == "grpc" {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}
		if !config.Unsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(ctx, opts...)
	} else if config.Mode == "http" {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
		}
		if config.Unsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
	} else if config.Mode == "mock" {
		return &Jeager{}
	} else {
		log.Fatal("Jaeger mode not implemented")
	}

	if err != nil {
		log.Fatalf("Error creating OTLP exporter: %v", err)
	}

	serviceNameAttr := attribute.String("service.name", name)
	attrs := []attribute.KeyValue{serviceNameAttr}
	resource := resource.NewWithAttributes("service.name", attrs...)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &Jeager{
		Config:        config,
		Exporter:      exporter,
		GracefullFunc: createShutdownFunc(exporter, tp),
		TraceProvider: tp,
	}
}

func createShutdownFunc(exporter sdktrace.SpanExporter, tp *sdktrace.TracerProvider) func() {
	return func() {
		if err := exporter.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down OTLP exporter: %v", err)
		}
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down OpenTelemetry Tracer Provider: %v", err)
		}
	}
}
