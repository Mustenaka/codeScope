package session

import "testing"

func TestBridgeRegistryRegisterReplacesOlderConnection(t *testing.T) {
	registry := NewBridgeRegistry()

	first := make(chan any, 1)
	second := make(chan any, 1)

	unregisterFirst := registry.Register("session-1", first)
	if !registry.IsConnected("session-1") {
		t.Fatalf("expected session to be connected after first register")
	}

	unregisterSecond := registry.Register("session-1", second)

	unregisterFirst()
	if !registry.IsConnected("session-1") {
		t.Fatalf("expected second bridge registration to remain active")
	}

	if err := registry.Send("session-1", "payload"); err != nil {
		t.Fatalf("send via replacement bridge: %v", err)
	}

	select {
	case got := <-second:
		if got != "payload" {
			t.Fatalf("expected payload on replacement bridge, got %#v", got)
		}
	default:
		t.Fatalf("expected replacement bridge to receive payload")
	}

	select {
	case got := <-first:
		t.Fatalf("expected original bridge to remain inactive, got %#v", got)
	default:
	}

	unregisterSecond()
	if registry.IsConnected("session-1") {
		t.Fatalf("expected session to be disconnected after unregister")
	}
}
