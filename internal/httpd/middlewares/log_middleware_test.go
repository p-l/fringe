package middlewares_test

import (
	"bufio"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/stretchr/testify/assert"
)

func mockLogger(t *testing.T) (mockLogger *log.Logger, logReader *os.File, logWriter *os.File) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		assert.Fail(t, "couldn't get os Pipe: %v", err)
	}

	mockLog := log.New(writer, "", log.LstdFlags)

	return mockLog, reader, writer
}

func closeMockLogger(t *testing.T, reader *os.File, writer *os.File) {
	t.Helper()

	if err := reader.Close(); err != nil {
		t.Log("error closing reader was ", err)
	}

	if err := writer.Close(); err != nil {
		t.Log("error closing writer was ", err)
	}
}

func readLines(t *testing.T, reader *os.File, lineCount int) []string {
	t.Helper()

	var output []string

	scanner := bufio.NewScanner(reader)

	for i := 0; i < lineCount; i++ {
		scanner.Scan()
		line := scanner.Text()
		output = append(output, line)
	}

	return output
}

func TestLogMiddleware_LogRequests(t *testing.T) {
	t.Parallel()

	t.Run("Log the HTTP Method and Path", func(t *testing.T) {
		t.Parallel()

		logger, reader, writer := mockLogger(t)
		defer closeMockLogger(t, reader, writer)

		logMiddleware := middlewares.NewLogMiddleware(logger)

		req := httptest.NewRequest(http.MethodGet, "/test/path/", nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(logMiddleware.LogRequests)
		router.HandleFunc("/test/path/", func(writer http.ResponseWriter, request *http.Request) {})
		router.ServeHTTP(res, req)

		lines := readLines(t, reader, 1)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		// Log Method
		assert.Contains(t, lines[0], "GET")
		assert.Contains(t, lines[0], "/test/path/")
	})

	t.Run("Redact password query param", func(t *testing.T) {
		t.Parallel()

		logger, reader, writer := mockLogger(t)
		defer closeMockLogger(t, reader, writer)

		logMiddleware := middlewares.NewLogMiddleware(logger)

		req := httptest.NewRequest(http.MethodGet, "/test/?password=secret&code=secret&token=secret&secret=super%20secret", nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(logMiddleware.LogRequests)
		router.HandleFunc("/test/", func(writer http.ResponseWriter, request *http.Request) {})
		router.ServeHTTP(res, req)

		lines := readLines(t, reader, 1)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		// Log Method
		assert.Contains(t, lines[0], "GET")
		assert.Contains(t, lines[0], "/test/")
		assert.Contains(t, lines[0], "password=redacted")
		assert.Contains(t, lines[0], "code=redacted")
		assert.Contains(t, lines[0], "token=redacted")
		assert.Contains(t, lines[0], "secret=redacted")
	})
}
