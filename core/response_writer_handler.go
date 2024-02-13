package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

func (rw *ResponseWriterHandler) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriterHandler) Write(b []byte) (int, error) {
	return rw.buffer.Write(b)
}

func (rw *ResponseWriterHandler) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		rw.emitBufferedData()
		flusher.Flush()
	}
}

func (rw *ResponseWriterHandler) emitBufferedData() {
	data := rw.buffer.Bytes()
	if len(data) == 0 {
		return
	}

	var testResult struct {
		Result json.RawMessage `json:"result"`
	}

	if err := json.Unmarshal(data, &testResult); err != nil {
		rw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.buffer.Reset()

	var result []json.RawMessage

	if testResult.Result[0] == '[' {
		if err := json.Unmarshal(testResult.Result, &result); err != nil {
			rw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		rw.entities = append(rw.entities, testResult.Result)
	}
}

func (rw *ResponseWriterHandler) Finalize() {
	isStream := rw.ResponseWriter.Header().Get("Transfer-Encoding") == "chunked"
	if !rw.headersCleaned {
		rw.headersCleaned = true
		rw.cleanGrpcHeaders()
	}

	if isStream && rw.StatusCode == 200 {
		if len(rw.entities) == 0 {
			if _, err := rw.ResponseWriter.Write([]byte("[]")); err != nil {
				log.Printf("ERROR occurred on Finalize HTTP request: %s\n", err)
			}
		} else if encodedData, err := json.Marshal(rw.entities); err == nil {
			if _, err = rw.ResponseWriter.Write(encodedData); err != nil {
				log.Printf("ERROR occurred on Finalize HTTP request: %s\n", err)
			}
		} else {
			rw.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		if _, err := rw.ResponseWriter.Write(rw.buffer.Bytes()); err != nil {
			log.Printf("ERROR occurred on Finalize HTTP request: %s\n", err)
		}
	}
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
