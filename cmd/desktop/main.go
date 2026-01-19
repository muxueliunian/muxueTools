// Package main is the entry point for MuxueTools desktop application.
// It embeds the web frontend in a WebView window for native-like experience.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"muxueTools/internal/api"
	"muxueTools/internal/config"

	"github.com/sirupsen/logrus"
	webview "github.com/webview/webview_go"
)

// Version information, set at build time via ldflags.
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		debug      = flag.Bool("debug", false, "Enable WebView debug mode (DevTools)")
		devMode    = flag.Bool("dev", false, "Development mode: connect to Vite dev server (default: http://localhost:5173)")
		devURL     = flag.String("dev-url", "http://localhost:5173", "Custom dev server URL (only used with -dev)")
		webRoot    = flag.String("webroot", "", "Path to web frontend files (default: auto-detect web/dist)")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Handle version flag
	if *showVer {
		fmt.Printf("MuxueTools Desktop %s\n", Version)
		fmt.Printf("  Build Time: %s\n", BuildTime)
		fmt.Printf("  Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logger.WithFields(logrus.Fields{
		"version": Version,
		"mode":    "desktop",
	}).Info("MuxueTools Desktop starting")

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

	// For desktop mode, use a random available port to avoid conflicts
	// Override the configured port with 0 to let OS assign a free port
	originalPort := cfg.Server.Port
	cfg.Server.Port = 0

	logger.WithFields(logrus.Fields{
		"original_port": originalPort,
		"keys":          len(cfg.Keys),
		"strategy":      cfg.Pool.Strategy,
	}).Info("Configuration loaded (using random port for desktop mode)")

	// Determine WebRoot path for static file serving (if not in dev mode)
	var resolvedWebRoot string
	if !*devMode {
		if *webRoot != "" {
			resolvedWebRoot = *webRoot
		} else {
			// Auto-detect: try relative path first (for development), then exe-relative
			candidates := []string{
				"web/dist",
				"../web/dist",
			}
			// Also try exe-relative path
			if exePath, err := os.Executable(); err == nil {
				exeDir := filepath.Dir(exePath)
				candidates = append(candidates, filepath.Join(exeDir, "web", "dist"))
				candidates = append(candidates, filepath.Join(exeDir, "..", "web", "dist"))
			}

			for _, candidate := range candidates {
				if indexPath := filepath.Join(candidate, "index.html"); fileExists(indexPath) {
					resolvedWebRoot = candidate
					logger.WithField("webroot", resolvedWebRoot).Info("Auto-detected web root")
					break
				}
			}
		}

		if resolvedWebRoot == "" && !*devMode {
			logger.Warn("Web root not found, frontend will not be available. Use -dev flag for development mode.")
		}
	}

	// Create server with WebRoot
	serverOpts := []api.ServerOption{
		api.WithVersion(Version),
		api.WithLogger(logger),
	}
	if resolvedWebRoot != "" {
		serverOpts = append(serverOpts, api.WithWebRoot(resolvedWebRoot))
	}

	server, err := api.NewServer(cfg, serverOpts...)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Create a listener to get the actual port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Fatalf("Failed to create listener: %v", err)
	}
	actualPort := listener.Addr().(*net.TCPAddr).Port
	serverAddr := fmt.Sprintf("http://localhost:%d", actualPort)

	// Update config with actual port so /api/config returns the real port
	cfg.Server.Port = actualPort

	logger.WithField("addr", serverAddr).Info("Server will listen on")

	// Create context for graceful shutdown (only cancel is needed)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to signal server errors
	serverErrCh := make(chan error, 1)

	// Start HTTP server in goroutine
	go func() {
		httpServer := &http.Server{
			Handler:      server.Engine(),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 120 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		// Serve using the listener we created
		logger.Info("Starting HTTP server...")
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("HTTP server error")
			serverErrCh <- err
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// Check if server started successfully
	select {
	case err := <-serverErrCh:
		logger.Fatalf("Server failed to start: %v", err)
	default:
		logger.Info("HTTP server started successfully")
	}

	// Create and configure WebView
	w := webview.New(*debug)
	if w == nil {
		logger.Fatal("Failed to create WebView window")
	}
	defer w.Destroy()

	w.SetTitle("MuxueTools")
	w.SetSize(1024, 768, webview.HintNone)

	// Determine navigation URL based on mode
	var navigateURL string
	if *devMode {
		// Development mode: use Vite dev server
		navigateURL = *devURL
		logger.WithField("mode", "development").Info("Using Vite dev server")
	} else {
		// Production mode: use embedded server
		navigateURL = serverAddr
	}

	w.Navigate(navigateURL)

	logger.WithFields(logrus.Fields{
		"url":     navigateURL,
		"api":     serverAddr,
		"width":   1024,
		"height":  768,
		"debug":   *debug,
		"devMode": *devMode,
	}).Info("WebView window created")

	// Run WebView (blocks until window is closed)
	w.Run()

	// Window closed, initiate graceful shutdown
	logger.Info("WebView window closed, shutting down...")

	// Cancel context to signal shutdown
	cancel()

	// Shutdown server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Warn("Server shutdown error")
	}

	// Close storage and other resources
	if err := server.Close(); err != nil {
		logger.WithError(err).Warn("Failed to close server resources")
	}

	logger.Info("MuxueTools Desktop exited")
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
