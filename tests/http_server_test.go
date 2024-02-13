package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/reversTeam/go-ms/core"
	"google.golang.org/grpc"
)

var httpServerConfig = core.HttpServerConfig{
	ServerConfig: core.ServerConfig{
		Host: "localhost",
		Port: 8080,
	},
	ReadTimeout:  10,
	WriteTimeout: 10,
}

func TestNewGoMsHttpServer(t *testing.T) {
	ctx := core.NewContext("test-service", core.JaegerConfig{Mode: "mock"})
	grpcServer := core.NewGoMsGrpcServer(ctx, &core.ServerConfig{}, []grpc.DialOption{}, map[string]core.Middleware{})
	httpServer := core.NewGoMsHttpServer(ctx, &httpServerConfig, grpcServer)
	if httpServer.Host != "localhost" || httpServer.Port != 8080 {
		t.Errorf("Expected host %s and port %d, got host %s and port %d", "localhost", 8080, httpServer.Host, httpServer.Port)
	}
}

func TestGoMsHttpServer_SetExporter(t *testing.T) {
	ctx := core.NewContext("test-service", core.JaegerConfig{Mode: "mock"})
	grpcServer := core.NewGoMsGrpcServer(ctx, &core.ServerConfig{}, []grpc.DialOption{}, map[string]core.Middleware{})
	httpServer := core.NewGoMsHttpServer(ctx, &httpServerConfig, grpcServer)
	exporter := &core.Exporter{}
	httpServer.SetExporter(exporter)
	if httpServer.Exporter != exporter {
		t.Errorf("Expected exporter to be set")
	}
}

func TestGoMsHttpServer_Handle(t *testing.T) {
	ctx := core.NewContext("test-service", core.JaegerConfig{Mode: "mock"})
	grpcServer := core.NewGoMsGrpcServer(ctx, &core.ServerConfig{}, []grpc.DialOption{}, map[string]core.Middleware{})
	httpServer := core.NewGoMsHttpServer(ctx, &httpServerConfig, grpcServer)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	httpServer.Handle("/", httpServer.Mux)

	handler.ServeHTTP(rr, req)

	status := rr.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGoMsHttpServer_StartAndGracefulStop(t *testing.T) {
	ctx := core.NewContext("test-service", core.JaegerConfig{Mode: "mock"})
	grpcServer := core.NewGoMsGrpcServer(ctx, &core.ServerConfig{}, []grpc.DialOption{}, map[string]core.Middleware{})
	httpServer := core.NewGoMsHttpServer(ctx, &httpServerConfig, grpcServer)

	go func() {
		if err := httpServer.Start(); err != nil {
			t.Errorf("Failed to start HTTP server: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	if err := httpServer.GracefulStop(); err != nil {
		t.Errorf("Failed to gracefully stop HTTP server: %v", err)
	}
}
