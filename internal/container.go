package internal

import (
	"github.com/chariotplatform/goapi/cache"
	"github.com/chariotplatform/goapi/config"
	"github.com/chariotplatform/goapi/logger"
	"github.com/chariotplatform/goapi/server"
	"github.com/chariotplatform/goapi/store"
)

func Init(config *config.AppConfig, log logger.Log, s *server.Server, db store.Database, cache cache.Cache) {
	// Repositories

	// Services

	// Handlers

	// Register routes

	log.Info("Ready to serve requests.")
}
