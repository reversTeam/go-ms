package tests

// import (
// 	"context"
// 	"testing"

// 	"github.com/reversTeam/go-ms/core"
// 	"github.com/stretchr/testify/assert"
// 	"google.golang.org/grpc"
// )

// type mockMiddleware struct {
// 	core.BaseMiddleware
// }

// func (m *mockMiddleware) Apply(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
// 	// Mock apply function logic here
// 	return ctx, req, nil
// }

// func TestBaseMiddlewareUnary(t *testing.T) {
// 	middleware := &mockMiddleware{}
// 	interceptor := middleware.Unary()

// 	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
// 		return req, nil
// 	}

// 	resp, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, handler)

// 	assert.Nil(t, err)
// 	assert.Equal(t, "request", resp)
// }

// func TestBaseMiddlewareStream(t *testing.T) {
// 	middleware := &mockMiddleware{}
// 	interceptor := middleware.Stream()

// 	handler := func(srv interface{}, stream grpc.ServerStream) error {
// 		return nil
// 	}

// 	mockStream := &mockServerStream{}

// 	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{}, handler)

// 	assert.Nil(t, err)
// }

// type mockServerStream struct {
// 	grpc.ServerStream
// }

// func (m *mockServerStream) Context() context.Context {
// 	return context.Background()
// }
