package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Celedhrim/btcleaner/internal/cleaner"
	"github.com/Celedhrim/btcleaner/internal/config"
	"github.com/Celedhrim/btcleaner/internal/logger"
	"github.com/Celedhrim/btcleaner/internal/server"
	"github.com/Celedhrim/btcleaner/internal/transmission"
)

var (
	// Version is set via ldflags during build
	Version = "dev"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load(Version)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Infof("BTCleaner %s starting...", Version)
	log.Infof("Transmission URL: %s", cfg.Transmission.URL)
	log.Infof("Min free space: %.2f GB", float64(cfg.Cleaner.MinFreeSpace)/(1024*1024*1024))
	log.Infof("Min torrents per tracker: %d", cfg.Cleaner.MinTorrentsPerTracker)
	
	if cfg.DryRun {
		log.Warn("DRY RUN MODE: No torrents will be actually removed")
	}

	// Create Transmission client
	client := transmission.NewClient(
		cfg.Transmission.URL,
		cfg.Transmission.Username,
		cfg.Transmission.Password,
	)

	// Test connection
	log.Info("Testing connection to Transmission...")
	if err := client.TestConnection(); err != nil {
		return fmt.Errorf("failed to connect to Transmission: %w", err)
	}
	log.Info("Successfully connected to Transmission")

	// Create cleaner
	clean := cleaner.New(
		client,
		cfg.Cleaner.MinFreeSpace,
		cfg.Cleaner.MinTorrentsPerTracker,
		cfg.DryRun,
		log,
	)

	// Start web server if enabled
	var webServer *server.Server
	if cfg.Server.Enabled {
		webServer = server.New(cfg.Server.Port, cfg.Server.WebRoot, Version, clean, client, log)
		// Connect logger to server for WebSocket broadcasting
		log.SetCallback(func(entry logger.LogEntry) {
			webServer.BroadcastLog(entry)
		})
		go func() {
			if err := webServer.Start(); err != nil {
				log.Errorf("Web server error: %v", err)
			}
		}()
	}

	// Run in appropriate mode
	if cfg.Daemon.Enabled {
		return runDaemon(clean, cfg, log, webServer)
	}
	
	return runOneShot(clean, log, webServer)
}

func runOneShot(clean *cleaner.Cleaner, log *logger.Logger, webServer *server.Server) error {
	log.Info("Running in one-shot mode")
	
	result, err := clean.Run()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	if result.NeedCleanup {
		log.Infof("Cleanup completed: removed %d torrents (%.2f GB freed)",
			result.RemovedCount,
			float64(result.RemovedSize)/(1024*1024*1024))
	} else {
		log.Info("No cleanup needed")
	}

	// Keep running if web server is enabled
	if webServer != nil {
		log.Info("Web server running, press Ctrl+C to stop")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Info("Shutting down...")
		return webServer.Stop()
	}

	return nil
}

func runDaemon(clean *cleaner.Cleaner, cfg *config.Config, log *logger.Logger, webServer *server.Server) error {
	log.Infof("Running in daemon mode (check interval: %v)", cfg.Daemon.CheckInterval)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker
	ticker := time.NewTicker(cfg.Daemon.CheckInterval)
	defer ticker.Stop()

	// Run immediately on start
	log.Info("Running initial cleanup check...")
	if err := runCleanupCheck(clean, log); err != nil {
		log.Errorf("Initial cleanup check failed: %v", err)
	}

	// Main daemon loop
	for {
		select {
		case <-ticker.C:
			log.Debug("Running periodic cleanup check...")
			if err := runCleanupCheck(clean, log); err != nil {
				log.Errorf("Cleanup check failed: %v", err)
			}

		case sig := <-sigChan:
			log.Infof("Received signal %v, shutting down gracefully...", sig)
			if webServer != nil {
				return webServer.Stop()
			}
			return nil
		}
	}
}

func runCleanupCheck(clean *cleaner.Cleaner, log *logger.Logger) error {
	result, err := clean.Run()
	if err != nil {
		return err
	}

	if result.NeedCleanup && result.RemovedCount > 0 {
		log.Infof("Cleanup completed: removed %d torrents (%.2f GB freed)",
			result.RemovedCount,
			float64(result.RemovedSize)/(1024*1024*1024))
	}

	return nil
}
