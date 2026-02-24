package config

import (
	"strings"
	"testing"

	"github.com/flemzord/sclaw/internal/core"
	"gopkg.in/yaml.v3"
)

// stubModule is a basic module for testing.
type stubModule struct {
	id string
}

func (m *stubModule) ModuleInfo() core.ModuleInfo {
	return core.ModuleInfo{
		ID:  core.ModuleID(m.id),
		New: func() core.Module { return &stubModule{id: m.id} },
	}
}

// configurableModule implements core.Configurable.
type configurableModule struct {
	stubModule
}

func (m *configurableModule) ModuleInfo() core.ModuleInfo {
	return core.ModuleInfo{
		ID:  core.ModuleID(m.id),
		New: func() core.Module { return &configurableModule{stubModule: stubModule{id: m.id}} },
	}
}

func (m *configurableModule) Configure(_ yaml.Node) error { return nil }

func registerStub(t *testing.T, id string) {
	t.Helper()
	core.RegisterModule(&stubModule{id: id})
	t.Cleanup(func() { core.ResetRegistry() })
}

func registerConfigurable(t *testing.T, id string) {
	t.Helper()
	core.RegisterModule(&configurableModule{stubModule: stubModule{id: id}})
	t.Cleanup(func() { core.ResetRegistry() })
}

func TestValidate_Valid(t *testing.T) {
	registerStub(t, "test.mod")
	cfg := &Config{
		Version: "1",
		Modules: map[string]yaml.Node{"test.mod": {}},
	}
	if err := Validate(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_MissingVersion(t *testing.T) {
	registerStub(t, "test.mod")
	cfg := &Config{
		Modules: map[string]yaml.Node{"test.mod": {}},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error should mention version: %v", err)
	}
}

func TestValidate_UnsupportedVersion(t *testing.T) {
	registerStub(t, "test.mod")
	cfg := &Config{
		Version: "99",
		Modules: map[string]yaml.Node{"test.mod": {}},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error should mention unsupported: %v", err)
	}
}

func TestValidate_EmptyModules(t *testing.T) {
	cfg := &Config{
		Version: "1",
		Modules: map[string]yaml.Node{},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for empty modules")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Errorf("error should mention at least one module: %v", err)
	}
}

func TestValidate_UnknownModule(t *testing.T) {
	cfg := &Config{
		Version: "1",
		Modules: map[string]yaml.Node{"unknown.mod": {}},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for unknown module")
	}
	if !strings.Contains(err.Error(), "unknown.mod") {
		t.Errorf("error should mention module ID: %v", err)
	}
}

func TestValidate_MultipleUnknown(t *testing.T) {
	cfg := &Config{
		Version: "1",
		Modules: map[string]yaml.Node{
			"bad.one": {},
			"bad.two": {},
		},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for unknown modules")
	}
	if !strings.Contains(err.Error(), "bad.one") || !strings.Contains(err.Error(), "bad.two") {
		t.Errorf("error should mention both modules: %v", err)
	}
}

func TestValidate_ConfigurableModuleMissingConfig(t *testing.T) {
	registerConfigurable(t, "need.config")
	cfg := &Config{
		Version: "1",
		Modules: map[string]yaml.Node{"need.config": {}},
	}
	// Should pass — configurable module has an entry.
	if err := Validate(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_ConfigurableModuleNoEntry(t *testing.T) {
	registerConfigurable(t, "need.config")
	registerStub(t, "other.mod")
	cfg := &Config{
		Version: "1",
		Modules: map[string]yaml.Node{"other.mod": {}},
	}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected error for configurable module without config entry")
	}
	if !strings.Contains(err.Error(), "need.config") {
		t.Errorf("error should mention need.config: %v", err)
	}
	if !strings.Contains(err.Error(), "requires configuration") {
		t.Errorf("error should mention requires configuration: %v", err)
	}
}
