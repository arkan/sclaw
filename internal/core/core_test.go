package core

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

// lifecycleMod tracks lifecycle calls and supports injecting errors.
type lifecycleMod struct {
	id       ModuleID
	startErr error
	stopErr  error

	mu       sync.Mutex
	started  bool
	stopped  bool
	stopLog  *[]ModuleID // shared slice to track stop order
	startLog *[]ModuleID // shared slice to track start order
}

func (m *lifecycleMod) ModuleInfo() ModuleInfo {
	return ModuleInfo{
		ID: m.id,
		New: func() Module {
			return &lifecycleMod{
				id:       m.id,
				startErr: m.startErr,
				stopErr:  m.stopErr,
				stopLog:  m.stopLog,
				startLog: m.startLog,
			}
		},
	}
}

func (m *lifecycleMod) Provision(_ *AppContext) error { return nil }
func (m *lifecycleMod) Validate() error               { return nil }

func (m *lifecycleMod) Start() error {
	if m.startLog != nil {
		m.mu.Lock()
		*m.startLog = append(*m.startLog, m.id)
		m.mu.Unlock()
	}
	m.started = true
	return m.startErr
}

func (m *lifecycleMod) Stop(_ context.Context) error {
	if m.stopLog != nil {
		m.mu.Lock()
		*m.stopLog = append(*m.stopLog, m.id)
		m.mu.Unlock()
	}
	m.stopped = true
	return m.stopErr
}

func newTestCtx() *AppContext {
	return NewAppContext(
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
		"/tmp/test-data",
		"/tmp/test-ws",
	)
}

func TestApp_LoadModules_Success(t *testing.T) {
	t.Cleanup(resetRegistry)

	RegisterModule(&lifecycleMod{id: "test.a"})
	RegisterModule(&lifecycleMod{id: "test.b"})

	app := NewApp(newTestCtx())
	if err := app.LoadModules([]string{"test.a", "test.b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(app.modules) != 2 {
		t.Errorf("got %d modules, want 2", len(app.modules))
	}
}

func TestApp_LoadModules_FailureRollback(t *testing.T) {
	t.Cleanup(resetRegistry)

	stopLog := &[]ModuleID{}
	RegisterModule(&lifecycleMod{id: "test.ok1", stopLog: stopLog})
	RegisterModule(&lifecycleMod{id: "test.ok2", stopLog: stopLog})

	// Register a module that fails provisioning.
	RegisterModule(&provisionFailMod{id: "test.fail"})

	app := NewApp(newTestCtx())
	err := app.LoadModules([]string{"test.ok1", "test.ok2", "test.fail"})
	if err == nil {
		t.Fatal("expected error on provision failure")
	}

	// The two successfully loaded modules should have been cleaned up.
	if len(*stopLog) != 2 {
		t.Errorf("expected 2 Stop calls during cleanup, got %d", len(*stopLog))
	}
}

type provisionFailMod struct{ id ModuleID }

func (m *provisionFailMod) ModuleInfo() ModuleInfo {
	return ModuleInfo{
		ID:  m.id,
		New: func() Module { return &provisionFailMod{id: m.id} },
	}
}

func (m *provisionFailMod) Provision(_ *AppContext) error {
	return errors.New("provision failed")
}

func (m *provisionFailMod) Stop(_ context.Context) error { return nil }

func TestApp_Start_Success(t *testing.T) {
	t.Cleanup(resetRegistry)

	startLog := &[]ModuleID{}
	RegisterModule(&lifecycleMod{id: "test.x", startLog: startLog})
	RegisterModule(&lifecycleMod{id: "test.y", startLog: startLog})

	app := NewApp(newTestCtx())
	if err := app.LoadModules([]string{"test.x", "test.y"}); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("start error: %v", err)
	}

	if len(*startLog) != 2 {
		t.Errorf("got %d starts, want 2", len(*startLog))
	}
	if (*startLog)[0] != "test.x" || (*startLog)[1] != "test.y" {
		t.Errorf("start order: %v", *startLog)
	}
}

func TestApp_Start_FailureRollback(t *testing.T) {
	t.Cleanup(resetRegistry)

	stopLog := &[]ModuleID{}
	RegisterModule(&lifecycleMod{id: "test.s1", stopLog: stopLog})
	RegisterModule(&lifecycleMod{id: "test.s2", startErr: errors.New("start failed"), stopLog: stopLog})

	app := NewApp(newTestCtx())
	if err := app.LoadModules([]string{"test.s1", "test.s2"}); err != nil {
		t.Fatalf("load error: %v", err)
	}

	err := app.Start()
	if err == nil {
		t.Fatal("expected start error")
	}

	// test.s1 was started and should be stopped during rollback.
	if len(*stopLog) != 1 || (*stopLog)[0] != "test.s1" {
		t.Errorf("expected rollback to stop test.s1, got: %v", *stopLog)
	}
}

func TestApp_Stop_ReverseOrder(t *testing.T) {
	t.Cleanup(resetRegistry)

	stopLog := &[]ModuleID{}
	RegisterModule(&lifecycleMod{id: "test.r1", stopLog: stopLog})
	RegisterModule(&lifecycleMod{id: "test.r2", stopLog: stopLog})
	RegisterModule(&lifecycleMod{id: "test.r3", stopLog: stopLog})

	app := NewApp(newTestCtx())
	if err := app.LoadModules([]string{"test.r1", "test.r2", "test.r3"}); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("start error: %v", err)
	}

	app.Stop()

	want := []ModuleID{"test.r3", "test.r2", "test.r1"}
	if len(*stopLog) != 3 {
		t.Fatalf("got %d stops, want 3", len(*stopLog))
	}
	for i, id := range *stopLog {
		if id != want[i] {
			t.Errorf("stop order[%d] = %q, want %q", i, id, want[i])
		}
	}
}

func TestApp_Run_SignalShutdown(t *testing.T) {
	t.Cleanup(resetRegistry)

	stopLog := &[]ModuleID{}
	RegisterModule(&lifecycleMod{id: "test.sig", stopLog: stopLog})

	app := NewApp(newTestCtx())
	if err := app.LoadModules([]string{"test.sig"}); err != nil {
		t.Fatalf("load error: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	// Give Run() time to set up signal handler and start modules.
	time.Sleep(50 * time.Millisecond)

	// Send SIGINT to ourselves.
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after SIGINT")
	}

	if len(*stopLog) != 1 || (*stopLog)[0] != "test.sig" {
		t.Errorf("expected module to be stopped, got: %v", *stopLog)
	}
}
