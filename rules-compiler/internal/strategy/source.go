package strategy

import (
	"github.com/alphanet/rules-compiler/internal/air"
)

// SourceContext holds all raw source data loaded from a strategy folder.
type SourceContext struct {
	Manifest        *air.Manifest
	StrategyMD      string
	Rules           []air.Rule
	Signals         []air.Signal
	SignalInterests []air.SignalInterest

	// Raw bytes for hashing
	ManifestRaw        []byte
	StrategyRaw        []byte
	RulesRaw           []byte
	SignalsRaw         []byte
	SignalInterestsRaw []byte
}
