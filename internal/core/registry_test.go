package core

import "testing"

type stubModule struct{ id ModuleID }

func (m stubModule) ModuleInfo() ModuleInfo {
	return ModuleInfo{
		ID:  m.id,
		New: func() Module { return &stubModule{id: m.id} },
	}
}

func TestRegisterAndGetModule(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(stubModule{id: "test.alpha"})

	info, ok := GetModule("test.alpha")
	if !ok {
		t.Fatal("expected module to be found")
	}
	if info.ID != "test.alpha" {
		t.Errorf("got ID %q, want %q", info.ID, "test.alpha")
	}
	if info.New == nil {
		t.Error("New should not be nil")
	}
}

func TestGetModule_NotFound(t *testing.T) {
	t.Cleanup(resetRegistry)

	_, ok := GetModule("nonexistent")
	if ok {
		t.Error("expected module not to be found")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(stubModule{id: "test.dup"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	RegisterModule(stubModule{id: "test.dup"})
}

func TestRegisterEmptyIDPanics(t *testing.T) {
	t.Cleanup(resetRegistry)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on empty ID")
		}
	}()
	RegisterModule(stubModule{id: ""})
}

func TestRegisterNilNewPanics(t *testing.T) {
	t.Cleanup(resetRegistry)

	mod := &nilNewModule{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil New")
		}
	}()
	RegisterModule(mod)
}

type nilNewModule struct{}

func (nilNewModule) ModuleInfo() ModuleInfo {
	return ModuleInfo{ID: "test.nilnew", New: nil}
}

func TestGetModules_Sorted(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(stubModule{id: "provider.openai"})
	RegisterModule(stubModule{id: "channel.discord"})
	RegisterModule(stubModule{id: "channel.telegram"})

	mods := GetModules()
	if len(mods) != 3 {
		t.Fatalf("got %d modules, want 3", len(mods))
	}

	want := []ModuleID{"channel.discord", "channel.telegram", "provider.openai"}
	for i, m := range mods {
		if m.ID != want[i] {
			t.Errorf("index %d: got %q, want %q", i, m.ID, want[i])
		}
	}
}

func TestGetModulesByNamespace(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(stubModule{id: "channel.discord"})
	RegisterModule(stubModule{id: "channel.telegram"})
	RegisterModule(stubModule{id: "provider.openai"})

	channels := GetModulesByNamespace("channel")
	if len(channels) != 2 {
		t.Fatalf("got %d channel modules, want 2", len(channels))
	}
	if channels[0].ID != "channel.discord" || channels[1].ID != "channel.telegram" {
		t.Errorf("unexpected modules: %v, %v", channels[0].ID, channels[1].ID)
	}
}

func TestGetModulesByNamespace_Empty(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(stubModule{id: "channel.discord"})

	mods := GetModulesByNamespace("nonexistent")
	if len(mods) != 0 {
		t.Errorf("got %d modules, want 0", len(mods))
	}
}
