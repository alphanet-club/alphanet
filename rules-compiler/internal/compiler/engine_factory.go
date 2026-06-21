package compiler

import (
	"context"
	"fmt"
	"strings"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/engines"
	"github.com/alphanet/rules-compiler/internal/provenance"
	"github.com/alphanet/rules-compiler/internal/reasoning"
	"github.com/alphanet/rules-compiler/internal/validation"
)

type EngineFactory struct {
	registry map[string]func() (engines.Engine, error)
}

func NewEngineFactory() *EngineFactory {
	factory := &EngineFactory{registry: map[string]func() (engines.Engine, error){}}
	factory.Register("tradingagents", func() (engines.Engine, error) {
		return &engines.TradingAgentsEngine{}, nil
	})
	factory.Register("TauricResearch/TradingAgents", func() (engines.Engine, error) {
		return &engines.TradingAgentsEngine{}, nil
	})
	factory.Register("ai-hedge-fund", func() (engines.Engine, error) {
		return &engines.AIHedgeFundEngine{}, nil
	})
	factory.Register("virattt/ai-hedge-fund", func() (engines.Engine, error) {
		return &engines.AIHedgeFundEngine{}, nil
	})
	return factory
}

func (f *EngineFactory) Register(name string, constructor func() (engines.Engine, error)) {
	f.registry[name] = constructor
}

func (f *EngineFactory) CreateEngine(name string) (engines.Engine, error) {
	constructor, ok := f.registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown engine: %s", name)
	}
	return constructor()
}

func (f *EngineFactory) CreateEngines(configs []air.EngineConfig) ([]engines.Engine, error) {
	var result []engines.Engine
	for _, cfg := range configs {
		if cfg.Enabled != nil && !*cfg.Enabled {
			continue
		}
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

func compileWithEngines(ctx context.Context, src *air.SourceContext, normRules []air.Rule, normPortfolio *air.AIRPortfolio, warnings []string, opts Options) (*CompileResult, error) {
	compilerMode := resolveOriginalMode(src.Manifest, opts.ModeOverride)
	engineConfigs := activeEngineConfigs(src.Manifest.Compiler.Engines)

	if compilerMode == "single" && len(engineConfigs) > 1 {
		warnings = append(warnings, fmt.Sprintf("mode single selected; using first enabled engine only: %s", engineConfigs[0].Name))
		engineConfigs = engineConfigs[:1]
	}

	factory := NewEngineFactory()
	engineList, err := factory.CreateEngines(engineConfigs)
	if err != nil {
		return nil, fmt.Errorf("creating engines: %w", err)
	}
	if len(engineList) == 0 {
		warnings = append(warnings, fmt.Sprintf("mode %q requested but no enabled engines are configured", compilerMode))
		return compileManual(src, normRules, normPortfolio, warnings, opts)
	}

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
	for _, sym := range src.Manifest.Universe.Symbols {
		if !strings.EqualFold(sym, "cash") {
			engineInput.Universe = append(engineInput.Universe, sym)
		}
	}

	var engineSignals []air.Signal
	var engineNotes []string
	var engineReports []engines.EngineReport

	for _, engine := range engineList {
		output, err := engine.Analyze(ctx, engineInput)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Engine %q error: %v", engine.Name(), err))
			continue
		}
		for _, sig := range output.Signals {
			if sig.Action == "" || sig.Action == "add" {
				engineSignals = append(engineSignals, sig.Signal)
			}
		}
		if strings.TrimSpace(output.Notes) != "" {
			engineNotes = append(engineNotes, output.Notes)
		}
		engineReports = append(engineReports, output.Reports...)
	}

	enrichedSrc := *src
	enrichedSrc.Signals = append(enrichedSrc.Signals, engineSignals...)

	decisionHierarchy := air.DefaultDecisionHierarchy()
	execConfig := air.EnrichExecutionConfig(air.DefaultExecutionConfig(), src.Manifest.Backtest)
	airData := air.BuildAIR(enrichedSrc, normPortfolio, normRules, decisionHierarchy, execConfig, nil)
	if len(engineReports) > 0 {
		reportRefs, reportArtifacts := buildAgentReportArtifacts(engineReports)
		airData.AgentReportContents = reportArtifacts
		airData.SignalInterests = append(airData.SignalInterests, extractSignalInterestsFromReports(engineReports)...)
		if airData.Extensions == nil {
			airData.Extensions = map[string]any{}
		}
		airData.Extensions["agent_reports"] = reportRefs
	}

	prov := provenance.Build(&enrichedSrc, compilerMode, engineConfigs, "", nil)
	canonical, err := air.CanonicalJSON(airData)
	if err != nil {
		return nil, fmt.Errorf("computing canonical AIR: %w", err)
	}
	irHash := air.HashBytes(canonical)
	outputHashes := computeOutputHashes(airData, irHash)
	prov = provenance.Build(&enrichedSrc, compilerMode, engineConfigs, irHash, outputHashes)
	if len(engineNotes) > 0 {
		prov.Notes = strings.TrimSpace(prov.Notes + "\n\nEngine notes:\n- " + strings.Join(engineNotes, "\n- "))
	}
	airData.Provenance = prov

	reason := reasoning.Build(&enrichedSrc, compilerMode, normRules, warnings)
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

func activeEngineConfigs(configs []air.EngineConfig) []air.EngineConfig {
	out := make([]air.EngineConfig, 0, len(configs))
	for _, cfg := range configs {
		if cfg.Enabled != nil && !*cfg.Enabled {
			continue
		}
		out = append(out, cfg)
	}
	return out
}

func resolveOriginalMode(m *air.Manifest, override string) string {
	if override != "" {
		return override
	}
	if m.Compiler.Mode != "" {
		return m.Compiler.Mode
	}
	return "manual"
}
