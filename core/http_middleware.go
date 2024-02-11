package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

func tracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		path := r.URL.Path
		host := r.Host

		ctx, span := Trace(r.Context(), "http", fmt.Sprintf("[%s]%s:%s", method, host, path))
		defer span.End()

		span.SetAttributes(attribute.String("http.method", method))
		span.SetAttributes(attribute.String("http.path", path))
		span.SetAttributes(attribute.String("http.host", host))
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))
		r = r.WithContext(ctx)

		rwh := NewResponseWriterHandler(w)
		defer rwh.cleanGrpcHeaders()
		next.ServeHTTP(rwh, r)
	})
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
