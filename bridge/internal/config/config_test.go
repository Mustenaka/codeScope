package config

import (
	"testing"
	"time"
)

func TestLoadProvidesDefaults(t *testing.T) {
	t.Setenv("CODESCOPE_BRIDGE_AGENT_NAME", "")
	t.Setenv("CODESCOPE_BRIDGE_SERVER_URL", "")
	t.Setenv("CODESCOPE_BRIDGE_WORKSPACE_ROOT", "")
	t.Setenv("CODESCOPE_BRIDGE_MACHINE_ID", "")
	t.Setenv("CODESCOPE_BRIDGE_SESSION_ID", "")
	t.Setenv("CODESCOPE_BRIDGE_CAPTURE_MODE", "")
	t.Setenv("CODESCOPE_BRIDGE_DISCOVERY_INTERVAL", "")
	t.Setenv("CODESCOPE_BRIDGE_SESSION_HEARTBEAT_INTERVAL", "")
	t.Setenv("CODESCOPE_BRIDGE_SESSION_STABILITY_WINDOW", "")
	t.Setenv("CODESCOPE_BRIDGE_PROMPT_FILE", "")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_COMMAND", "")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_ARGS", "")
	t.Setenv("CODESCOPE_BRIDGE_PROMPT_STATE_FILE", "")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_RESTART_MAX", "")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_RESTART_DELAY", "")
	t.Setenv("CODESCOPE_BRIDGE_SOURCE_MODE", "")
	t.Setenv("CODESCOPE_BRIDGE_SOURCE_FILE", "")

	cfg := Load()

	if cfg.AgentName != "bridge" {
		t.Fatalf("expected default agent name, got %q", cfg.AgentName)
	}

	if cfg.ServerURL != "ws://localhost:8080/ws/bridge" {
		t.Fatalf("expected default server url, got %q", cfg.ServerURL)
	}

	if cfg.WorkspaceRoot != "." {
		t.Fatalf("expected default workspace root, got %q", cfg.WorkspaceRoot)
	}

	if cfg.MachineID == "" {
		t.Fatal("expected generated machine id")
	}

	if cfg.SessionID == "" {
		t.Fatal("expected generated session id")
	}

	if cfg.CaptureMode != "discovery" {
		t.Fatalf("expected default capture mode discovery, got %q", cfg.CaptureMode)
	}

	if cfg.DiscoveryInterval != 5*time.Second {
		t.Fatalf("expected default discovery interval 5s, got %s", cfg.DiscoveryInterval)
	}

	if cfg.SessionHeartbeatInterval != 15*time.Second {
		t.Fatalf("expected default session heartbeat interval 15s, got %s", cfg.SessionHeartbeatInterval)
	}

	if cfg.SessionStabilityWindow != 10*time.Second {
		t.Fatalf("expected default session stability window 10s, got %s", cfg.SessionStabilityWindow)
	}

	if cfg.PromptFile != "" {
		t.Fatalf("expected empty default prompt file override, got %q", cfg.PromptFile)
	}

	if cfg.PromptStateFile != "" {
		t.Fatalf("expected empty default prompt state file override, got %q", cfg.PromptStateFile)
	}

	if cfg.ManagedCommand != "" {
		t.Fatalf("expected empty managed command, got %q", cfg.ManagedCommand)
	}

	if len(cfg.ManagedArgs) != 0 {
		t.Fatalf("expected empty managed args, got %#v", cfg.ManagedArgs)
	}

	if cfg.ManagedRestartMax != 3 {
		t.Fatalf("expected default managed restart max 3, got %d", cfg.ManagedRestartMax)
	}

	if cfg.ManagedRestartDelay != time.Second {
		t.Fatalf("expected default managed restart delay 1s, got %s", cfg.ManagedRestartDelay)
	}

	if cfg.SourceMode != "reader" {
		t.Fatalf("expected default source mode, got %q", cfg.SourceMode)
	}

	if cfg.SourceFile != "" {
		t.Fatalf("expected empty source file, got %q", cfg.SourceFile)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("CODESCOPE_BRIDGE_AGENT_NAME", "codex")
	t.Setenv("CODESCOPE_BRIDGE_SERVER_URL", "ws://example.com/agent")
	t.Setenv("CODESCOPE_BRIDGE_WORKSPACE_ROOT", "D:/workspace")
	t.Setenv("CODESCOPE_BRIDGE_MACHINE_ID", "machine-1")
	t.Setenv("CODESCOPE_BRIDGE_SESSION_ID", "session-1")
	t.Setenv("CODESCOPE_BRIDGE_CAPTURE_MODE", "reader")
	t.Setenv("CODESCOPE_BRIDGE_DISCOVERY_INTERVAL", "3s")
	t.Setenv("CODESCOPE_BRIDGE_SESSION_HEARTBEAT_INTERVAL", "9s")
	t.Setenv("CODESCOPE_BRIDGE_SESSION_STABILITY_WINDOW", "12s")
	t.Setenv("CODESCOPE_BRIDGE_PROMPT_FILE", "D:/workspace/.codescope/prompts/inbox.jsonl")
	t.Setenv("CODESCOPE_BRIDGE_PROMPT_STATE_FILE", "D:/workspace/.codescope/prompts/inbox.state.json")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_COMMAND", "codex")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_ARGS", "run --stdio")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_RESTART_MAX", "5")
	t.Setenv("CODESCOPE_BRIDGE_MANAGED_RESTART_DELAY", "250ms")
	t.Setenv("CODESCOPE_BRIDGE_SOURCE_MODE", "jsonl")
	t.Setenv("CODESCOPE_BRIDGE_SOURCE_FILE", "D:/workspace/.codescope/events.jsonl")

	cfg := Load()

	if cfg.AgentName != "codex" {
		t.Fatalf("expected overridden agent name, got %q", cfg.AgentName)
	}

	if cfg.ServerURL != "ws://example.com/agent" {
		t.Fatalf("expected overridden server url, got %q", cfg.ServerURL)
	}

	if cfg.WorkspaceRoot != "D:/workspace" {
		t.Fatalf("expected overridden workspace root, got %q", cfg.WorkspaceRoot)
	}

	if cfg.MachineID != "machine-1" {
		t.Fatalf("expected overridden machine id, got %q", cfg.MachineID)
	}

	if cfg.SessionID != "session-1" {
		t.Fatalf("expected overridden session id, got %q", cfg.SessionID)
	}

	if cfg.CaptureMode != "reader" {
		t.Fatalf("expected capture mode override, got %q", cfg.CaptureMode)
	}

	if cfg.DiscoveryInterval != 3*time.Second {
		t.Fatalf("expected discovery interval override, got %s", cfg.DiscoveryInterval)
	}

	if cfg.SessionHeartbeatInterval != 9*time.Second {
		t.Fatalf("expected session heartbeat interval override, got %s", cfg.SessionHeartbeatInterval)
	}

	if cfg.SessionStabilityWindow != 12*time.Second {
		t.Fatalf("expected session stability window override, got %s", cfg.SessionStabilityWindow)
	}

	if cfg.PromptFile != "D:/workspace/.codescope/prompts/inbox.jsonl" {
		t.Fatalf("expected overridden prompt file, got %q", cfg.PromptFile)
	}

	if cfg.PromptStateFile != "D:/workspace/.codescope/prompts/inbox.state.json" {
		t.Fatalf("expected overridden prompt state file, got %q", cfg.PromptStateFile)
	}

	if cfg.ManagedCommand != "codex" {
		t.Fatalf("expected managed command, got %q", cfg.ManagedCommand)
	}

	if len(cfg.ManagedArgs) != 2 || cfg.ManagedArgs[0] != "run" || cfg.ManagedArgs[1] != "--stdio" {
		t.Fatalf("expected managed args to be parsed, got %#v", cfg.ManagedArgs)
	}

	if cfg.ManagedRestartMax != 5 {
		t.Fatalf("expected managed restart max 5, got %d", cfg.ManagedRestartMax)
	}

	if cfg.ManagedRestartDelay != 250*time.Millisecond {
		t.Fatalf("expected managed restart delay 250ms, got %s", cfg.ManagedRestartDelay)
	}

	if cfg.SourceMode != "jsonl" {
		t.Fatalf("expected source mode override, got %q", cfg.SourceMode)
	}

	if cfg.SourceFile != "D:/workspace/.codescope/events.jsonl" {
		t.Fatalf("expected source file override, got %q", cfg.SourceFile)
	}
}
