package core

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Definition of GoMsHttpServer struct
type GoMsHttpServer struct {
	Ctx      *Context
	Grpc     *GoMsGrpcServer
	Host     string
	Port     int
	Server   *http.Server
	State    GoMsServerState
	mux      *http.ServeMux
	Mux      *runtime.ServeMux
	services []GoMsServiceInterface
	Exporter *Exporter
}

// Init GoMsHttpServer
func NewGoMsHttpServer(ctx *Context, config *HttpServerConfig, grpcServer *GoMsGrpcServer) *GoMsHttpServer {
	uri := fmt.Sprintf("%s:%d", config.Host, config.Port)
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithMetadata(forwardHeaders),
	}
	mux := http.NewServeMux()

	return &GoMsHttpServer{
		Ctx:  ctx,
		Grpc: grpcServer,
		Host: config.Host,
		Port: config.Port,
		Server: &http.Server{
			Addr:           uri,
			Handler:        mux,
			ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		State:    Init,
		mux:      mux,
		Mux:      runtime.NewServeMux(muxOpts...),
		services: make([]GoMsServiceInterface, 0),
		Exporter: nil,
	}
}

// Set the exporter
func (o *GoMsHttpServer) SetExporter(exporter *Exporter) {
	o.Exporter = exporter
}

func (o *GoMsHttpServer) Handle(path string, mux *runtime.ServeMux) {
	o.mux.Handle(path, tracingMiddleware(mux))
}

// Register service on the http server
func (o *GoMsHttpServer) Register(service GoMsServiceInterface) error {
	return service.GetHandler().RegisterHttp(o, service)
}

// Add service in the local services list
func (o *GoMsHttpServer) AddService(service GoMsServiceInterface) {
	o.services = append(o.services, service)
}

// Register each service in the http server
func (o *GoMsHttpServer) startServices() error {
	for _, service := range o.services {
		log.Printf("[HTTP] Register service %s\n", service.GetName())
		err := o.Register(service)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start the http server, ready for handle connexion
func (o *GoMsHttpServer) Start() error {
	err := o.startServices()
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("%s:%d", o.Host, o.Port)
	log.Printf("[HTTP] Server listen on http://%s\n", uri)

	o.Handle("/", o.Mux)

	go func(httpServer *http.Server) {
		err := httpServer.ListenAndServe()
		// we can't catch this error, also we log it
		if err != nil {
			log.Println("[HTTP] Error listen: ", err)
		}
	}(o.Server)
	return err
}

// Catch the SIG_TERM and exit cleanly
func (o *GoMsHttpServer) GracefulStop() error {
	log.Println("[HTTP] Graceful Stop")
	return o.Server.Shutdown(o.Ctx.Main)
}
