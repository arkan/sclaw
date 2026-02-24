package config

import (
	"errors"
	"fmt"

	"github.com/flemzord/sclaw/internal/core"
)

// Validate checks the structural validity of a Config.
// It verifies the version field, ensures modules are present,
// and checks that all referenced module IDs exist in the registry.
// It also enforces that Configurable modules have a config entry.
func Validate(cfg *Config) error {
	var errs []error

	if cfg.Version == "" {
		errs = append(errs, errors.New("config: version field is required"))
	} else if cfg.Version != "1" {
		errs = append(errs, fmt.Errorf("config: unsupported version %q (supported: \"1\")", cfg.Version))
	}

	if len(cfg.Modules) == 0 {
		errs = append(errs, errors.New("config: at least one module must be configured"))
	}

	for id := range cfg.Modules {
		if _, ok := core.GetModule(id); !ok {
			errs = append(errs, fmt.Errorf("config: unknown module %q", id))
		}
	}

	// Strict check: registered Configurable modules must have a config entry.
	for _, info := range core.GetModules() {
		mod := info.New()
		if _, ok := mod.(core.Configurable); ok {
			if _, exists := cfg.Modules[string(info.ID)]; !exists {
				errs = append(errs, fmt.Errorf("config: module %q requires configuration but has no entry", info.ID))
			}
		}
	}

	return errors.Join(errs...)
}
