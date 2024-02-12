package core

import (
	"context"

	"google.golang.org/grpc"
)

type GoMsStreamMiddlewareFunc func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
type GoMsMiddlewareFunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
type GoMsServiceFunc func(*Context, string, ServiceConfig) GoMsServiceInterface
