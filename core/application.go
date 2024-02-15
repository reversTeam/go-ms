package core

import (
	"log"
	"time"

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
	clientManager       *GrpcClientManager
}

func NewApplication(config *Config, services map[string]GoMsServiceFunc, middlewares map[string]Middleware) *Application {
	ctx := NewContext(config.Name, config.Jaeger)
	var grpcServer *GoMsGrpcServer = nil
	var httpServer *GoMsHttpServer = nil

	if config.Grpc != nil {
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		grpcServer = NewGoMsGrpcServer(ctx, config.Grpc, opts, middlewares)
	}

	if config.Http != nil {
		httpServer = NewGoMsHttpServer(ctx, config.Http, grpcServer)
	}

	app := &Application{
		ctx:                 ctx,
		config:              *config,
		grpcServer:          grpcServer,
		httpServer:          httpServer,
		Services:            make(map[string]GoMsServiceInterface),
		servicesConstructor: make(map[string]GoMsServiceFunc),
		clientManager:       NewGrpcClientManager(),
	}

	app.RegisterServices(services)

	AddServerGracefulStop(httpServer)
	AddServerGracefulStop(grpcServer)

	return app
}

func NewApplicationFromConfigFile(configPath string, services map[string]GoMsServiceFunc, middlewares map[string]Middleware) *Application {
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

	if err := o.clientManager.InitConnections(); err != nil {
		log.Fatal(err)
	}

	for name, service := range o.Services {
		conn, err := o.clientManager.GetConnection(name)
		if err != nil {
			log.Fatal(err)
		}
		o.clientManager.AddClient(name, service.GetClient(conn))
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
	o.clientManager.AddServer(name, o.config.Grpc.Host, o.config.Grpc.Port, 50*time.Second)
	o.Services[name].SetClientManager(o.clientManager)
}

func (o *Application) InitService(name string, config ServiceConfig) GoMsServiceInterface {
	if constructor, ok := o.servicesConstructor[name]; ok {
		service := constructor(o.ctx, name, config)

		o.grpcServer.AddService(service)

		if config.Http {
			o.httpServer.AddService(service)
		}

		return service
	}
	return nil
}

func RegisterServiceMap[T GoMsServiceInterface](constructor func(*Context, string, ServiceConfig) T) func(*Context, string, ServiceConfig) T {
	return func(ctx *Context, name string, config ServiceConfig) T {
		return constructor(ctx, name, config)
	}
}
