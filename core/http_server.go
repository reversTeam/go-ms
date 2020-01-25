package core

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
)

// Definition of GoMsHttpServer struct
type GoMsHttpServer struct {
	Ctx      context.Context
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
func NewGoMsHttpServer(ctx context.Context, host string, port int, grpcServer *GoMsGrpcServer) *GoMsHttpServer {
	uri := fmt.Sprintf("%s:%d", host, port)
	mux := http.NewServeMux()
	return &GoMsHttpServer{
		Ctx:  ctx,
		Grpc: grpcServer,
		Host: host,
		Port: port,
		Server: &http.Server{
			Addr:           uri,
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		State:    Init,
		mux:      mux,
		Mux:      runtime.NewServeMux(),
		services: make([]GoMsServiceInterface, 0),
		Exporter: nil,
	}
}

// Set the exporter
func (o *GoMsHttpServer) SetExporter(exporter *Exporter) {
	o.Exporter = exporter
}

// If the exporter is setup, add http handler for catch metrics
func (o *GoMsHttpServer) Handle(path string, mux *runtime.ServeMux) {
	// if o.exporter != nil {
	// o.mux.Handle(path, o.exporter.HandleHttpHandler(mux))
	// } else {
	o.mux.Handle(path, mux)
	// }
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
	log.Printf("[HTTP] Server listen on %s\n", uri)

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
	return o.Server.Shutdown(o.Ctx)
}
