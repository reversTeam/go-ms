package core

import (
	"context"

	"google.golang.org/grpc"
)

type GoMsMiddlewareFunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
type GoMsServiceFunc func(*Context, string, ServiceConfig) GoMsServiceInterface
