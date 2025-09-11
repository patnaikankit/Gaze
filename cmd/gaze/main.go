package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/patnaikankit/Gaze/internal/config"
	"github.com/patnaikankit/Gaze/pkg/logger"
)

func main() {
	logger.SetLevel(logger.DEBUG)

	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Error("Failed to load configuration: %v", err)
	}

	logger.Info("Gaze is starting up...")
	logger.Debug("Using watch dir: %s", cfg.WatchDir)
	logger.Debug("Build command: %s", cfg.BuildCMD)
	logger.Debug("Run command: %s", cfg.RunCMD)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Gaze is shutting down...")
}
