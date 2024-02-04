package child

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/reversTeam/go-ms/core"
	pb "github.com/reversTeam/go-ms/services/child/protobuf"
	"github.com/reversTeam/go-ms/services/goms"
	ms "github.com/reversTeam/go-ms/services/goms/protobuf"
	"golang.org/x/net/context"
)

// Define the service structure
type ChildService struct {
	*goms.GoMsService
	config core.ServiceConfig
}

// Instanciate the service without dependency because it's role of ServiceFactory
func NewService(name string, config core.ServiceConfig) *ChildService {
	s := &ChildService{
		GoMsService: goms.NewService(name, config),
	}

	return s
}

// This method is required for redister your service on the Http server
func (o *ChildService) RegisterHttp(gh *core.GoMsHttpServer, endpoint string) error {
	return pb.RegisterChildHandlerFromEndpoint(gh.Ctx, gh.Mux, endpoint, gh.Grpc.Opts)
}

// This method is required for redister your service on the Grpc server
func (o *ChildService) RegisterGrpc(gs *core.GoMsGrpcServer) {
	pb.RegisterChildServer(gs.Server, o)
}

// Endpoint :
//   - grpc : List
//   - http : Get /child
func (o *ChildService) List(ctx context.Context, in *empty.Empty) (*ms.GoMsResponse, error) {
	return &ms.GoMsResponse{
		Message: "Child List",
	}, nil
}

// Endpoint :
//   - grpc : Create
//   - http : POST /child
func (o *ChildService) Create(ctx context.Context, in *empty.Empty) (*ms.GoMsResponse, error) {
	return &ms.GoMsResponse{
		Message: "Child Create",
	}, nil
}

// Endpoint :
//   - grpc : Get
//   - http : GET /child/{id}
func (o *ChildService) Get(ctx context.Context, in *ms.GoMsEntityRequest) (*ms.GoMsResponse, error) {
	return &ms.GoMsResponse{
		Message: "Child View",
	}, nil
}

// Endpoint :
//   - grpc : Update
//   - http : PATCH /child/{id}
func (o *ChildService) Update(ctx context.Context, in *ms.GoMsEntityRequest) (*ms.GoMsResponse, error) {
	return &ms.GoMsResponse{
		Message: "Child Update",
	}, nil
}

// Endpoint :
//   - grpc : Delete
//   - http : PATCH /child/{id}
func (o *ChildService) Delete(ctx context.Context, in *ms.GoMsEntityRequest) (*ms.GoMsResponse, error) {
	return &ms.GoMsResponse{
		Message: "Child Delete",
	}, nil
}
