package core

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
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

func forwardHeaders(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	excludeHeaders := map[string]bool{
		"connection":        true,
		"keep-alive":        true,
		"proxy-connection":  true,
		"transfer-encoding": true,
		"upgrade":           true,
	}

	for name, values := range req.Header {
		if _, ok := excludeHeaders[strings.ToLower(name)]; !ok {
			for _, value := range values {
				md.Append(strings.ToLower(name), value)
			}
		}
	}
	return md
}

// Init GoMsHttpServer
func NewGoMsHttpServer(ctx *Context, host string, port int, grpcServer *GoMsGrpcServer) *GoMsHttpServer {
	uri := fmt.Sprintf("%s:%d", host, port)
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithMetadata(forwardHeaders),
	}
	mux := http.NewServeMux()

	return &GoMsHttpServer{
		Ctx:  ctx,
		Grpc: grpcServer,
		Host: host,
		Port: port,
		Server: &http.Server{
			Addr:           uri,
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
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

// // TODO : middleware example trace
func (o *GoMsHttpServer) trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		path := r.URL.Path
		host := r.Host

		tracer := otel.Tracer("http")
		newCtx, span := tracer.Start(r.Context(), fmt.Sprintf("[%s]%s:%s", method, host, path), trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(attribute.String("http.method", method))
		span.SetAttributes(attribute.String("http.path", path))
		span.SetAttributes(attribute.String("http.host", host))
		otel.GetTextMapPropagator().Inject(newCtx, propagation.HeaderCarrier(r.Header))
		defer span.End()
		r = r.WithContext(newCtx)

		rwh := NewResponseWriterHandler(w)
		next.ServeHTTP(rwh, r)
	})
}

func (o *GoMsHttpServer) Handle(path string, mux *runtime.ServeMux) {
	o.mux.Handle(path, o.trace(mux))
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
