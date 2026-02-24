package core

import "testing"

func TestModuleID_Namespace(t *testing.T) {
	tests := []struct {
		id   ModuleID
		want string
	}{
		{"channel.telegram", "channel"},
		{"provider.anthropic", "provider"},
		{"a.b.c", "a.b"},
		{"solo", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tt.id.Namespace(); got != tt.want {
			t.Errorf("ModuleID(%q).Namespace() = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestModuleID_Name(t *testing.T) {
	tests := []struct {
		id   ModuleID
		want string
	}{
		{"channel.telegram", "telegram"},
		{"provider.anthropic", "anthropic"},
		{"a.b.c", "c"},
		{"solo", "solo"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tt.id.Name(); got != tt.want {
			t.Errorf("ModuleID(%q).Name() = %q, want %q", tt.id, got, tt.want)
		}
	}
}
