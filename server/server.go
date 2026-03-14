package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chariotplatform/goapi/config"
	"github.com/chariotplatform/goapi/logger"
	"github.com/chariotplatform/goapi/middleware"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	HttpServer *http.Server
	RootRouter *mux.Router
	ApiRouter  *mux.Router
	Started    chan struct{}
}

func NewServer(config *config.AppConfig, log logger.Log) (*Server, func()) {
	// Router
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()

	// Routes
	router.HandleFunc("/status", middleware.StatusHandler(config.StartedAt, log)).Methods("GET")
	router.Handle("/metrics", promhttp.Handler())

	// Middlewares
	middleware.PrometheusInit()
	logOpts := middleware.DefaultLoggerOptions()
	router.Use(middleware.HTTPLoggerMiddleware(log, logOpts))
	router.Use(middleware.TrackMetricsMiddleware([]string{"/status", "/metrics"}))

	// CORS
	headersOk := handlers.AllowedHeaders([]string{
		"X-Requested-With",
		"Content-Type",
		"Authorization"})
	originsOk := handlers.AllowedOrigins(config.HTTP.CORSAllowedOrigins)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "PUT", "OPTIONS"})
	credentialsOk := handlers.AllowCredentials()
	corsHandler := handlers.CORS(originsOk, headersOk, methodsOk, credentialsOk)(router)

	// Server
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.HTTP.Host, config.HTTP.Port),
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	shutdown := func() {
		log.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v\n", err)
		}
		log.Info("Server gracefully stopped.")
	}

	return &Server{
		HttpServer: &server,
		RootRouter: router,
		ApiRouter:  apiRouter,
		Started:    make(chan struct{}),
	}, shutdown
}

func StartServer(server *Server, logger logger.Log) {
	go func() {
		logger.Infof("Starting server on %s", server.HttpServer.Addr)

		// notify main thread
		close(server.Started)

		if err := server.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Error starting server on %v\n", err)
		}
	}()
}
