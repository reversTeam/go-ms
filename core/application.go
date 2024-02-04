package core

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

type Application struct {
	ctx                 context.Context
	config              Config
	grpcServer          *GoMsGrpcServer
	httpServer          *GoMsHttpServer
	Services            map[string]GoMsServiceInterface
	servicesConstructor map[string]func(string, ServiceConfig) GoMsServiceInterface
}

func NewApplication(config *Config, services map[string]func(string, ServiceConfig) GoMsServiceInterface) *Application {
	ctx := context.Background()
	var grpcServer *GoMsGrpcServer = nil
	var httpServer *GoMsHttpServer = nil

	if config.Grpc != nil {
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		grpcServer = NewGoMsGrpcServer(ctx, config.Grpc.Host, config.Grpc.Port, opts)
	}

	if config.Http != nil {
		httpServer = NewGoMsHttpServer(ctx, config.Http.Host, config.Http.Port, grpcServer)
	}

	app := &Application{
		ctx:                 ctx,
		config:              *config,
		grpcServer:          grpcServer,
		httpServer:          httpServer,
		Services:            make(map[string]GoMsServiceInterface),
		servicesConstructor: make(map[string]func(string, ServiceConfig) GoMsServiceInterface),
	}

	app.RegisterServices(services)

	AddServerGracefulStop(httpServer)
	AddServerGracefulStop(grpcServer)

	return app
}

func NewApplicationFromConfigFile(configPath string, services map[string]func(string, ServiceConfig) GoMsServiceInterface) *Application {
	config, err := NewConfig(configPath)
	if err != nil {
		log.Panic(err)
	}

	return NewApplication(config, services)
}

func (o *Application) Start() {
	done := CatchStopSignals()

	err := o.httpServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = o.grpcServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func (o *Application) RegisterServices(services map[string]func(string, ServiceConfig) GoMsServiceInterface) {
	for serviceName, serviceFunc := range services {
		o.RegisterService(serviceName, serviceFunc)
	}
}

func (o *Application) RegisterService(name string, constructor func(string, ServiceConfig) GoMsServiceInterface) {
	o.servicesConstructor[name] = constructor
	o.Services[name] = o.InitService(name, o.config.Services[name])
}

func (o *Application) InitService(name string, config ServiceConfig) GoMsServiceInterface {
	if constructor, ok := o.servicesConstructor[name]; ok {
		service := constructor(name, config)

		if config.Grpc == true {
			o.grpcServer.AddService(service)
		}

		if config.Http == true {
			o.httpServer.AddService(service)
		}

		return service
	}
	return nil
}

func RegisterServiceMap[T GoMsServiceInterface](constructor func(string, ServiceConfig) T) func(string, ServiceConfig) T {
	return func(name string, config ServiceConfig) T {
		log.Printf("[%s] %v", name, config)
		return constructor(name, config)
	}
}
