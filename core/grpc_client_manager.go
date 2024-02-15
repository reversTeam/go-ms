package core

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
