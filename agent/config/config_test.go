package config

import "testing"

func TestCapabilitiesFromConfig(t *testing.T) {
	cfg := Default()
	caps := cfg.ProviderCapabilities()
	if !caps.SupportsToolCalling {
		t.Fatal("expected tool calling")
	}
}
