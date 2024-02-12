package core

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

// Define the GRPC Server Struct
type GoMsGrpcServer struct {
	Ctx             *Context
	Host            string
	Port            int
	Server          *grpc.Server
	State           GoMsServerState
	listener        net.Listener
	services        []GoMsServiceInterface
	Opts            []grpc.DialOption
	Exporter        *Exporter
	Middlewares     map[string]Middleware
	MiddlewaresConf map[string]map[string][]string
}

// Create a grpc server
func NewGoMsGrpcServer(appCtx *Context, config *ServerConfig, opts []grpc.DialOption, middlewares map[string]Middleware) *GoMsGrpcServer {
	return &GoMsGrpcServer{
		Ctx:             appCtx,
		Host:            config.Host,
		Port:            config.Port,
		Server:          nil,
		State:           Init,
		listener:        nil,
		services:        make([]GoMsServiceInterface, 0),
		Opts:            opts,
		Exporter:        nil,
		Middlewares:     middlewares,
		MiddlewaresConf: make(map[string]map[string][]string),
	}
}

// Start listen tcp socket for handle grpc service
func (o *GoMsGrpcServer) Listen() (err error) {
	uri := fmt.Sprintf("%s:%d", o.Host, o.Port)
	o.listener, err = net.Listen("tcp", uri)
	if err != nil {
		o.State = Error
		log.Println("failed to listen: %v", err)
	}
	o.State = Listen
	log.Printf("[GRPC] Server listen on %s\n", uri)
	return err
}

// Set the exporter
func (o *GoMsGrpcServer) SetExporter(exporter *Exporter) {
	o.Exporter = exporter
}

// Register service on the grpc server
func (o *GoMsGrpcServer) Register(service GoMsServiceInterface) {
	service.GetHandler().RegisterGrpc(o, service)
}

// Add service to the local service array, need for register later
func (o *GoMsGrpcServer) AddService(service GoMsServiceInterface) {
	o.services = append(o.services, service)
	o.MiddlewaresConf[service.GetName()] = service.GetMiddlewaresConf()
}

// Register services to the grpc server
func (o *GoMsGrpcServer) startServices() {
	for _, service := range o.services {
		log.Printf("[GRPC] Register service %s\n", service.GetName())
		o.Register(service)
	}
}

func (o *GoMsGrpcServer) InitServer() {
	options := []grpc.ServerOption{
		grpc.UnaryInterceptor(chainInterceptors(loggingMiddleware, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return applyMiddleware(ctx, nil, req, info, handler, o.Middlewares, o.MiddlewaresConf)
		})),
		grpc.StreamInterceptor(chainStreamInterceptors(loggingStreamMiddleware, func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			_, e := applyMiddleware(stream.Context(), srv, stream, info, handler, o.Middlewares, o.MiddlewaresConf)
			return e
		})),
	}

	o.Server = grpc.NewServer(options...)
}

// Start a grpc server ready for handle connexion
func (o *GoMsGrpcServer) Start() error {
	o.InitServer()
	err := o.Listen()
	if err != nil {
		return err
	}
	o.startServices()
	go func(grpcServer *grpc.Server) {
		err := grpcServer.Serve(o.listener)
		// we can't catch this error, also we log it
		if err != nil {
			log.Fatal(err)
		}
	}(o.Server)
	o.State = Ready
	return nil
}

// Graceful stop, when SIG_TERM is send
func (o *GoMsGrpcServer) GracefulStop() error {
	log.Println("[GRPC] Graceful Stop")
	if o.isGracefulStopable() {
		o.Server.GracefulStop()
	}
	return nil
}

// Centralize GracefulStopable state
func (o *GoMsGrpcServer) isGracefulStopable() bool {
	switch o.State {
	case
		Ready,
		Listen:
		return true
	}
	return false
}
