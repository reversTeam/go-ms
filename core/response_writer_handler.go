package core

import (
	"fmt"
	"net/http"
	"strings"
)

type ResponseWriterHandler struct {
	http.ResponseWriter
	StatusCode     int
	headersCleaned bool
}

func NewResponseWriterHandler(w http.ResponseWriter) *ResponseWriterHandler {
	return &ResponseWriterHandler{ResponseWriter: w, StatusCode: http.StatusOK}
}

func (rw *ResponseWriterHandler) WriteHeader(code int) {
	if !rw.headersCleaned {
		rw.cleanGrpcHeaders()
		rw.headersCleaned = true
	}
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriterHandler) Write(b []byte) (int, error) {
	if !rw.headersCleaned {
		rw.cleanGrpcHeaders()
		rw.headersCleaned = true
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriterHandler) cleanGrpcHeaders() {
	headers := rw.ResponseWriter.Header()
	for name := range headers {
		if strings.HasPrefix(strings.ToLower(name), "grpc-metadata-") {
			newName := name[14:]

			if _, ok := headers[newName]; !ok {
				newName = fmt.Sprintf("x-%s", newName)
				values := headers[name]
				delete(headers, name)
				for _, value := range values {
					rw.ResponseWriter.Header().Add(newName, value)
				}
			} else {
				delete(headers, name)
			}
		}
	}
}
