package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"log"
)

type ResponseWriterHandler struct {
	http.ResponseWriter
	StatusCode     int
	headersCleaned bool
	buffer         bytes.Buffer
	entities       []interface{}
}

func NewResponseWriterHandler(w http.ResponseWriter) *ResponseWriterHandler {
	return &ResponseWriterHandler{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		entities:       make([]interface{}, 0),
	}
}

func (o *ResponseWriterHandler) WriteHeader(code int) {
	headers := o.Header().Values("Grpc-Metadata-Http-Status-Code")
	if len(headers) > 0 {
		if c, err := strconv.Atoi(headers[0]); err == nil {
			o.StatusCode = c
		}
	} else {
		o.StatusCode = 200
	}

	o.ResponseWriter.WriteHeader(o.StatusCode)
}

func (o *ResponseWriterHandler) Write(b []byte) (int, error) {
	if !o.headersCleaned {
		o.headersCleaned = true
		o.cleanGrpcHeaders()
	}
	return o.buffer.Write(b)
}

func (o *ResponseWriterHandler) Flush() {
	if flusher, ok := o.ResponseWriter.(http.Flusher); ok {
		o.emitBufferedData()
		flusher.Flush()
	}
}

func (o *ResponseWriterHandler) emitBufferedData() {
	data := o.buffer.Bytes()
	if len(data) == 0 {
		return
	}

	var testResult struct {
		Result json.RawMessage `json:"result"`
	}

	if err := json.Unmarshal(data, &testResult); err != nil {
		o.ResponseWriter.WriteHeader(o.StatusCode)
		return
	}

	o.buffer.Reset()

	var result []json.RawMessage

	if testResult.Result[0] == '[' {
		if err := json.Unmarshal(testResult.Result, &result); err != nil {
			o.ResponseWriter.WriteHeader(o.StatusCode)
			return
		}
	} else {
		o.entities = append(o.entities, testResult.Result)
	}
}

func (o *ResponseWriterHandler) Finalize(ctx context.Context) {
	isStream := o.ResponseWriter.Header().Get("Transfer-Encoding") == "chunked"

	if isStream && o.StatusCode == 200 {
		if len(o.entities) == 0 {
			if _, err := o.ResponseWriter.Write([]byte("[]")); err != nil {
				log.Printf("ERROR occurred on Finalize HTTP request: %s\n", err)
			}
		} else if encodedData, err := json.Marshal(o.entities); err == nil {
			if _, err = o.ResponseWriter.Write(encodedData); err != nil {
				log.Printf("ERROR occurred on Finalize HTTP request: %s\n", err)
			}
		} else {
			o.ResponseWriter.WriteHeader(o.StatusCode)
		}
	} else {
		if _, err := o.ResponseWriter.Write(o.buffer.Bytes()); err != nil {
			log.Printf("ERROR occurred on Finalize HTTP request: %s\n", err)
		}
	}
}

func (o *ResponseWriterHandler) cleanGrpcHeaders() {
	headers := o.ResponseWriter.Header()
	for name := range headers {
		if strings.HasPrefix(strings.ToLower(name), "grpc-metadata-") {
			newName := name[14:]

			if _, ok := headers[newName]; !ok {
				newName = fmt.Sprintf("x-%s", newName)
				values := headers[name]
				delete(headers, name)
				for _, value := range values {
					o.ResponseWriter.Header().Set(newName, value)
				}
			} else {
				delete(headers, name)
			}
		}
	}
}
