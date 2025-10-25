package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/dyammarcano/crew-das-closures/internal/core"
	"github.com/dyammarcano/crew-das-closures/internal/model"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	responseJSON(w, http.StatusOK, &model.HealthResponse{Status: "ok"})
}

func findServiceHandler(aks *core.Core) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		intentData, err := extractIntentFromRequest(r)
		if err != nil {
			responseJSON(w, http.StatusOK, &model.FindServiceResponse{
				Success: false,
				Error:   fmt.Errorf("invalid request: %w", err).Error(),
			})
			return
		}

		serviceResponse, err := aks.AskQuestion(intentData)
		if err != nil {
			responseJSON(w, http.StatusOK, &model.FindServiceResponse{
				Success: false,
				Error:   fmt.Errorf("internal server error: %w", err).Error(),
			})
			return
		}

		responseJSON(w, http.StatusOK, serviceResponse)
	}
}

func extractIntentFromRequest(r *http.Request) ([]byte, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// statusOKWriter forces any status code to 200 OK
// It preserves headers and body content but overrides the status line.
// This is used to comply with the requirement that all HTTP errors return 200.

type statusOKWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *statusOKWriter) WriteHeader(_ int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(http.StatusOK)
}

func forceStatusOK(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusOKWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
	})
}

var respBufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func responseJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")

	buf := respBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	enc := json.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		buf.Reset()
		respBufPool.Put(buf)
		// Keep compliance: always return 200 with JSON body, even on encoding errors
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":false,"error":"internal server error"}`))
		return
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
	buf.Reset()
	respBufPool.Put(buf)
}
