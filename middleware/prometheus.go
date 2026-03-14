package middleware

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goapi_requests_total",
			Help: "Total number of requests processed by the goapi api server.",
		},
		[]string{"path", "status"},
	)

	ErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goapi_requests_errors_total",
			Help: "Total number of error requests processed by the goapi api server.",
		},
		[]string{"path", "status"},
	)
)

func PrometheusInit() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(ErrorCount)
}

func TrackMetricsMiddleware(ignoredPaths []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip internal endpoints
			if slices.Contains(ignoredPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Capture status
			rw, ok := w.(*responseWriter)
			if !ok {
				// Should not happen if HTTPLoggerMiddleware is applied first,
				// but let's be safe.
				rw = &responseWriter{w, http.StatusOK}
				w = rw
			}

			defer func() {
				path := r.URL.Path
				if route := mux.CurrentRoute(r); route != nil {
					if tpl, err := route.GetPathTemplate(); err == nil {
						path = tpl
					}
				}

				statusStr := strconv.Itoa(rw.status)
				RequestCount.WithLabelValues(path, statusStr).Inc()
				if rw.status >= 400 {
					ErrorCount.WithLabelValues(path, statusStr).Inc()
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
