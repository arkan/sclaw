// Package config handles YAML configuration loading, environment variable
// expansion, and structural validation for sclaw.
package config

import "gopkg.in/yaml.v3"

// Config is the top-level configuration structure.
type Config struct {
	// Version is the config format version. Currently only "1" is supported.
	Version string `yaml:"version"`

	// Modules maps module IDs to their raw YAML configuration.
	// Keys must match registered module IDs (e.g. "channel.telegram").
	Modules map[string]yaml.Node `yaml:"modules"`
}
