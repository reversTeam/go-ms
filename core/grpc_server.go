package core

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// wrappedServerStream is a wrapper around grpc.ServerStream to override the Context method.
type wrappedServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.Ctx
}

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
		log.Printf("failed to listen: %v", err)
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

func (o *GoMsGrpcServer) UnaryErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	parentCtx, span := Trace(ctx, "gprc", info.FullMethod)
	defer span.End()

	middlewareCtx, middlewaresSpan := Trace(parentCtx, "gprc", "middlewares")
	res, err := applyMiddleware(middlewareCtx, nil, req, info, handler, o.Middlewares, o.MiddlewaresConf)
	if err != nil {
		ctx = middlewareCtx
	}
	middlewaresSpan.End()

	if err == nil {
		handlerCtx, handlderSpan := Trace(parentCtx, "gprc", "handler")
		res, err = handler(handlerCtx, req)
		handlderSpan.End()
		if err != nil {
			ctx = handlerCtx
		}
	}

	if err != nil {
		parentCtx = context.WithValue(ctx, grpcErrorKey, err)
		httpError := err.(*HttpError)
		if httpError.Code > 0 {
			if mdL, ok := metadata.FromIncomingContext(parentCtx); ok {
				md1 := metadata.Pairs("http-status-code", fmt.Sprintf("%d", httpError.Code))
				md = metadata.Join(mdL, md1)
			}
		}
	}

	if e := grpc.SendHeader(ctx, md); e != nil {
		log.Printf("Warning: %s", e)
	}

	return res, err
}

func StreamErrorInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	wrappedStream := &wrappedServerStream{ServerStream: ss}
	err := handler(srv, wrappedStream)
	if err != nil {
		wrappedStream.Ctx = context.WithValue(ss.Context(), grpcErrorKey, err)
	}
	return err
}

func (o *GoMsGrpcServer) InitServer() {
	options := []grpc.ServerOption{
		//*
		grpc.UnaryInterceptor(o.UnaryErrorInterceptor),
		grpc.StreamInterceptor(StreamErrorInterceptor),
		/*/
		grpc.UnaryInterceptor(chainInterceptors(loggingMiddleware, func(parentCtx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			requestCtx, sp := Trace(parentCtx, "gprc", "request-processus")
			defer sp.End()
			log.Printf("BEFORE APPLY MIDDLEWARE : %v\n", o.Middlewares)
			res, err := applyMiddleware(requestCtx, nil, req, info, handler, o.Middlewares, o.MiddlewaresConf)
			log.Println("AFTER APPLY MIDDLEWARE")

			if err == nil {
				res, err = handler(requestCtx, req)
			}

			if err != nil {
				// Vous pouvez traiter l'erreur ici si nécessaire avant de la stocker dans le contexte.
				httpError := err.(*HttpError)
				ctx := context.WithValue(parentCtx, "grpcError", httpError)
				if httpError.Code > 0 {
					log.Printf("ERROR = %v - %d\n", httpError, httpError.Code)
					md := metadata.Pairs("http-status-code", fmt.Sprintf("%d", httpError.Code))
					grpc.SendHeader(ctx, md)
				}
			}

			log.Printf("RESPONSE = %v\n", res)

			return res, err
		})),
		grpc.StreamInterceptor(chainStreamInterceptors(loggingStreamMiddleware, func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			ctx := stream.Context()
			ctx, sp := Trace(ctx, "gprc", "request-processus")
			defer sp.End()
			_, err := applyMiddleware(ctx, srv, stream, info, handler, o.Middlewares, o.MiddlewaresConf)
			return err
		})),
		//*/
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
		if err := grpcServer.Serve(o.listener); err != nil {
			// we can't catch this error, also we log it
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
