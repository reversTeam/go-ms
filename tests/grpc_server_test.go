package tests

import (
	"context"
	"net"
	"testing"

	core "github.com/reversTeam/go-ms/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	go func() {
		if err := s.Serve(lis); err != nil {
			panic("Server exited with error")
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestNewGoMsGrpcServer(t *testing.T) {
	ctx := &core.Context{}
	config := &core.ServerConfig{Host: "localhost", Port: 10001}
	opts := []grpc.DialOption{grpc.WithInsecure()}
	middlewares := make(map[string]core.Middleware)

	server := core.NewGoMsGrpcServer(ctx, config, opts, middlewares)

	if server.Host != "localhost" || server.Port != 10001 {
		t.Errorf("NewGoMsGrpcServer() = %v, want %v", server.Host, "localhost")
	}
}

func TestGoMsGrpcServer_GracefulStop(t *testing.T) {
	ctx := &core.Context{}
	config := &core.ServerConfig{Host: "localhost", Port: 10003}
	opts := []grpc.DialOption{grpc.WithContextDialer(bufDialer)}
	middlewares := make(map[string]core.Middleware)

	server := core.NewGoMsGrpcServer(ctx, config, opts, middlewares)
	server.Listen()
	server.Start()

	err := server.GracefulStop()
	if err != nil {
		t.Fatalf("GracefulStop() failed with %v", err)
	}

	if server.State == core.Ready {
		t.Errorf("server.State = %v, want not %v", server.State, core.Ready)
	}
}
