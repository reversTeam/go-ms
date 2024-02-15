package core

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GrpcClientManager struct {
	connections    map[string]*grpc.ClientConn
	configurations map[string]*GrpcClientConfiguration
	clients        map[string]any
	muConf         *sync.RWMutex
	muConn         *sync.RWMutex
	muClient       *sync.RWMutex
}

type GrpcClientConfiguration struct {
	Host    string
	Port    int
	Timeout time.Duration
}

func NewGrpcClientManager() *GrpcClientManager {
	return &GrpcClientManager{
		connections:    nil,
		configurations: make(map[string]*GrpcClientConfiguration),
		clients:        make(map[string]any),
		muConf:         new(sync.RWMutex),
		muConn:         new(sync.RWMutex),
		muClient:       new(sync.RWMutex),
	}
}

func (o *GrpcClientManager) AddServer(name string, host string, port int, timeout time.Duration) {
	o.muConf.RLock()
	_, exists := o.configurations[name]
	o.muConf.RUnlock()
	if exists {
		return
	}
	conf := &GrpcClientConfiguration{
		Host:    host,
		Port:    port,
		Timeout: timeout,
	}

	o.muConf.Lock()
	o.configurations[name] = conf
	o.muConf.Unlock()
}

func (o *GrpcClientManager) AddClient(name string, client any) {
	o.muConf.RLock()
	_, exists := o.clients[name]
	o.muConf.RUnlock()
	if exists {
		return
	}

	o.muConf.Lock()
	o.clients[name] = client
	o.muConf.Unlock()
}

func (o *GrpcClientManager) GetClient(name string) (any, error) {
	o.muClient.RLock()
	client, exists := o.clients[name]
	o.muClient.RUnlock()
	if !exists {
		return nil, fmt.Errorf("[GRPC] client not found : %s\n", name)
	}
	return client, nil
}

func (o *GrpcClientManager) GetConnection(name string) (*grpc.ClientConn, error) {
	o.muConn.RLock()
	conn, exists := o.connections[name]
	o.muConn.RUnlock()
	if !exists {
		return nil, fmt.Errorf("[GRPC] connections not found : %s\n", name)
	}
	return conn, nil
}

func (o *GrpcClientManager) InitConnections() error {
	o.muConf.RLock()
	confs := o.configurations
	o.muConf.RUnlock()

	connections := make(map[string]*grpc.ClientConn)
	for name, conf := range confs {
		addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			return err
		}

		connections[name] = conn
	}

	o.muConn.Lock()
	o.connections = connections
	o.muConn.Unlock()
	return nil
}

func Call[T any](ctx context.Context, cm *GrpcClientManager, serviceName, methodName string, params any) (T, error) {
	client, err := cm.GetClient(serviceName)
	if err != nil {
		var zero T
		return zero, err
	}

	clientValue := reflect.ValueOf(client)
	methodValue := clientValue.MethodByName(methodName)

	if !methodValue.IsValid() {
		var zero T
		return zero, fmt.Errorf("method %s was not found for service %s", methodName, serviceName)
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, MetadataReaderWriter{md})

	ctx = metadata.NewOutgoingContext(ctx, md)

	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(params),
	}

	responseValues := methodValue.Call(args)
	if len(responseValues) != 2 {
		var zero T
		return zero, fmt.Errorf("unexpected response from method %s for service %s", methodName, serviceName)
	}

	if errValue := responseValues[1].Interface(); errValue != nil {
		var zero T
		return zero, errValue.(error)
	}

	response := responseValues[0].Interface()
	if response, ok := response.(T); ok {
		return response, nil
	} else {
		var zero T
		return zero, fmt.Errorf("unable to convert response to expected type for method %s of service %s", methodName, serviceName)
	}
}
