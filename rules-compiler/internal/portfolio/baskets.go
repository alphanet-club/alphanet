package portfolio

import (
	"fmt"

	"github.com/alphanet/rules-compiler/internal/air"
)

// normalizeBaskets validates candidate baskets and returns normalized versions.
func normalizeBaskets(baskets []air.CandidateBasket) ([]air.CandidateBasket, []string) {
	var warnings []string
	seen := make(map[string]int)

	for i := range baskets {
		b := &baskets[i]

		if b.BasketID == "" {
			warnings = append(warnings, fmt.Sprintf("candidate_baskets[%d]: basket_id is required, skipping", i))
			continue
		}

		if j, ok := seen[b.BasketID]; ok {
			warnings = append(warnings, fmt.Sprintf("candidate_baskets: duplicate basket_id %q at indices %d and %d, keeping first", b.BasketID, j, i))
			continue
		}
		seen[b.BasketID] = i

		if len(b.Symbols) == 0 {
			warnings = append(warnings, fmt.Sprintf("candidate_basket %q: no symbols defined", b.BasketID))
		}

		// Ensure max_weight is <= 1.0
		if b.MaxWeight != nil && *b.MaxWeight > 1.0 {
			warnings = append(warnings, fmt.Sprintf("candidate_basket %q: max_weight > 1.0, capping at 1.0", b.BasketID))
			cap := 1.0
			b.MaxWeight = &cap
		}
	}

	return baskets, warnings
}

// normalizeSelectionPolicy validates and normalizes the selection policy.
func normalizeSelectionPolicy(policy air.SelectionPolicy, baskets []air.CandidateBasket) (air.SelectionPolicy, []string) {
	var warnings []string

	if policy.DefaultMethod == "" {
		policy.DefaultMethod = "equal_weight"
	}

	// Build set of valid basket IDs
	basketIDs := make(map[string]bool)
	for _, b := range baskets {
		basketIDs[b.BasketID] = true
	}

	// Validate basket references
	for bid := range policy.Baskets {
		if !basketIDs[bid] {
			warnings = append(warnings, fmt.Sprintf("selection_policy: basket %q not found in candidate_baskets", bid))
		}
	}

	if policy.RebalanceThreshold == 0 && policy.DefaultMethod != "" {
		policy.RebalanceThreshold = 0.05
	}

	return policy, warnings
}
