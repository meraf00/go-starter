package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/chariotplatform/goapi/cache"
	"github.com/chariotplatform/goapi/config"
	"github.com/chariotplatform/goapi/internal"
	"github.com/chariotplatform/goapi/logger"
	"github.com/chariotplatform/goapi/server"
	"github.com/chariotplatform/goapi/store"
)

func main() {
	log := logger.NewLogger()
	config := config.LoadConfig(log)

	db, shutdownDB, err := store.NewDatabase(config.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}

	redisCache, shutdownRedis, err := cache.NewRedis(config.Redis, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis: ", err)
	}

	httpServer, shutdownServer := server.NewServer(config, log)

	// Start HTTP Server
	server.StartServer(httpServer, log)

	// wait till server starts
	<-httpServer.Started

	// Initialize routes
	internal.Init(config, log, httpServer, db, redisCache)

	// Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	<-shutdownChan
	shutdownServer()
	shutdownDB()
	shutdownRedis()

	log.Info("Application exited successfully.")
}
