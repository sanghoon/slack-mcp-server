package config

import (
	"os"
	"strconv"
	"time"
)

const defaultCacheTTLHours = 24

// GetCacheTTL returns the cache TTL from environment variable or default
func GetCacheTTL() time.Duration {
	ttlStr := os.Getenv("SLACK_MCP_CACHE_TTL_HOURS")
	if ttlStr == "" {
		return time.Duration(defaultCacheTTLHours) * time.Hour
	}

	hours, err := strconv.Atoi(ttlStr)
	if err != nil || hours <= 0 {
		return time.Duration(defaultCacheTTLHours) * time.Hour
	}

	return time.Duration(hours) * time.Hour
}

// GetUsersCache returns the users cache file path
func GetUsersCache() string {
	if cache := os.Getenv("SLACK_MCP_USERS_CACHE"); cache != "" {
		return cache
	}
	return ".users_cache.json"
}

// GetChannelsCache returns the channels cache file path
func GetChannelsCache() string {
	if cache := os.Getenv("SLACK_MCP_CHANNELS_CACHE"); cache != "" {
		return cache
	}
	return ".channels_cache.json"
}

// IsDemoMode checks if the application is running in demo mode
func IsDemoMode() bool {
	return os.Getenv("SLACK_MCP_XOXP_TOKEN") == "demo" ||
		os.Getenv("SLACK_MCP_XOXB_TOKEN") == "demo" ||
		(os.Getenv("SLACK_MCP_XOXC_TOKEN") == "demo" && os.Getenv("SLACK_MCP_XOXD_TOKEN") == "demo")
}
