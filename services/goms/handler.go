package goms

import (
	"fmt"
	"github.com/reversTeam/go-ms/core"
)

type GoMsHandler struct{}

func NewHandler() core.GoMsHandlerInterface {
	return &GoMsHandler{}
}

func (o *GoMsHandler) RegisterHttp(gh *core.GoMsHttpServer, s core.GoMsServiceInterface) error {
	endpoint := fmt.Sprintf("%s:%d", gh.Grpc.Host, gh.Grpc.Port)
	return s.RegisterHttp(gh, endpoint)
}

func (o *GoMsHandler) RegisterGrpc(gs *core.GoMsGrpcServer, s core.GoMsServiceInterface) {
	s.RegisterGrpc(gs)
}
