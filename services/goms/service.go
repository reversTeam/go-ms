package goms

import (
	// "fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/reversTeam/go-ms/core"
	pb "github.com/reversTeam/go-ms/services/goms/protobuf"
	"golang.org/x/net/context"
)

// Define the service structure
type GoMsService struct {
	Name    string
	Handler core.GoMsHandlerInterface
	Metrics *GoMsMetrics
}

// Instanciate the service without dependency because it's role of ServiceFactory
func NewService(name string) *GoMsService {
	o := &GoMsService{
		Name:    name,
		Handler: NewHandler(),
		Metrics: NewGoMsMetrics(name),
	}
	return o
}

// Get the service name
func (o *GoMsService) GetName() string {
	return o.Name
}

// Get Handler describe in the interface
func (o *GoMsService) GetHandler() core.GoMsHandlerInterface {
	return o.Handler
}

// This method is required for redister your service on the Http server
func (o *GoMsService) RegisterHttp(gh *core.GoMsHttpServer, endpoint string) error {
	return pb.RegisterGoMsHandlerFromEndpoint(gh.Ctx, gh.Mux, endpoint, gh.Grpc.Opts)
}

// This method is required for redister your service on the Grpc server
func (o *GoMsService) RegisterGrpc(gs *core.GoMsGrpcServer) {
	pb.RegisterGoMsServer(gs.Server, o)
}

// Endpoint :
//  - grpc : List
//  - http : Get /goms
func (o *GoMsService) List(ctx context.Context, in *empty.Empty) (*pb.GoMsResponse, error) {
	return &pb.GoMsResponse{
		Message: "GoMs List",
	}, nil
}

// Endpoint :
//  - grpc : Create
//  - http : POST /goms
func (o *GoMsService) Create(ctx context.Context, in *empty.Empty) (*pb.GoMsResponse, error) {
	return &pb.GoMsResponse{
		Message: "GoMs Create",
	}, nil
}

// Endpoint :
//  - grpc : Get
//  - http : GET /goms/{id}
func (o *GoMsService) Get(ctx context.Context, in *pb.GoMsEntityRequest) (*pb.GoMsResponse, error) {
	return &pb.GoMsResponse{
		Message: "GoMs View",
	}, nil
}

// Endpoint :
//  - grpc : Update
//  - http : PATCH /goms/{id}
func (o *GoMsService) Update(ctx context.Context, in *pb.GoMsEntityRequest) (*pb.GoMsResponse, error) {
	return &pb.GoMsResponse{
		Message: "GoMs Update",
	}, nil
}

// Endpoint :
//  - grpc : Delete
//  - http : PATCH /goms/{id}
func (o *GoMsService) Delete(ctx context.Context, in *pb.GoMsEntityRequest) (*pb.GoMsResponse, error) {
	return &pb.GoMsResponse{
		Message: "GoMs Delete",
	}, nil
}
