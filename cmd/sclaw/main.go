// Package main is the entry point for the sclaw CLI.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/flemzord/sclaw/internal/core"
	"github.com/spf13/cobra"
)

// Set by goreleaser ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "sclaw",
		Short:         "A plugin-first, self-hosted personal AI assistant",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(versionCmd(), startCmd(), configCmd())
	return root
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and compiled modules",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("sclaw %s (commit: %s, built: %s)\n", version, commit, date)
			mods := core.GetModules()
			if len(mods) == 0 {
				fmt.Println("\nNo compiled modules.")
				return
			}
			fmt.Println("\nCompiled modules:")
			for _, mod := range mods {
				fmt.Printf("  %s\n", mod.ID)
			}
		},
	}
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start sclaw with all configured modules",
		RunE: func(_ *cobra.Command, _ []string) error {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			appCtx := core.NewAppContext(logger, defaultDataDir(), defaultWorkspace())
			app := core.NewApp(appCtx)

			// Phase 1: load all registered modules (no config filtering yet).
			mods := core.GetModules()
			ids := make([]string, 0, len(mods))
			for _, mod := range mods {
				ids = append(ids, string(mod.ID))
			}

			if err := app.LoadModules(ids); err != nil {
				return err
			}

			return app.Run()
		},
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Validate configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
			appCtx := core.NewAppContext(logger, defaultDataDir(), defaultWorkspace())

			for _, info := range core.GetModules() {
				if _, err := appCtx.LoadModule(string(info.ID)); err != nil {
					return fmt.Errorf("module %s: %w", info.ID, err)
				}
			}

			fmt.Println("Configuration OK")
			return nil
		},
	})
	return cmd
}

func defaultDataDir() string {
	if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		return filepath.Join(dir, "sclaw")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sclaw", "data")
}

func defaultWorkspace() string {
	dir, _ := os.Getwd()
	return dir
}
