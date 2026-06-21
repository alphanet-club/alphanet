package provenance

import (
	"sort"

	"github.com/alphanet/rules-compiler/internal/air"
)

// Build creates a Provenance struct from the source context and compiler mode.
func Build(src *air.SourceContext, mode string, engines []air.EngineConfig, irHash string, outputHashes map[string]string) *air.Provenance {
	prov := &air.Provenance{
		StrategyName:    src.Manifest.Name,
		StrategyID:      src.Manifest.StrategyID,
		SpecVersion:     src.Manifest.SpecVersion,
		CompilerVersion: "v0.1.0",
		GeneratedAt:     src.GeneratedAt,
		CompilerMode:    mode,
		EnsembleMethod:  src.Manifest.Compiler.EnsembleMethod,
		SourceFiles:     sourceFiles(src),
		SourceHashes:    air.HashSource(src),
		Outputs:         outputHashes,
		IRSHA256:        irHash,
		Notes:           src.Manifest.Compiler.Notes,
	}

	// Training window
	if src.Manifest.Compiler.TrainingWindow != nil {
		tw := *src.Manifest.Compiler.TrainingWindow
		prov.TrainingWindow = &tw
	}

	// Engines
	if len(engines) > 0 {
		prov.Engines = engines
	} else if len(src.Manifest.Compiler.Engines) > 0 {
		prov.Engines = src.Manifest.Compiler.Engines
	}

	// Sort source files for determinism
	if prov.SourceFiles != nil {
		sort.Strings(prov.SourceFiles)
	}

	return prov
}

func sourceFiles(src *air.SourceContext) []string {
	files := []string{"manifest.json", "strategy.md", "rules.json"}
	if src.SignalsRaw != nil {
		files = append(files, "signals.json")
	}
	if src.SignalInterestsRaw != nil {
		files = append(files, "signal_interests.json")
	}
	return files
}
