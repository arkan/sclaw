package core

import "strings"

// ModuleID is a hierarchical dot-separated identifier (e.g., "channel.telegram").
type ModuleID string

// Namespace returns the portion before the last dot.
// For "channel.telegram" it returns "channel".
// For a single-segment ID it returns "".
func (id ModuleID) Namespace() string {
	idx := strings.LastIndex(string(id), ".")
	if idx < 0 {
		return ""
	}
	return string(id)[:idx]
}

// Name returns the portion after the last dot.
// For "channel.telegram" it returns "telegram".
// For a single-segment ID it returns the full ID.
func (id ModuleID) Name() string {
	idx := strings.LastIndex(string(id), ".")
	if idx < 0 {
		return string(id)
	}
	return string(id)[idx+1:]
}

// ModuleInfo holds metadata about a module, returned by Module.ModuleInfo().
type ModuleInfo struct {
	// ID is the unique hierarchical identifier (e.g., "channel.telegram").
	ID ModuleID

	// New returns a new, empty instance of the module's type.
	// The returned value should be a pointer.
	New func() Module
}

// Module is the base interface that all sclaw modules must implement.
type Module interface {
	ModuleInfo() ModuleInfo
}
