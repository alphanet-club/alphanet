package strategy

import (
	"github.com/alphanet/rules-compiler/internal/air"
)

// SourceContext holds all raw source data loaded from a strategy folder.
type SourceContext struct {
	Manifest   *air.Manifest
	StrategyMD string
	Rules      []air.Rule

	// Raw bytes for hashing
	ManifestRaw []byte
	StrategyRaw []byte
	RulesRaw    []byte
}
