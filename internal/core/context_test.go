package core

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"
)

func TestAppContext_ForModule(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx := NewAppContext(logger, "/data", "/workspace")
	child := ctx.ForModule("channel.telegram")

	child.Logger.Info("hello")

	if !bytes.Contains(buf.Bytes(), []byte("channel.telegram")) {
		t.Errorf("expected child logger to contain module ID, got: %s", buf.String())
	}
}

func TestAppContext_LoadModule(t *testing.T) {
	t.Cleanup(resetRegistry)

	provisioned := false
	validated := false

	RegisterModule(&trackingModule{
		id:          "test.loadmod",
		onProvision: func() { provisioned = true },
		onValidate:  func() { validated = true },
	})

	ctx := NewAppContext(nil, "/data", "/ws")
	mod, err := ctx.LoadModule("test.loadmod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod == nil {
		t.Fatal("expected non-nil module")
	}
	if !provisioned {
		t.Error("expected Provision to be called")
	}
	if !validated {
		t.Error("expected Validate to be called")
	}
}

func TestAppContext_LoadModule_UnknownID(t *testing.T) {
	t.Cleanup(resetRegistry)

	ctx := NewAppContext(nil, "/data", "/ws")
	_, err := ctx.LoadModule("does.not.exist")
	if err == nil {
		t.Fatal("expected error for unknown module")
	}
}

func TestAppContext_LoadModule_ProvisionError(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(&trackingModule{
		id:           "test.provfail",
		provisionErr: errors.New("provision boom"),
	})

	ctx := NewAppContext(nil, "/data", "/ws")
	_, err := ctx.LoadModule("test.provfail")
	if err == nil {
		t.Fatal("expected error on provision failure")
	}
}

func TestAppContext_LoadModule_ValidateError(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(&trackingModule{
		id:          "test.valfail",
		validateErr: errors.New("validate boom"),
	})

	ctx := NewAppContext(nil, "/data", "/ws")
	_, err := ctx.LoadModule("test.valfail")
	if err == nil {
		t.Fatal("expected error on validate failure")
	}
}

// trackingModule is a test helper that tracks lifecycle calls.
type trackingModule struct {
	id           ModuleID
	onProvision  func()
	onValidate   func()
	provisionErr error
	validateErr  error
}

func (m *trackingModule) ModuleInfo() ModuleInfo {
	id := m.id
	return ModuleInfo{
		ID: id,
		New: func() Module {
			return &trackingModule{
				id:           id,
				onProvision:  m.onProvision,
				onValidate:   m.onValidate,
				provisionErr: m.provisionErr,
				validateErr:  m.validateErr,
			}
		},
	}
}

func (m *trackingModule) Provision(_ *AppContext) error {
	if m.onProvision != nil {
		m.onProvision()
	}
	return m.provisionErr
}

func (m *trackingModule) Validate() error {
	if m.onValidate != nil {
		m.onValidate()
	}
	return m.validateErr
}
