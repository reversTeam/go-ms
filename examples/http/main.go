package main

import (
	"flag"
	"log"

	"github.com/reversTeam/go-ms/core"
	"github.com/reversTeam/go-ms/services/child"
	"github.com/reversTeam/go-ms/services/goms"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// Default flag values for GRPC server
	GRPC_DEFAULT_HOST = "127.0.0.1"
	GRPC_DEFAULT_PORT = 42001

	// Default flag values for http server
	HTTP_DEFAULT_HOST = "127.0.0.1"
	HTTP_DEFAULT_PORT = 8080

	// Default flag values for GRPC server
	EXPORTER_DEFAULT_HOST     = "127.0.0.1"
	EXPORTER_DEFAULT_PORT     = 4242
	EXPORTER_DEFAULT_PATH     = "/metrics"
	EXPORTER_DEFAULT_INTERVAL = 1
)

var (
	// flags for Grpc server
	grpcHost = flag.String("grpc-host", GRPC_DEFAULT_HOST, "Grpc listening host")
	grpcPort = flag.Int("grpc-port", GRPC_DEFAULT_PORT, "Grpc listening port")

	// flags for http server
	httpHost = flag.String("http-host", HTTP_DEFAULT_HOST, "http gateway host")
	httpPort = flag.Int("http-port", HTTP_DEFAULT_PORT, "http gateway port")

	// flags for Exporter server
	exporterHost     = flag.String("exporter-host", EXPORTER_DEFAULT_HOST, "Exporter listening host")
	exporterPort     = flag.Int("exporter-port", EXPORTER_DEFAULT_PORT, "Exporter listening port")
	exporterPath     = flag.String("exporter-path", EXPORTER_DEFAULT_PATH, "Exporter listening path")
	exporterInterval = flag.Int("exporter-interval", EXPORTER_DEFAULT_INTERVAL, "Exporter listening interval")
)

func main() {
	// Instantiate context in background
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Parse flags
	flag.Parse()

	// Create a gateway configuration
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	// setup exporter
	exporterServer := core.NewExporter(ctx, *exporterHost, *exporterPort, *exporterPath, *exporterInterval)

	// setup servers
	grpcServer := core.NewGoMsGrpcServer(ctx, *grpcHost, *grpcPort, opts)
	httpServer := core.NewGoMsHttpServer(ctx, *httpHost, *httpPort, grpcServer)

	// setup services
	gomsService := goms.NewService("goms", core.ServiceConfig{})
	childService := child.NewService("child", core.ServiceConfig{})

	_ = childService

	// Register service to the http and grpc server
	httpServer.AddService(gomsService)
	httpServer.AddService(childService)

	// Graceful stop servers
	core.AddServerGracefulStop(httpServer)
	core.AddServerGracefulStop(exporterServer)
	// Catch ctrl + c
	done := core.CatchStopSignals()

	// Start exporter server
	exporterServer.Start()
	// Start Http Server
	err := httpServer.Start()
	if err != nil {
		log.Fatal("An error occured, the http server can be running", err)
	}

	<-done
}
