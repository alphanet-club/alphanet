package compiler

import (
	"context"
	"fmt"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/engines"
	"github.com/alphanet/rules-compiler/internal/provenance"
	"github.com/alphanet/rules-compiler/internal/reasoning"
	"github.com/alphanet/rules-compiler/internal/validation"
)

// EngineFactory creates engine instances based on configuration.
type EngineFactory struct {
	registry map[string]func() (engines.Engine, error)
}

// NewEngineFactory creates a new engine factory with default engines registered.
func NewEngineFactory() *EngineFactory {
	factory := &EngineFactory{
		registry: make(map[string]func() (engines.Engine, error)),
	}

	// Register TradingAgents engine adapter (both short and full names)
	factory.Register("tradingagents", func() (engines.Engine, error) {
		return &engines.TradingAgentsEngine{}, nil
	})
	factory.Register("TauricResearch/TradingAgents", func() (engines.Engine, error) {
		return &engines.TradingAgentsEngine{}, nil
	})

	// Register AI Hedge Fund engine adapter (both short and full names)
	factory.Register("ai-hedge-fund", func() (engines.Engine, error) {
		return &engines.AIHedgeFundEngine{}, nil
	})
	factory.Register("virattt/ai-hedge-fund", func() (engines.Engine, error) {
		return &engines.AIHedgeFundEngine{}, nil
	})

	return factory
}

// Register adds an engine constructor to the factory.
func (f *EngineFactory) Register(name string, constructor func() (engines.Engine, error)) {
	f.registry[name] = constructor
}

// CreateEngine creates an engine instance by name.
func (f *EngineFactory) CreateEngine(name string) (engines.Engine, error) {
	constructor, exists := f.registry[name]
	if !exists {
		return nil, fmt.Errorf("unknown engine: %s", name)
	}
	return constructor()
}

// CreateEngines creates multiple engine instances from configuration and initialises them.
func (f *EngineFactory) CreateEngines(configs []air.EngineConfig) ([]engines.Engine, error) {
	var result []engines.Engine
	for _, cfg := range configs {
		engine, err := f.CreateEngine(cfg.Name)
		if err != nil {
			return nil, fmt.Errorf("creating engine %s: %w", cfg.Name, err)
		}
		if err := engine.Init(cfg); err != nil {
			return nil, fmt.Errorf("initialising engine %s: %w", cfg.Name, err)
		}
		result = append(result, engine)
	}
	return result, nil
}

// compileWithEngines handles "single" and "ensemble" modes by calling agent engines
// and merging their output into the AIR artifact.
func compileWithEngines(ctx context.Context, src *air.SourceContext, normRules []air.Rule, normPortfolio *air.AIRPortfolio, warnings []string, opts Options) (*CompileResult, error) {
	// Determine the original compiler mode
	compilerMode := resolveOriginalMode(src.Manifest, opts.ModeOverride)

	// Create engine factory
	factory := NewEngineFactory()

	// Attempt to create engines from manifest configuration
	engineList, err := factory.CreateEngines(src.Manifest.Compiler.Engines)
	if err != nil {
		return nil, fmt.Errorf("creating engines: %w", err)
	}

	if len(engineList) == 0 {
		allWarnings := warnings
		allWarnings = append(allWarnings, fmt.Sprintf("Mode %q requires engine support but no engines configured.", compilerMode))
		return compileManual(src, normRules, normPortfolio, allWarnings, opts)
	}

	// Build engine input from source context
	engineInput := engines.EngineInput{
		Manifest:       *src.Manifest,
		StrategyMD:     src.StrategyMD,
		Rules:          src.Rules,
		Signals:        src.Signals,
		Relations:      src.Relations,
		Regimes:        src.Regimes,
		TrainingWindow: src.Manifest.Compiler.TrainingWindow,
		Portfolio:      src.Manifest.Portfolio,
	}

	// Collect symbols for engine analysis
	for _, sym := range src.Manifest.Universe.Symbols {
		if sym != "cash" && sym != "CASH" {
			engineInput.Universe = append(engineInput.Universe, sym)
		}
	}

	// Call each engine and merge results
	var allSignals []air.Signal
	var allNotes []string

	for _, engine := range engineList {
		output, err := engine.Analyze(ctx, engineInput)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Engine '%s' error: %v", engine.Name(), err))
			continue
		}

		// Merge signals from engine output
		for _, sig := range output.Signals {
			if sig.Action == "add" || sig.Action == "" {
				allSignals = append(allSignals, sig.Signal)
			}
		}

		// Collect notes
		if output.Notes != "" {
			allNotes = append(allNotes, output.Notes)
		}
	}

	// Create enriched source context with engine-provided signals
	enrichedSrc := *src
	enrichedSrc.Signals = append(enrichedSrc.Signals, allSignals...)

	// Build execution config enriched with backtest data
	decisionHierarchy := air.DefaultDecisionHierarchy()
	execConfig := air.EnrichExecutionConfig(air.DefaultExecutionConfig(), src.Manifest.Backtest)

	// Build AI R with engine-enriched signals
	airData := air.BuildAIR(enrichedSrc, normPortfolio, normRules, decisionHierarchy, execConfig, nil)

	// Build provenance with engine metadata
	prov := provenance.Build(&enrichedSrc, compilerMode, src.Manifest.Compiler.Engines, "", nil)

	// Compute IR hash
	canonical, err := air.CanonicalJSON(airData)
	if err != nil {
		return nil, fmt.Errorf("computing canonical AIR: %w", err)
	}
	irHash := air.HashBytes(canonical)
	outputHashes := computeOutputHashes(airData, irHash)

	// Rebuild provenance with hashes
	prov = provenance.Build(&enrichedSrc, compilerMode, src.Manifest.Compiler.Engines, irHash, outputHashes)
	airData.Provenance = prov

	// Build reasoning
	reason := reasoning.Build(&enrichedSrc, compilerMode, normRules, warnings)

	// Build validation report
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

// resolveOriginalMode returns the original compiler mode, accounting for overrides.
func resolveOriginalMode(m *air.Manifest, override string) string {
	if override != "" {
		return override
	}
	if m.Compiler.Mode != "" {
		return m.Compiler.Mode
	}
	return "manual"
}
