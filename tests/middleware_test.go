package tests

import (
	"context"
	"testing"

	"github.com/reversTeam/go-ms/core"
	"google.golang.org/grpc"
)

type mockMiddleware struct {
	core.Middleware
}

func (m *mockMiddleware) Apply(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
	return ctx, req, nil
}

func TestBaseMiddlewareUnary(t *testing.T) {
	middleware := core.BaseMiddleware{Middleware: &mockMiddleware{}}
	interceptor := middleware.Unary()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	}

	_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Errorf("Interceptor returned an error: %v", err)
	}
}

func TestBaseMiddlewareStream(t *testing.T) {
	middleware := core.BaseMiddleware{Middleware: &mockMiddleware{}}
	interceptor := middleware.Stream()

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	mockStream := &mockServerStream{}

	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{}, handler)
	if err != nil {
		t.Errorf("Interceptor returned an error: %v", err)
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func (m *mockServerStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) RecvMsg(msg interface{}) error {
	return nil
}
