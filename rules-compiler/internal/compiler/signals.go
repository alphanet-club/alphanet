package compiler

import (
	"strings"

	"github.com/alphanet/rules-compiler/internal/air"
)

// autoGenerateSignals walks rules' conditions and selection policy ranking criteria
// to find all referenced signal IDs, then creates Signal definitions for any that
// are not already defined. This implements the manual mode requirement:
// "All signals referenced by rules must be defined by the compiler."
func autoGenerateSignals(signals []air.Signal, rules []air.Rule, selectionPolicy *air.SelectionPolicy) ([]air.Signal, []string) {
	var warnings []string

	// Build set of existing signal IDs
	existing := make(map[string]bool)
	for _, s := range signals {
		existing[s.SignalID] = true
	}

	// Collect all referenced signal IDs from rules and selection policy
	referenced := make(map[string]bool)
	for _, rule := range rules {
		collectSignalRefs(&rule.When, referenced)
	}

	// Also collect from selection policy ranking criteria
	if selectionPolicy != nil {
		for _, basketPol := range selectionPolicy.Baskets {
			for _, rank := range basketPol.Ranking {
				if rank.Signal != "" {
					referenced[rank.Signal] = true
				}
			}
		}
	}

	// Generate signals for referenced IDs that don't exist yet
	generated := make([]air.Signal, 0)
	for signalID := range referenced {
		if existing[signalID] {
			continue
		}
		sig := inferSignal(signalID)
		generated = append(generated, sig)
		existing[signalID] = true
		warnings = append(warnings, "auto-generated signal: "+signalID)
	}

	if len(generated) > 0 {
		signals = append(signals, generated...)
	}

	return signals, warnings
}

// collectSignalRefs recursively walks a condition tree to find all signal references.
func collectSignalRefs(cond *air.Condition, refs map[string]bool) {
	if cond == nil {
		return
	}
	if cond.Signal != "" {
		refs[cond.Signal] = true
	}
	for i := range cond.All {
		collectSignalRefs(&cond.All[i], refs)
	}
	for i := range cond.Any {
		collectSignalRefs(&cond.Any[i], refs)
	}
	if cond.Not != nil {
		collectSignalRefs(cond.Not, refs)
	}
}

// inferSignal creates a Signal struct with reasonable defaults based on naming patterns.
func inferSignal(signalID string) air.Signal {
	sig := air.Signal{
		SignalID:  signalID,
		Family:    inferFamily(signalID),
		Type:      inferType(signalID),
		Window:    inferWindow(signalID),
		Unit:      inferUnit(signalID),
		Transform: inferTransform(signalID),
	}

	// Default family if not matched
	if sig.Family == "" {
		sig.Family = "custom"
	}

	return sig
}

func inferFamily(signalID string) string {
	lower := strings.ToLower(signalID)

	if strings.Contains(lower, "wti") || strings.Contains(lower, "crude") ||
		strings.Contains(lower, "brent") || strings.Contains(lower, "oil") {
		return "macro"
	}
	if strings.Contains(lower, "ust") || strings.Contains(lower, "treasury") ||
		strings.Contains(lower, "yield") || strings.Contains(lower, "rate") ||
		strings.Contains(lower, "bond") || strings.Contains(lower, "fed") ||
		strings.Contains(lower, "libor") || strings.Contains(lower, "spread") {
		return "macro"
	}
	if strings.Contains(lower, "sentiment") || strings.Contains(lower, "news") ||
		strings.Contains(lower, "twitter") || strings.Contains(lower, "social") {
		return "sentiment"
	}
	if strings.Contains(lower, "volatility") || strings.Contains(lower, "vix") ||
		strings.Contains(lower, "vxvix") || strings.Contains(lower, "vxn") {
		return "volatility"
	}
	if strings.Contains(lower, "volume") || strings.Contains(lower, "flow") ||
		strings.Contains(lower, "money") || strings.Contains(lower, "inflow") {
		return "market_flow"
	}
	if strings.Contains(lower, "gdp") || strings.Contains(lower, "cpi") ||
		strings.Contains(lower, "unemployment") || strings.Contains(lower, "inflation") ||
		strings.Contains(lower, "pmi") || strings.Contains(lower, "industrial") {
		return "global_macro"
	}
	if strings.Contains(lower, "pe") || strings.Contains(lower, "pb") ||
		strings.Contains(lower, "dividend") || strings.Contains(lower, "earnings") ||
		strings.Contains(lower, "valuation") || strings.Contains(lower, "book") {
		return "valuation"
	}
	if strings.Contains(lower, "relative_strength") || strings.Contains(lower, "momentum") ||
		strings.Contains(lower, "strength") {
		return "custom"
	}

	return "custom"
}

func inferType(signalID string) string {
	lower := strings.ToLower(signalID)

	if strings.Contains(lower, "change_") || strings.Contains(lower, "_chg") {
		return "percent_change"
	}
	if strings.Contains(lower, "sentiment") {
		return "sentiment_score"
	}
	if strings.Contains(lower, "volatility") || strings.Contains(lower, "vol") {
		return "volatility"
	}
	if strings.Contains(lower, "relative_strength") || strings.Contains(lower, "strength") {
		return "ranking"
	}
	if strings.Contains(lower, "z_score") || strings.Contains(lower, "zscore") {
		return "z_score"
	}
	if strings.Contains(lower, "moving_average") || strings.Contains(lower, "sma") ||
		strings.Contains(lower, "ema") {
		return "moving_average"
	}
	if strings.Contains(lower, "correlation") || strings.Contains(lower, "corr") {
		return "correlation"
	}

	return "level"
}

func inferWindow(signalID string) string {
	// Look for patterns like 20d, 7d, 60d, 3m, 1y
	parts := strings.Split(signalID, "_")
	for _, part := range parts {
		if len(part) >= 2 {
			hasDigit := false
			hasUnit := false
			for _, c := range part {
				if c >= '0' && c <= '9' {
					hasDigit = true
				}
			}
			lastChar := part[len(part)-1]
			if lastChar == 'd' || lastChar == 'w' || lastChar == 'm' || lastChar == 'q' || lastChar == 'y' {
				hasUnit = true
			}
			if hasDigit && hasUnit {
				return part
			}
		}
	}
	return ""
}

func inferUnit(signalID string) string {
	lower := strings.ToLower(signalID)

	if strings.Contains(lower, "percent") || strings.Contains(lower, "pct") {
		return "percent"
	}
	if strings.Contains(lower, "bps") || strings.Contains(lower, "basis") {
		return "basis_points"
	}
	if strings.Contains(lower, "dollar") || strings.Contains(lower, "usd") {
		return "dollars"
	}

	// Default unit based on family
	if inferFamily(signalID) == "volatility" {
		return "percent"
	}
	if inferFamily(signalID) == "sentiment" {
		return "score"
	}

	return ""
}

func inferTransform(signalID string) string {
	lower := strings.ToLower(signalID)

	if strings.Contains(lower, "change_") || strings.Contains(lower, "_chg") {
		if strings.Contains(lower, "pct") || strings.Contains(lower, "percent") {
			return "percent_change"
		}
		return "change"
	}
	if strings.Contains(lower, "sentiment") {
		return "sentiment_score"
	}
	if strings.Contains(lower, "volatility") || strings.Contains(lower, "vol") {
		return "volatility"
	}
	if strings.Contains(lower, "relative_strength") || strings.Contains(lower, "strength") {
		return "rank"
	}
	if strings.Contains(lower, "z_score") || strings.Contains(lower, "zscore") {
		return "z_score"
	}
	if strings.Contains(lower, "moving_average") || strings.Contains(lower, "sma") ||
		strings.Contains(lower, "ema") {
		return "moving_average"
	}
	if strings.Contains(lower, "correlation") || strings.Contains(lower, "corr") {
		return "correlation"
	}

	return "level"
}
