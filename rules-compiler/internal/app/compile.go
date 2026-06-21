package app

import (
	"context"
	"fmt"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/compiler"
)

// RunCompile orchestrates the full compilation workflow from a strategy directory.
// It loads sources, compiles, and writes outputs.
func RunCompile(ctx context.Context, strategyDir string, opts compiler.Options) error {
	// Resolve output directory
	outDir := opts.OutDir
	if outDir == "" {
		outDir = strategyDir + "/compiled"
	}

	// Run compilation
	result, err := compiler.Compile(ctx, strategyDir, opts)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("Compilation mode: %s\n", resolveModeLabel(opts.ModeOverride))
		fmt.Printf("Strategy: %s\n", strategyDir)
		fmt.Printf("Output: %s\n", outDir)
		fmt.Printf("Validation status: %s\n", result.ValidationReport.Status)
		if len(result.Warnings) > 0 {
			fmt.Printf("Warnings (%d):\n", len(result.Warnings))
			for _, w := range result.Warnings {
				fmt.Printf("  - %s\n", w)
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
		fmt.Printf("Dry-run complete. Would write to: %s\n", outDir)
	} else {
		fmt.Printf("Compilation complete. Output written to: %s\n", outDir)
	}

	return nil
}

func resolveModeLabel(modeOverride string) string {
	if modeOverride != "" {
		return modeOverride
	}
	return "manual (default)"
}
