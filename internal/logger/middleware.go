package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/domurdoc/gophermart/internal/httputil"
)

type requestData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	requestData *requestData
}

func (w loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.requestData.size += n
	return n, err
}

func (w loggingResponseWriter) WriteHeader(statusCode int) {
	w.requestData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func NewRequestLogger(log *zap.SugaredLogger) httputil.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method
			requestData := requestData{}
			w = loggingResponseWriter{w, &requestData}
			h.ServeHTTP(w, r)
			duration := time.Since(start)
			log.Infow(
				"request",
				"uri", uri,
				"method", method,
				"duration", duration,
				"status", requestData.status,
				"size", requestData.size,
			)
		})
	}
}
