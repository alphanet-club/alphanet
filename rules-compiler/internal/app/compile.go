package app

import (
	"context"
	"fmt"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/compiler"
	"github.com/alphanet/rules-compiler/internal/terminal"
)

// RunCompile orchestrates the full compilation workflow from a strategy directory.
// It loads sources, compiles, and writes outputs.
func RunCompile(ctx context.Context, strategyDir string, opts compiler.Options) error {
	// Resolve output directory
	outDir := opts.OutDir
	if outDir == "" {
		outDir = strategyDir + "/compiled"
	}

	if opts.Verbose {
		terminal.Step("Starting AlphaNet compile")
		terminal.Info("  Strategy: %s", strategyDir)
		terminal.Info("  Output: %s", outDir)
		terminal.Info("  Requested mode: %s", resolveModeLabel(opts.ModeOverride))
		if opts.EngineOverride != "" {
			terminal.Info("  Requested engine: %s", opts.EngineOverride)
		}
	}

	// Run compilation
	result, err := compiler.Compile(ctx, strategyDir, opts)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	if opts.Verbose {
		terminal.Step("Compilation summary")
		terminal.Info("  Mode: %s", resolveModeLabel(opts.ModeOverride))
		if opts.EngineOverride != "" {
			terminal.Info("  Engine override: %s", opts.EngineOverride)
		}
		terminal.Info("  Strategy: %s", strategyDir)
		terminal.Info("  Output: %s", outDir)
		if result.ValidationReport.Status == "valid" {
			terminal.Success("  Validation status: %s", result.ValidationReport.Status)
		} else if result.ValidationReport.Status == "warning" {
			terminal.Warn("  Validation status: %s", result.ValidationReport.Status)
		} else {
			terminal.Error("  Validation status: %s", result.ValidationReport.Status)
		}
		if len(result.Warnings) > 0 {
			terminal.Warn("  Warnings (%d):", len(result.Warnings))
			for _, w := range result.Warnings {
				terminal.Warn("    - %s", w)
			}
		}
	}

	// If validate-only, print and return
	if opts.ValidateOnly {
		fmt.Println("=== Validation Report ===")
		vr := result.ValidationReport
		fmt.Printf("Status: %s\n", vr.Status)
		for _, c := range vr.Checks {
			fmt.Printf("  [%s] %s", c.Status, c.Name)
			if c.Detail != "" {
				fmt.Printf(": %s", c.Detail)
			}
			fmt.Println()
		}
		for _, w := range vr.Warnings {
			fmt.Printf("  [WARN] %s\n", w)
		}
		for _, e := range vr.Errors {
			fmt.Printf("  [ERROR] %s\n", e)
		}
		return nil
	}

	// Write outputs
	if err := air.WriteOutputs(result.AIR, result.Provenance, result.Reasoning, result.ValidationReport, outDir, opts.DryRun); err != nil {
		return fmt.Errorf("writing outputs: %w", err)
	}

	if opts.DryRun {
		terminal.Success("Dry-run complete. Would write to: %s", outDir)
	} else {
		terminal.Success("Compilation complete. Output written to: %s", outDir)
	}

	return nil
}

func resolveModeLabel(modeOverride string) string {
	if modeOverride != "" {
		return modeOverride
	}
	return "manual (default)"
}
