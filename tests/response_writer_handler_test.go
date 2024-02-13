package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reversTeam/go-ms/core"
	"github.com/stretchr/testify/assert"
)

func TestResponseWriterHandler_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := core.NewResponseWriterHandler(recorder)

	rw.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestResponseWriterHandler_WriteAndFlush_Withoutdata(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Transfer-Encoding", "chunked")
	recorder.Header().Set("Content-Type", "application/json")
	rw := core.NewResponseWriterHandler(recorder)

	rw.Flush()
	rw.Finalize()

	responseBody := recorder.Body.Bytes()
	assert.NotEmpty(t, responseBody, "The body should not be empty after flushing and finalizing")

	assert.Equal(t, http.StatusOK, recorder.Code, "The HTTP status code should be 200 OK")

	responseBodyStr := string(responseBody)

	expectedOutput := `[]`
	assert.Contains(t, responseBodyStr, expectedOutput, "The response body should contain the expected JSON output")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Content-Type should be application/json")
}

func TestResponseWriterHandler_WriteAndFlush(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Transfer-Encoding", "chunked")
	recorder.Header().Set("Content-Type", "application/json")
	rw := core.NewResponseWriterHandler(recorder)

	data := []byte(`{"result":{"key":"value"}}`)
	_, err := rw.Write(data)
	assert.NoError(t, err)

	rw.Flush()
	rw.Finalize()

	responseBody := recorder.Body.Bytes()
	assert.NotEmpty(t, responseBody, "The body should not be empty after flushing and finalizing")

	assert.Equal(t, http.StatusOK, recorder.Code, "The HTTP status code should be 200 OK")

	responseBodyStr := string(responseBody)

	expectedOutput := `[{"key":"value"}]`
	assert.Contains(t, responseBodyStr, expectedOutput, "The response body should contain the expected JSON output")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Content-Type should be application/json")
}

func TestResponseWriterHandler_MultipleWriteAndFlush(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Transfer-Encoding", "chunked")
	recorder.Header().Set("Content-Type", "application/json")
	rw := core.NewResponseWriterHandler(recorder)

	data := []byte(`{"result":{"key":"value"}}`)
	_, err := rw.Write(data)
	assert.NoError(t, err)
	rw.Flush()

	data = []byte(`{"result":{"key":"value"}}`)
	_, err = rw.Write(data)
	assert.NoError(t, err)
	rw.Flush()

	rw.Finalize()

	responseBody := recorder.Body.Bytes()
	assert.NotEmpty(t, responseBody, "The body should not be empty after flushing and finalizing")

	assert.Equal(t, http.StatusOK, recorder.Code, "The HTTP status code should be 200 OK")

	responseBodyStr := string(responseBody)

	expectedOutput := `[{"key":"value"},{"key":"value"}]`
	assert.Contains(t, responseBodyStr, expectedOutput, "The response body should contain the expected JSON output")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Content-Type should be application/json")
}

func TestResponseWriterHandler_HeaderManipulation(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := core.NewResponseWriterHandler(recorder)

	recorder.Header().Set("Grpc-Metadata-Test", "value")

	rw.Finalize()

	assert.Empty(t, recorder.Header().Get("Grpc-Metadata-Test"), "Grpc-Metadata-Test header should be removed or transformed")
	assert.Equal(t, "value", recorder.Header().Get("x-test"), "Grpc-Metadata-Test should be X-Test")
}

func TestResponseWriterHandler_Finalize(t *testing.T) {
	recorder := httptest.NewRecorder()
	rw := core.NewResponseWriterHandler(recorder)

	data := []byte(`{"key":"value"}`)
	_, err := rw.Write(data)
	assert.NoError(t, err)

	rw.Finalize()

	receivedData := recorder.Body.Bytes()
	assert.NotEmpty(t, receivedData, "Buffered data should be written to the ResponseWriter")
}
