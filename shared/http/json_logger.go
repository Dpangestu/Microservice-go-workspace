package http

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type logRec struct {
	TS        string `json:"ts"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latencyMs"`
	ClientIP  string `json:"clientIp,omitempty"`
	ReqID     string `json:"reqId"`
	Svc       string `json:"service"`
}

func JSONLogger(next http.Handler) http.Handler {
	svc := os.Getenv("SERVICE_NAME")
	if svc == "" {
		svc = "unknown"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := r.Header.Get("X-Correlation-Id")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ww := &writer{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		rec := logRec{
			TS:        time.Now().Format(time.RFC3339Nano),
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    ww.status,
			LatencyMs: time.Since(start).Milliseconds(),
			ClientIP:  r.Header.Get("X-Forwarded-For"),
			ReqID:     reqID,
			Svc:       svc,
		}
		_ = json.NewEncoder(os.Stdout).Encode(rec)
	})
}

type writer struct {
	http.ResponseWriter
	status int
}

func (w *writer) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
