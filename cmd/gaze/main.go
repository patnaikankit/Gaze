package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/patnaikankit/Gaze/internal/builder"
	"github.com/patnaikankit/Gaze/internal/config"
	"github.com/patnaikankit/Gaze/internal/runner"
	"github.com/patnaikankit/Gaze/internal/watcher"
	"github.com/patnaikankit/Gaze/pkg/logger"
)

func main() {
	printBanner()

	// Load config
	cfg, err := config.ParseConfig() // adjust path if needed
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		return
	}

	// Initialize runner and builder
	appRunner := runner.New(cfg)
	appBuilder := builder.New(cfg)

	// Setup channels
	eventChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start file watcher
	fileWatcher, err := watcher.NewWatcher(eventChan, cfg)
	if err != nil {
		logger.Error("Failed to create file watcher: %v", err)
	}
	if err := fileWatcher.Start(); err != nil {
		logger.Error("Failed to start file watcher: %v", err)
	}

	// Initial build
	logger.Warn("Performing initial build...")
	buildResult := appBuilder.Build()
	startServer(buildResult, appRunner)

	for {
		select {
		case <-eventChan:
			logger.Info("🔄 File changes detected, rebuilding...")
			buildResult := appBuilder.Build()
			startServer(buildResult, appRunner)

		case sig := <-signalChan:
			logger.Warn("⛔ Received signal %v, shutting down...", sig)
			appRunner.Stop()
			fileWatcher.Stop() // instead of DoneChannel
			return
		}
	}
}

func startServer(buildResult builder.BuildResult, appRunner *runner.Runner) {
	if buildResult.Success {
		appRunner.Stop()
		if err := appRunner.Run(); err != nil {
			logger.Error("Failed to restart application: %v", err)
		}
	} else {
		logger.Error("Build failed:\n%s", buildResult.Output)
	}
}

func printBanner() {
	banner := `
   ____                     
  / ___| __ _ _______   
 | |  _ / _' |_  / _ \
 | |_| | (_| |/ /  __/
  \____|\__,_/___\___|  
                             
`
	fmt.Println(banner)
	logger.Info("👀 Gaze — Hot reload for Go applications")
	logger.Info("----------------------------------------")
}
