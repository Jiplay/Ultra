package middleware

import (
	"log"
	"net/http"
	"time"
)

const (
	colorBlue  = "\033[34m"
	colorReset = "\033[0m"
)

// LoggingMiddleware logs all incoming HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the request in blue
		if r.URL.RawQuery != "" {
			log.Printf("%s[%s] %s?%s%s", colorBlue, r.Method, r.URL.Path, r.URL.RawQuery, colorReset)
		} else {
			log.Printf("%s[%s] %s%s", colorBlue, r.Method, r.URL.Path, colorReset)
		}

		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Log completion with status code and duration in blue
		duration := time.Since(start)
		log.Printf("%s[%s] %s - %d (%v)%s", colorBlue, r.Method, r.URL.Path, lrw.statusCode, duration, colorReset)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
