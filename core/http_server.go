package core

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Definition of GoMsHttpServer struct
type GoMsHttpServer struct {
	Ctx      *Context
	Grpc     *GoMsGrpcServer
	Host     string
	Port     int
	Server   *http.Server
	State    GoMsServerState
	mux      *http.ServeMux
	Mux      *runtime.ServeMux
	services []GoMsServiceInterface
	Exporter *Exporter
}

// func CustomHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
// 	// Déterminer le code de statut et le message en fonction de l'erreur
// 	// var httpErr *HttpError

// 	// if grpcErr, ok := ctx.Value("grpcError").(HttpError); ok && grpcErr.Code != 0 {
// 	// 	log.Printf("GRPC ERR IS : %v \n", grpcErr)
// 	// }

// 	log.Printf("HTTP STATUS CODE: %v - %v\n", ctx.Value("httpStatusCode"))

// 	// log.Printf("REQUEST CONTEXT IN NUX : %v \n", req.Context().Value("statusCode"))
// 	log.Printf("HEADER: %v\n", w.Header())
// 	switch e := err.(type) {
// 	case *HttpError:
// 		log.Printf("IS HTTP ERROR: %v\n", e)
// 		// httpErr = &HttpError{Code: http.StatusForbidden, Message: e.Error()}
// 	default:
// 		log.Printf("NOT HTTP ERROR: %v\n", e)
// 		// httpErr = &HttpError{Code: 508, Message: err.Error()}
// 	}

// 	// Sérialiser HttpError en JSON pour la réponse
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("x-test-ok", "pio")

// 	req = req.WithContext(ctx)
// 	rwh := NewResponseWriterHandler(w)
// 	log.Printf("WRITE STATUS CODE: %v\n", rwh.StatusCode)

// 	log.Printf("ICI: %v - %v\n", ctx.Value("httpStatusCode"), rwh.Header())
// 	md, _ := metadata.FromIncomingContext(req.Context())
// 	log.Printf("REQUEST CONTEXT IN METADATA : %v \n", md)
// 	log.Printf("REQUEST CONTEXT IN CTX : %v \n", ctx.Value("statusCode"))
// 	log.Printf("REQUEST CONTEXT IN REQ : %v \n", req.Context().Value("statusCode"))

// 	// next.ServeHTTP(rwh, req)

// 	// rwh.WriteHeader(httpErr.Code)
// 	// if err := json.NewEncoder(w).Encode(httpErr); err != nil {
// 	// 	log.Printf("Échec de l'encodage de l'erreur JSON: %v", err)
// 	// }
// 	// defer rwh.Finalize(ctx)
// }

// func HttpMiddlewareHandle(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		md, ok := runtime.ServerMetadataFromContext(r.Context())
// 		if ok && md.HeaderMD != nil {
// 			if errorCode, exists := md.HeaderMD["http-status-code"]; exists {
// 				w.Header().Set("X-Error-Code", errorCode[0])
// 			}
// 		}
// 		log.Printf("---> METADATA: %v\n", md)

// 		rwh := NewResponseWriterHandler(w)
// 		ctx := r.Context()
// 		r = r.WithContext(ctx)

// 		next.ServeHTTP(rwh, r)

// 		// log.Printf("ICI: %v - %v\n", ctx.Value("httpStatusCode"), rwh.Header())
// 		// h.ServeHTTP(w, r)
// 	})
// }

// func DefaultHeaderMatcher(key string) (string, bool) {
// 	// 	switch key = textproto.CanonicalMIMEHeaderKey(key); {
// 	// 	case isPermanentHTTPHeader(key):
// 	// 		return MetadataPrefix + key, true
// 	// 	case strings.HasPrefix(key, MetadataHeaderPrefix):
// 	// 		return key[len(MetadataHeaderPrefix):], true
// 	// 	}
// 	log.Printf("CANONICAL : %v\n", textproto.CanonicalMIMEHeaderKey(key))
// 	return "", false
// }

// func WithForwardResponseOption(ctx context.Context, w http.ResponseWriter, msg protoreflect.ProtoMessage) error {
// 	log.Printf("FORWARD RESPONSE EACH PART: %+v\n", msg)

// 	md, _ := metadata.FromIncomingContext(ctx)
// 	log.Printf("METADATA: %v\n", md)

// 	return nil
// }

// Init GoMsHttpServer
func NewGoMsHttpServer(ctx *Context, config *HttpServerConfig, grpcServer *GoMsGrpcServer) *GoMsHttpServer {
	uri := fmt.Sprintf("%s:%d", config.Host, config.Port)
	mux := http.NewServeMux()
	muxOpts := []runtime.ServeMuxOption{
		// runtime.WithForwardResponseOption(WithForwardResponseOption),
		runtime.WithMetadata(forwardHeaders),
		// runtime.WithErrorHandler(CustomHTTPErrorHandler),
		// runtime.WithOutgoingHeaderMatcher(DefaultHeaderMatcher),
	}

	return &GoMsHttpServer{
		Ctx:  ctx,
		Grpc: grpcServer,
		Host: config.Host,
		Port: config.Port,
		Server: &http.Server{
			Addr:           uri,
			Handler:        mux,
			ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		State:    Init,
		mux:      mux,
		Mux:      runtime.NewServeMux(muxOpts...),
		services: make([]GoMsServiceInterface, 0),
		Exporter: nil,
	}
}

// Set the exporter
func (o *GoMsHttpServer) SetExporter(exporter *Exporter) {
	o.Exporter = exporter
}

func (o *GoMsHttpServer) Handle(path string, mux *runtime.ServeMux) {
	o.mux.Handle(path, tracingMiddleware(mux))
	// o.mux.Handle(path, HttpMiddlewareHandle(mux))
}

// Register service on the http server
func (o *GoMsHttpServer) Register(service GoMsServiceInterface) error {
	return service.GetHandler().RegisterHttp(o, service)
}

// Add service in the local services list
func (o *GoMsHttpServer) AddService(service GoMsServiceInterface) {
	o.services = append(o.services, service)
}

// Register each service in the http server
func (o *GoMsHttpServer) startServices() error {
	for _, service := range o.services {
		log.Printf("[HTTP] Register service %s\n", service.GetName())
		err := o.Register(service)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start the http server, ready for handle connexion
func (o *GoMsHttpServer) Start() error {
	err := o.startServices()
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("%s:%d", o.Host, o.Port)
	log.Printf("[HTTP] Server listen on http://%s\n", uri)

	o.Handle("/", o.Mux)

	go func(httpServer *http.Server) {
		err := httpServer.ListenAndServe()
		// we can't catch this error, also we log it
		if err != nil {
			log.Println("[HTTP] Error listen: ", err)
		}
	}(o.Server)
	return err
}

// Catch the SIG_TERM and exit cleanly
func (o *GoMsHttpServer) GracefulStop() error {
	log.Println("[HTTP] Graceful Stop")
	return o.Server.Shutdown(o.Ctx.Main)
}
