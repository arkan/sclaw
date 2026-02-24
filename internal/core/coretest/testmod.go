// Package coretest provides test helpers and fake modules for core tests.
package coretest

import (
	"context"

	"github.com/flemzord/sclaw/internal/core"
)

func init() {
	core.RegisterModule(FakeModule{})
}

// FakeModule is a test module that tracks lifecycle calls.
type FakeModule struct {
	ProvisionCalled bool
	ValidateCalled  bool
	StartCalled     bool
	StopCalled      bool

	ProvisionErr error
	ValidateErr  error
	StartErr     error
	StopErr      error
}

// ModuleInfo returns the module's metadata.
func (FakeModule) ModuleInfo() core.ModuleInfo {
	return core.ModuleInfo{
		ID:  "test.fake",
		New: func() core.Module { return &FakeModule{} },
	}
}

// Provision implements core.Provisioner.
func (m *FakeModule) Provision(_ *core.AppContext) error {
	m.ProvisionCalled = true
	return m.ProvisionErr
}

// Validate implements core.Validator.
func (m *FakeModule) Validate() error {
	m.ValidateCalled = true
	return m.ValidateErr
}

// Start implements core.Starter.
func (m *FakeModule) Start() error {
	m.StartCalled = true
	return m.StartErr
}

// Stop implements core.Stopper.
func (m *FakeModule) Stop(_ context.Context) error {
	m.StopCalled = true
	return m.StopErr
}

// Interface guards.
var (
	_ core.Module      = (*FakeModule)(nil)
	_ core.Provisioner = (*FakeModule)(nil)
	_ core.Validator   = (*FakeModule)(nil)
	_ core.Starter     = (*FakeModule)(nil)
	_ core.Stopper     = (*FakeModule)(nil)
)
