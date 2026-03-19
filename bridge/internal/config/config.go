package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AgentName                string
	ServerURL                string
	WorkspaceRoot            string
	MachineID                string
	SessionID                string
	CaptureMode              string
	DiscoveryInterval        time.Duration
	SessionHeartbeatInterval time.Duration
	SessionStabilityWindow   time.Duration
	SourceMode               string
	SourceFile               string
	PromptFile               string
	PromptStateFile          string
	ManagedCommand           string
	ManagedArgs              []string
	ManagedRestartMax        int
	ManagedRestartDelay      time.Duration
}

func Load() Config {
	return Config{
		AgentName:                envOrDefault("CODESCOPE_BRIDGE_AGENT_NAME", "bridge"),
		ServerURL:                envOrDefault("CODESCOPE_BRIDGE_SERVER_URL", "ws://localhost:8080/ws/bridge"),
		WorkspaceRoot:            envOrDefault("CODESCOPE_BRIDGE_WORKSPACE_ROOT", "."),
		MachineID:                envOrDefault("CODESCOPE_BRIDGE_MACHINE_ID", defaultMachineID()),
		SessionID:                envOrDefault("CODESCOPE_BRIDGE_SESSION_ID", defaultSessionID()),
		CaptureMode:              envOrDefault("CODESCOPE_BRIDGE_CAPTURE_MODE", "discovery"),
		DiscoveryInterval:        parseDurationOrDefault("CODESCOPE_BRIDGE_DISCOVERY_INTERVAL", 5*time.Second),
		SessionHeartbeatInterval: parseDurationOrDefault("CODESCOPE_BRIDGE_SESSION_HEARTBEAT_INTERVAL", 15*time.Second),
		SessionStabilityWindow:   parseDurationOrDefault("CODESCOPE_BRIDGE_SESSION_STABILITY_WINDOW", 10*time.Second),
		SourceMode:               envOrDefault("CODESCOPE_BRIDGE_SOURCE_MODE", "reader"),
		SourceFile:               envOrDefault("CODESCOPE_BRIDGE_SOURCE_FILE", ""),
		PromptFile:               envOrDefault("CODESCOPE_BRIDGE_PROMPT_FILE", ""),
		PromptStateFile:          envOrDefault("CODESCOPE_BRIDGE_PROMPT_STATE_FILE", ""),
		ManagedCommand:           envOrDefault("CODESCOPE_BRIDGE_MANAGED_COMMAND", ""),
		ManagedArgs:              splitArgs(os.Getenv("CODESCOPE_BRIDGE_MANAGED_ARGS")),
		ManagedRestartMax:        parseIntOrDefault("CODESCOPE_BRIDGE_MANAGED_RESTART_MAX", 3),
		ManagedRestartDelay:      parseDurationOrDefault("CODESCOPE_BRIDGE_MANAGED_RESTART_DELAY", time.Second),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func defaultMachineID() string {
	hostname, err := os.Hostname()
	if err == nil && strings.TrimSpace(hostname) != "" {
		return fmt.Sprintf("machine-%s", sanitizeToken(hostname))
	}
	return "machine-" + randomToken()
}

func defaultSessionID() string {
	return "session-" + randomToken()
}

func sanitizeToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "-")
	if value == "" {
		return randomToken()
	}
	return value
}

func randomToken() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}

func splitArgs(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return strings.Fields(value)
}

func parseIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid integer for %s=%q, using default=%d", key, value, fallback)
		return fallback
	}
	return parsed
}

func parseDurationOrDefault(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid duration for %s=%q, using default=%s", key, value, fallback)
		return fallback
	}
	return parsed
}
