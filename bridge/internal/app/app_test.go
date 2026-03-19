package app

import "testing"

func TestNewDefaultsToDiscoveryCollector(t *testing.T) {
	t.Setenv("CODESCOPE_BRIDGE_CAPTURE_MODE", "")

	application := New()

	if application.cfg.CaptureMode != "discovery" {
		t.Fatalf("expected default capture mode discovery, got %q", application.cfg.CaptureMode)
	}
}
