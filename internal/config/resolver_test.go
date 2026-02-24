package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestResolve_SortedOrder(t *testing.T) {
	cfg := &Config{
		Modules: map[string]yaml.Node{
			"provider.openai":    {},
			"channel.telegram":   {},
			"channel.discord":    {},
			"provider.anthropic": {},
		},
	}
	ids := Resolve(cfg)
	want := []string{"channel.discord", "channel.telegram", "provider.anthropic", "provider.openai"}
	if len(ids) != len(want) {
		t.Fatalf("got %d IDs, want %d", len(ids), len(want))
	}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("ids[%d] = %q, want %q", i, id, want[i])
		}
	}
}

func TestResolve_Empty(t *testing.T) {
	cfg := &Config{
		Modules: map[string]yaml.Node{},
	}
	ids := Resolve(cfg)
	if len(ids) != 0 {
		t.Errorf("got %d IDs, want 0", len(ids))
	}
}
