package core

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Application struct {
	ctx                 *Context
	config              Config
	grpcServer          *GoMsGrpcServer
	httpServer          *GoMsHttpServer
	Services            map[string]GoMsServiceInterface
	servicesConstructor map[string]GoMsServiceFunc
}

func NewApplication(config *Config, services map[string]GoMsServiceFunc, middlewares map[string]GoMsMiddlewareFunc) *Application {
	ctx := NewContext(config.Name, config.Jaeger)
	var grpcServer *GoMsGrpcServer = nil
	var httpServer *GoMsHttpServer = nil

	if config.Grpc != nil {
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		grpcServer = NewGoMsGrpcServer(ctx, config.Grpc.Host, config.Grpc.Port, opts, middlewares)
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
		servicesConstructor: make(map[string]GoMsServiceFunc),
	}

	app.RegisterServices(services)

	AddServerGracefulStop(httpServer)
	AddServerGracefulStop(grpcServer)

	return app
}

func NewApplicationFromConfigFile(configPath string, services map[string]GoMsServiceFunc, middlewares map[string]GoMsMiddlewareFunc) *Application {
	config, err := NewConfig(configPath)
	if err != nil {
		log.Panic(err)
	}

	return NewApplication(config, services, middlewares)
}

func (o *Application) Start() {
	done := CatchStopSignals()

	err := o.grpcServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = o.httpServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	<-done
	o.ctx.Jeager.GracefullFunc()
}

func (o *Application) RegisterServices(services map[string]GoMsServiceFunc) {
	for serviceName, serviceFunc := range services {
		o.RegisterService(serviceName, serviceFunc)
	}
}

func (o *Application) RegisterService(name string, constructor GoMsServiceFunc) {
	o.servicesConstructor[name] = constructor
	o.Services[name] = o.InitService(name, o.config.Services[name])
}

func (o *Application) InitService(name string, config ServiceConfig) GoMsServiceInterface {
	if constructor, ok := o.servicesConstructor[name]; ok {
		service := constructor(o.ctx, name, config)

		o.grpcServer.AddService(service)

		if config.Http == true {
			o.httpServer.AddService(service)
		}

		return service
	}
	return nil
}

func RegisterServiceMap[T GoMsServiceInterface](constructor func(*Context, string, ServiceConfig) T) func(*Context, string, ServiceConfig) T {
	return func(ctx *Context, name string, config ServiceConfig) T {
		log.Printf("[%s] %v", name, config)
		return constructor(ctx, name, config)
	}
}
