package core

import (
	"context"

	"google.golang.org/grpc"
)

type GoMsHandlerInterface interface {
	RegisterHttp(*GoMsHttpServer, GoMsServiceInterface) error
	RegisterGrpc(*GoMsGrpcServer, GoMsServiceInterface)
}

// Definition of service interface for register http & grpc server
type GoMsServiceInterface interface {
	GetName() string
	GetHandler() GoMsHandlerInterface

	SetClientManager(*GrpcClientManager)

	RegisterHttp(*GoMsHttpServer, string) error
	RegisterGrpc(*GoMsGrpcServer)
	GetClient(conn *grpc.ClientConn) any

	GetMiddlewaresConf() map[string][]string
}

// Defition of ServerGracefulStopableInterface for http & grpc server graceful stop
type GoMsServerGracefulStopableInterface interface {
	GracefulStop() error
}

// Definition of GoMsMetricsInterface
type GoMsMetricsInterface interface {
	GetServiceName() string
}

type Middleware interface {
	Apply(ctx context.Context, req interface{}) (context.Context, interface{}, error)
	Unary() grpc.UnaryServerInterceptor
	Stream() grpc.StreamServerInterceptor
}
