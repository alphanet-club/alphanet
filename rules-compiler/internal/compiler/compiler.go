package compiler

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/manifest"
	"github.com/alphanet/rules-compiler/internal/portfolio"
	"github.com/alphanet/rules-compiler/internal/provenance"
	"github.com/alphanet/rules-compiler/internal/reasoning"
	"github.com/alphanet/rules-compiler/internal/rules"
	"github.com/alphanet/rules-compiler/internal/strategy"
	"github.com/alphanet/rules-compiler/internal/validation"
)

// CompileResult holds all outputs from a compilation run.
type CompileResult struct {
	AIR              *air.AIR
	Provenance       *air.Provenance
	Reasoning        string
	ValidationReport *air.ValidationReport
	Warnings         []string
}

// Options configures the compilation.
type Options struct {
	ModeOverride  string
	SpecDir       string
	OutDir        string
	DryRun        bool
	ValidateOnly  bool
	Verbose       bool
	EmitReasoning bool
}

// Compile runs the full compilation pipeline.
func Compile(ctx context.Context, strategyDir string, opts Options) (*CompileResult, error) {
	// 1. Load source files
	src, err := loadSources(strategyDir)
	if err != nil {
		return nil, fmt.Errorf("loading sources: %w", err)
	}

	// 2. Validate manifest
	if err := manifest.ValidateManifest(src.Manifest); err != nil {
		return nil, fmt.Errorf("validating manifest: %w", err)
	}

	// 3. Normalize rules
	normRules, ruleWarnings, err := normalizeRules(src.Rules)
	if err != nil {
		return nil, fmt.Errorf("normalizing rules: %w", err)
	}

	// 4. Normalize portfolio
	normPortfolio, portWarnings := portfolio.NormalizePortfolio(src.Manifest)

	// 5. Resolve compiler mode
	mode := resolveMode(src.Manifest, opts.ModeOverride)

	// Merge warnings
	var allWarnings []string
	allWarnings = append(allWarnings, ruleWarnings...)
	allWarnings = append(allWarnings, portWarnings...)

	// 6. Compile based on mode
	switch mode {
	case "none":
		return compileNone(src, opts)
	case "manual":
		return compileManual(src, normRules, normPortfolio, allWarnings, opts)
	case "single", "ensemble":
		return compileWithEngines(ctx, src, normRules, normPortfolio, allWarnings, opts)
	default:
		return nil, fmt.Errorf("unknown compiler mode: %q", mode)
	}
}

// loadSources loads all source files from the strategy directory.
func loadSources(dir string) (*air.SourceContext, error) {
	m, err := manifest.LoadManifest(dir)
	if err != nil {
		return nil, err
	}

	manifestRaw, err := os.ReadFile(dir + "/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("reading manifest.json for hash: %w", err)
	}

	stratMD, err := strategy.LoadStrategyMD(dir)
	if err != nil {
		return nil, err
	}

	strategyRaw, _ := os.ReadFile(dir + "/strategy.md")

	rulesList, rulesRaw, err := rules.LoadRules(dir)
	if err != nil {
		return nil, err
	}
	if rulesList == nil {
		rulesList = []air.Rule{}
	}
	if rulesRaw == nil {
		rulesRaw = []byte{}
	}

	now := time.Now().UTC().Format(time.RFC3339)

	return &air.SourceContext{
		Manifest:    m,
		StrategyMD:  stratMD,
		Rules:       rulesList,
		Signals:     []air.Signal{},
		Relations:   []air.Relation{},
		Regimes:     []air.Regime{},
		GeneratedAt: now,
		ManifestRaw: manifestRaw,
		StrategyRaw: strategyRaw,
		RulesRaw:    rulesRaw,
	}, nil
}

func normalizeRules(rulesList []air.Rule) ([]air.Rule, []string, error) {
	var warnings []string
	for i := range rulesList {
		rulesList[i].Enabled = true
	}
	_, err := rules.NormalizeRules(rulesList)
	if err != nil {
		return rulesList, warnings, err
	}
	return rulesList, warnings, nil
}

func resolveMode(m *air.Manifest, override string) string {
	if override != "" {
		return override
	}
	if m.Compiler.Mode != "" {
		return m.Compiler.Mode
	}
	return "manual"
}

func compileNone(src *air.SourceContext, opts Options) (*CompileResult, error) {
	prov := provenance.Build(src, "none", nil, "", nil)
	reason := reasoning.Build(src, "none", nil, nil)
	vr := &air.ValidationReport{Status: "valid"}

	return &CompileResult{
		AIR:              nil,
		Provenance:       prov,
		Reasoning:        reason,
		ValidationReport: vr,
	}, nil
}

func compileManual(src *air.SourceContext, normRules []air.Rule, normPortfolio *air.AIRPortfolio, warnings []string, opts Options) (*CompileResult, error) {
	decisionHierarchy := air.DefaultDecisionHierarchy()
	execConfig := air.DefaultExecutionConfig()

	airData := air.BuildAIR(*src, normPortfolio, normRules, decisionHierarchy, execConfig, nil)
	prov := provenance.Build(src, "manual", nil, "", nil)

	canonical, err := air.CanonicalJSON(airData)
	if err != nil {
		return nil, fmt.Errorf("computing canonical AIR: %w", err)
	}
	irHash := air.HashBytes(canonical)
	outputHashes := computeOutputHashes(airData, irHash)

	prov = provenance.Build(src, "manual", nil, irHash, outputHashes)
	airData.Provenance = prov

	reason := reasoning.Build(src, "manual", normRules, warnings)
	vr := validation.BuildReport(airData, warnings, src.Manifest.SpecVersion)

	if opts.ValidateOnly {
		return &CompileResult{
			AIR:              nil,
			Provenance:       prov,
			Reasoning:        reason,
			ValidationReport: vr,
			Warnings:         warnings,
		}, nil
	}

	return &CompileResult{
		AIR:              airData,
		Provenance:       prov,
		Reasoning:        reason,
		ValidationReport: vr,
		Warnings:         warnings,
	}, nil
}

func computeOutputHashes(airData *air.AIR, irHash string) map[string]string {
	hashes := make(map[string]string)
	hashes["strategy.ir.json"] = irHash
	hashes["reasoning.md"] = "sha256:example"
	hashes["validation-report.json"] = "sha256:example"
	return hashes
}
