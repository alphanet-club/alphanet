package portfolio

import (
	"fmt"
	"math"

	"github.com/alphanet/rules-compiler/internal/air"
)

// NormalizePortfolio validates and normalizes the portfolio configuration from the manifest.
// It returns the AIR portfolio and any warnings.
func NormalizePortfolio(m *air.Manifest) (*air.AIRPortfolio, []string) {
	var warnings []string

	p := &air.AIRPortfolio{}

	// 1. Base currency
	p.BaseCurrency = m.Portfolio.BaseCurrency
	if p.BaseCurrency == "" {
		p.BaseCurrency = "USD"
	}

	// 2. Starting cash
	p.StartingCash = m.Portfolio.StartingCash
	if p.StartingCash <= 0 {
		p.StartingCash = 100000
	}

	// 3. Initial allocation
	if m.Portfolio.InitialAllocation != nil {
		alloc := *m.Portfolio.InitialAllocation
		normalized := normalizeInitialAllocation(&alloc, p.StartingCash, &warnings)
		p.InitialAllocation = &normalized
	} else {
		// Default: 100% cash
		p.InitialAllocation = &air.InitialAllocation{
			Mode:      "cash",
			Positions: nil,
		}
	}

	// 4. Candidate baskets
	if len(m.Portfolio.CandidateBaskets) > 0 {
		baskets, basketWarnings := normalizeBaskets(m.Portfolio.CandidateBaskets)
		p.CandidateBaskets = baskets
		warnings = append(warnings, basketWarnings...)
	}

	// 5. Selection policy
	if m.Portfolio.SelectionPolicy != nil {
		pol, polWarnings := normalizeSelectionPolicy(*m.Portfolio.SelectionPolicy, p.CandidateBaskets)
		p.SelectionPolicy = &pol
		warnings = append(warnings, polWarnings...)
	}

	// 6. Targets
	p.Targets = m.Portfolio.Targets
	if p.Targets == nil {
		p.Targets = map[string]float64{}
	}

	// 7. Soft targets
	p.SoftTargets = m.Portfolio.SoftTargets
	if p.SoftTargets == nil {
		p.SoftTargets = []air.SoftTarget{}
	}

	// 8. Constraints
	if m.Portfolio.Constraints != nil {
		constraints := *m.Portfolio.Constraints
		normalizeConstraints(&constraints, &warnings)
		p.Constraints = &constraints
	}

	// 9. Risk budgets
	p.RiskBudgets = m.Portfolio.RiskBudgets

	// 10. Rebalance
	if m.Portfolio.Rebalance != nil {
		rb := *m.Portfolio.Rebalance
		if rb.Frequency == "" {
			rb.Frequency = "daily"
		}
		if rb.Threshold == 0 {
			rb.Threshold = 0.05
		}
		p.Rebalance = &rb
	} else {
		p.Rebalance = &air.RebalancePolicy{
			Frequency: "daily",
			Threshold: 0.05,
		}
	}

	return p, warnings
}

func normalizeInitialAllocation(alloc *air.InitialAllocation, startingCash float64, warnings *[]string) air.InitialAllocation {
	switch alloc.Mode {
	case "", "cash":
		return air.InitialAllocation{Mode: "cash"}
	case "weights":
		total := 0.0
		for _, pos := range alloc.Positions {
			total += pos.Weight
		}
		if math.Abs(total-1.0) > 0.01 {
			*warnings = append(*warnings, fmt.Sprintf("initial_allocation: weights sum to %.4f, expected ~1.0", total))
		}
		return *alloc
	case "dollars":
		total := 0.0
		for _, pos := range alloc.Positions {
			total += pos.Amount
		}
		if total > startingCash {
			*warnings = append(*warnings, fmt.Sprintf("initial_allocation: dollar amounts (%.2f) exceed starting_cash (%.2f)", total, startingCash))
		}
		return *alloc
	default:
		*warnings = append(*warnings, fmt.Sprintf("initial_allocation: unknown mode %q, defaulting to cash", alloc.Mode))
		return air.InitialAllocation{Mode: "cash"}
	}
}

func normalizeConstraints(c *air.PortfolioConstraints, warnings *[]string) {
	if c.MaxLeverage == nil {
		def := 1.0
		c.MaxLeverage = &def
	}
	if c.MaxSinglePosition != nil && *c.MaxSinglePosition > 1.0 {
		*warnings = append(*warnings, "constraints: max_single_position > 1.0, capping at 1.0")
		cap := 1.0
		c.MaxSinglePosition = &cap
	}
}
