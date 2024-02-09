package core

type GoMsHandlerInterface interface {
	RegisterHttp(*GoMsHttpServer, GoMsServiceInterface) error
	RegisterGrpc(*GoMsGrpcServer, GoMsServiceInterface)
}

// Definition of service interface for register http & grpc server
type GoMsServiceInterface interface {
	GetName() string
	GetHandler() GoMsHandlerInterface

	RegisterHttp(*GoMsHttpServer, string) error
	RegisterGrpc(*GoMsGrpcServer)

	Log(message string)
}

// Defition of ServerGracefulStopableInterface for http & grpc server graceful stop
type GoMsServerGracefulStopableInterface interface {
	GracefulStop() error
}

// Definition of GoMsMetricsInterface
type GoMsMetricsInterface interface {
	GetServiceName() string
}
