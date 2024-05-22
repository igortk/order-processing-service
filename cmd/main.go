package main

import (
	log "github.com/sirupsen/logrus"
	"order-processing-service/config"
	"order-processing-service/services"
	"order-processing-service/util"
)

func main() {
	cfg := config.GetConfig()

	initLogs(&cfg.LoggerConfig)

	server := initServer(cfg)

	server.Run()
}
func initServer(cfg *config.Config) *services.Server {
	server := services.NewServer(cfg)
	server.InitHandlers()
	server.InitConsumers()

	return server
}

func initLogs(cfg *config.LoggerConfig) {
	logLvl, err := log.ParseLevel(cfg.Level)
	util.IsError(err, config.ErrParseLog)
	log.SetLevel(logLvl)
}
