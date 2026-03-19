package app

import (
	"testing"

	"codescope/bridge/internal/capture"
)

func TestNewDefaultsToDiscoveryCollector(t *testing.T) {
	t.Setenv("CODESCOPE_BRIDGE_CAPTURE_MODE", "")

	application := New()

	if application.cfg.CaptureMode != "discovery" {
		t.Fatalf("expected default capture mode discovery, got %q", application.cfg.CaptureMode)
	}
}

func TestNewCaptureSourceIncludesSemanticCaptureSourceInDiscoveryMode(t *testing.T) {
	t.Setenv("CODESCOPE_BRIDGE_CAPTURE_MODE", "discovery")

	application := New()

	multi, ok := application.source.(capture.MultiSource)
	if !ok {
		t.Fatalf("expected discovery mode to use MultiSource, got %T", application.source)
	}
	if len(multi.Sources()) < 2 {
		t.Fatalf("expected discovery + semantic sources, got %d source(s)", len(multi.Sources()))
	}
	if _, ok := multi.Sources()[1].(capture.SemanticCaptureSource); !ok {
		t.Fatalf("expected second source to be SemanticCaptureSource, got %T", multi.Sources()[1])
	}
}
