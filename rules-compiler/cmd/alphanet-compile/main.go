package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/alphanet/rules-compiler/internal/app"
	"github.com/alphanet/rules-compiler/internal/compiler"
)

func main() {
	// Define flags
	specDir := flag.String("spec", "specs/v0.1", "Path to specs directory")
	outDir := flag.String("out", "", "Output directory (default: <strategy-dir>/compiled)")
	mode := flag.String("mode", "", "Override compiler mode (none, manual, single, ensemble)")
	dryRun := flag.Bool("dry-run", false, "Validate and print without writing files")
	validateOnly := flag.Bool("validate-only", false, "Run validation only, skip compilation")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: alphanet-compile <strategy-dir> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Compile an AlphaNet strategy folder into AIR (AlphaNet Intermediate Representation).\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  strategy-dir    Path to strategy folder containing manifest.json, strategy.md, rules.json\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validate position arg
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	strategyDir := flag.Arg(0)

	// Validate directory exists
	info, err := os.Stat(strategyDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: strategy directory %q: %v\n", strategyDir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %q is not a directory\n", strategyDir)
		os.Exit(1)
	}

	opts := compiler.Options{
		ModeOverride: *mode,
		SpecDir:      *specDir,
		OutDir:       *outDir,
		DryRun:       *dryRun,
		ValidateOnly: *validateOnly,
		Verbose:      *verbose,
	}

	ctx := context.Background()
	if err := app.RunCompile(ctx, strategyDir, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
