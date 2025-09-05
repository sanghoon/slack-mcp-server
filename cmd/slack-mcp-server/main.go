package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/korotovsky/slack-mcp-server/pkg/config"
	"github.com/korotovsky/slack-mcp-server/pkg/provider"
	"github.com/korotovsky/slack-mcp-server/pkg/server"
	"github.com/mattn/go-isatty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultSseHost = "127.0.0.1"
var defaultSsePort = 13080

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or sse)")
	flag.Parse()

	logger, err := newLogger(transport)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	err = validateToolConfig(os.Getenv("SLACK_MCP_ADD_MESSAGE_TOOL"))
	if err != nil {
		logger.Fatal("error in SLACK_MCP_ADD_MESSAGE_TOOL",
			zap.String("context", "console"),
			zap.Error(err),
		)
	}

	p := provider.New(transport, logger)
	s := server.NewMCPServer(p, logger)

	// Start cache initialization in background
	go func() {
		initUsersCache(p, logger)
		initChannelsCache(p, logger)

		if ready, _ := p.IsReady(); ready {
			logger.Info("Slack MCP Server caches are ready",
				zap.String("context", "console"),
			)
		}

		// Start cache refresh ticker after initial load
		startCacheRefreshTicker(p, logger)
	}()

	switch transport {
	case "stdio":
		if err := s.ServeStdio(); err != nil {
			logger.Fatal("Server error",
				zap.String("context", "console"),
				zap.Error(err),
			)
		}
	case "sse":
		host := os.Getenv("SLACK_MCP_HOST")
		if host == "" {
			host = defaultSseHost
		}
		port := os.Getenv("SLACK_MCP_PORT")
		if port == "" {
			port = strconv.Itoa(defaultSsePort)
		}

		sseServer := s.ServeSSE(":" + port)
		logger.Info(
			fmt.Sprintf("SSE server listening on %s", fmt.Sprintf("%s:%s/sse", host, port)),
			zap.String("context", "console"),
			zap.String("host", host),
			zap.String("port", port),
		)

		if ready, _ := p.IsReady(); !ready {
			logger.Info("Slack MCP Server is still warming up caches",
				zap.String("context", "console"),
			)
		}

		if err := sseServer.Start(host + ":" + port); err != nil {
			logger.Fatal("Server error",
				zap.String("context", "console"),
				zap.Error(err),
			)
		}
	default:
		logger.Fatal("Invalid transport type",
			zap.String("context", "console"),
			zap.String("transport", transport),
			zap.String("allowed", "stdio,sse"),
		)
	}
}

func initUsersCache(p *provider.ApiProvider, logger *zap.Logger) {
	logger.Info("Caching users collection...",
		zap.String("context", "console"),
	)

	if config.IsDemoMode() {
		logger.Info("Demo credentials are set, skip",
			zap.String("context", "console"),
		)
		return
	}

	err := p.RefreshUsers(context.Background())
	if err != nil {
		logger.Error("Error refreshing users cache",
			zap.String("context", "console"),
			zap.Error(err),
		)
		// Don't fatal here, let the server continue
	} else {
		logger.Info("Users cache initialized successfully",
			zap.String("context", "console"),
		)
	}
}

func initChannelsCache(p *provider.ApiProvider, logger *zap.Logger) {
	logger.Info("Caching channels collection...",
		zap.String("context", "console"),
	)

	if os.Getenv("SLACK_MCP_XOXP_TOKEN") == "demo" || os.Getenv("SLACK_MCP_XOXB_TOKEN") == "demo" || (os.Getenv("SLACK_MCP_XOXC_TOKEN") == "demo" && os.Getenv("SLACK_MCP_XOXD_TOKEN") == "demo") {
		logger.Info("Demo credentials are set, skip.",
			zap.String("context", "console"),
		)
		return
	}

	err := p.RefreshChannels(context.Background())
	if err != nil {
		logger.Error("Error refreshing channels cache",
			zap.String("context", "console"),
			zap.Error(err),
		)
		// Don't fatal here, let the server continue
	} else {
		logger.Info("Channels cache initialized successfully",
			zap.String("context", "console"),
		)
	}
}

func validateToolConfig(config string) error {
	if config == "" || config == "true" || config == "1" {
		return nil
	}

	items := strings.Split(config, ",")
	hasNegated := false
	hasPositive := false

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.HasPrefix(item, "!") {
			hasNegated = true
		} else {
			hasPositive = true
		}
	}

	if hasNegated && hasPositive {
		return fmt.Errorf("cannot mix allowed and disallowed (! prefixed) channels")
	}

	return nil
}

func newLogger(transport string) (*zap.Logger, error) {
	atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)
	if envLevel := os.Getenv("SLACK_MCP_LOG_LEVEL"); envLevel != "" {
		if err := atomicLevel.UnmarshalText([]byte(envLevel)); err != nil {
			fmt.Printf("Invalid log level '%s': %v, using 'info'\n", envLevel, err)
		}
	}

	useJSON := shouldUseJSONFormat()
	useColors := shouldUseColors() && !useJSON

	outputPath := "stdout"
	if transport == "stdio" {
		outputPath = "stderr"
	}

	var config zap.Config

	if useJSON {
		config = zap.Config{
			Level:            atomicLevel,
			Development:      false,
			Encoding:         "json",
			OutputPaths:      []string{outputPath},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:       "timestamp",
				LevelKey:      "level",
				NameKey:       "logger",
				MessageKey:    "message",
				StacktraceKey: "stacktrace",
				EncodeLevel:   zapcore.LowercaseLevelEncoder,
				EncodeTime:    zapcore.RFC3339TimeEncoder,
				EncodeCaller:  zapcore.ShortCallerEncoder,
			},
		}
	} else {
		config = zap.Config{
			Level:            atomicLevel,
			Development:      true,
			Encoding:         "console",
			OutputPaths:      []string{outputPath},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:          "timestamp",
				LevelKey:         "level",
				NameKey:          "logger",
				MessageKey:       "msg",
				StacktraceKey:    "stacktrace",
				EncodeLevel:      getConsoleLevelEncoder(useColors),
				EncodeTime:       zapcore.ISO8601TimeEncoder,
				EncodeCaller:     zapcore.ShortCallerEncoder,
				ConsoleSeparator: " | ",
			},
		}
	}

	logger, err := config.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}

	logger = logger.With(zap.String("app", "slack-mcp-server"))

	return logger, err
}

// shouldUseJSONFormat determines if JSON format should be used
func shouldUseJSONFormat() bool {
	if format := os.Getenv("SLACK_MCP_LOG_FORMAT"); format != "" {
		return strings.ToLower(format) == "json"
	}

	if env := os.Getenv("ENVIRONMENT"); env != "" {
		switch strings.ToLower(env) {
		case "production", "prod", "staging":
			return true
		case "development", "dev", "local":
			return false
		}
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" ||
		os.Getenv("DOCKER_CONTAINER") != "" ||
		os.Getenv("container") != "" {
		return true
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return true
	}

	return false
}

func shouldUseColors() bool {
	if colorEnv := os.Getenv("SLACK_MCP_LOG_COLOR"); colorEnv != "" {
		return colorEnv == "true" || colorEnv == "1"
	}

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	if env := os.Getenv("ENVIRONMENT"); env == "development" || env == "dev" {
		return isatty.IsTerminal(os.Stdout.Fd())
	}

	return isatty.IsTerminal(os.Stdout.Fd())
}

func getConsoleLevelEncoder(useColors bool) zapcore.LevelEncoder {
	if useColors {
		return zapcore.CapitalColorLevelEncoder
	}
	return zapcore.CapitalLevelEncoder
}

func startCacheRefreshTicker(p *provider.ApiProvider, logger *zap.Logger) {
	// Get cache refresh interval from config
	interval := config.GetCacheRefreshInterval()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Cache refresh ticker started",
		zap.String("context", "console"),
		zap.Duration("interval", interval),
	)

	for range ticker.C {
		// Skip if demo mode
		if config.IsDemoMode() {
			continue
		}

		logger.Info("Refreshing caches from API",
			zap.String("context", "console"),
		)

		ctx := context.Background()

		// Delete users cache file to force refresh from API
		usersCache := config.GetUsersCache()
		if err := os.Remove(usersCache); err != nil && !os.IsNotExist(err) {
			logger.Warn("Failed to remove users cache file",
				zap.String("cache_file", usersCache),
				zap.Error(err),
			)
		}

		// Refresh users from API
		if err := p.RefreshUsers(ctx); err != nil {
			logger.Error("Failed to refresh users cache",
				zap.String("context", "console"),
				zap.Error(err),
			)
		}

		// Delete channels cache file to force refresh from API
		channelsCache := config.GetChannelsCache()
		if err := os.Remove(channelsCache); err != nil && !os.IsNotExist(err) {
			logger.Warn("Failed to remove channels cache file",
				zap.String("cache_file", channelsCache),
				zap.Error(err),
			)
		}

		// Refresh channels from API
		if err := p.RefreshChannels(ctx); err != nil {
			logger.Error("Failed to refresh channels cache",
				zap.String("context", "console"),
				zap.Error(err),
			)
		}
	}
}
