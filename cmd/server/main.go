// Package main is the entry point for the MuxueTools server.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"muxueTools/internal/api"
	"muxueTools/internal/config"

	"github.com/sirupsen/logrus"
)

// Version information, set at build time via ldflags.
var (
	Version   = "0.3.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		showHelp   = flag.Bool("help", false, "Show help message")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Handle flags
	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	if *showVer {
		printVersion()
		os.Exit(0)
	}

	// Initialize logger for startup
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Print banner
	printBanner(logger)

	// Set application version for update checks
	config.SetVersion(Version)

	// Load configuration
	var err error
	if *configPath != "" {
		err = config.InitFromFile(*configPath)
	} else {
		err = config.Init()
	}

	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := config.Get()
	logger.WithFields(logrus.Fields{
		"port":     cfg.Server.Port,
		"keys":     len(cfg.Keys),
		"strategy": cfg.Pool.Strategy,
	}).Info("Configuration loaded")

	// Check if we have any keys
	if len(cfg.Keys) == 0 {
		logger.Warn("No API keys configured. Add keys to config.yaml or via /api/keys endpoint.")
	}

	// Create and start server
	server, err := api.NewServer(
		cfg,
		api.WithVersion(Version),
		api.WithLogger(logger),
	)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Start server in goroutine
	go func() {
		if err := server.Run(); err != nil {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	logger.WithFields(logrus.Fields{
		"addr":    server.Addr(),
		"version": Version,
	}).Info("Server started successfully")

	// Print access information
	fmt.Printf("\n  ✓ Server is running at http://%s\n", server.Addr())
	fmt.Printf("  ✓ OpenAI API endpoint: http://%s/v1/chat/completions\n", server.Addr())
	fmt.Printf("  ✓ Models list: http://%s/v1/models\n", server.Addr())
	fmt.Printf("  ✓ Health check: http://%s/health\n", server.Addr())
	fmt.Printf("  ✓ Admin API: http://%s/api/keys\n\n", server.Addr())

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received, shutting down gracefully...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
		os.Exit(1)
	}

	logger.Info("Server exited properly")
}

// printBanner prints the application banner.
func printBanner(logger *logrus.Logger) {
	banner := `
    __  ___      __      ___    ____  ____
   /  |/  /_  __/ /___  /   |  / __ \/  _/
  / /|_/ / / / / / __ \/ /| | / /_/ // /  
 / /  / / /_/ / / / / / ___ |/ ____// /   
/_/  /_/\__,_/_/_/ /_/_/  |_/_/   /___/   
                                          
  Gemini to OpenAI API Proxy
`
	fmt.Println(banner)
	logger.WithFields(logrus.Fields{
		"version":    Version,
		"build_time": BuildTime,
		"commit":     GitCommit,
	}).Info("MuxueTools starting")
}

// printVersion prints version information.
func printVersion() {
	fmt.Printf("MuxueTools %s\n", Version)
	fmt.Printf("  Build Time: %s\n", BuildTime)
	fmt.Printf("  Git Commit: %s\n", GitCommit)
}

// printUsage prints usage information.
func printUsage() {
	fmt.Println("MuxueTools - Gemini to OpenAI API Proxy")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  MuxueTools [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  MuxueTools                     Start with default config")
	fmt.Println("  MuxueTools -config ./my.yaml   Start with custom config file")
	fmt.Println("  MuxueTools -version            Show version information")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  MXLN_SERVER_PORT            Override server port")
	fmt.Println("  MXLN_POOL_STRATEGY          Override pool strategy")
	fmt.Println("  MXLN_LOGGING_LEVEL          Override log level")
}
