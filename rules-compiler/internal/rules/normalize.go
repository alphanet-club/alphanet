package rules

import (
	"fmt"

	"github.com/alphanet/rules-compiler/internal/air"
)

// NormalizeRules validates and normalizes a set of rules.
// It checks for duplicate rule IDs, validates required fields, and sets defaults.
func NormalizeRules(rules []air.Rule) ([]air.Rule, error) {
	var errs []string
	seen := make(map[string]int)

	for i, r := range rules {
		if r.RuleID == "" {
			errs = append(errs, fmt.Sprintf("rules[%d]: rule_id is required", i))
			continue
		}
		if j, ok := seen[r.RuleID]; ok {
			errs = append(errs, fmt.Sprintf("rules[%d]: duplicate rule_id %q (first at index %d)", i, r.RuleID, j))
			continue
		}
		seen[r.RuleID] = i

		// Default enabled to true
		if !r.Enabled && r.RuleID != "" {
			// json.Unmarshal defaults bools to false; we need to detect unset vs explicitly false.
			// We'll set default in the compiler pipeline instead.
		}

		if r.Layer == "" {
			errs = append(errs, fmt.Sprintf("rule %q: layer is required", r.RuleID))
		}
		if r.Priority <= 0 {
			errs = append(errs, fmt.Sprintf("rule %q: priority must be > 0", r.RuleID))
		}
		if len(r.Then) == 0 {
			errs = append(errs, fmt.Sprintf("rule %q: at least one action in 'then' is required", r.RuleID))
		}
	}

	if len(errs) > 0 {
		return rules, fmt.Errorf("rule validation errors: %s", joinErrors(errs))
	}

	return rules, nil
}

func joinErrors(errs []string) string {
	result := ""
	for i, e := range errs {
		if i > 0 {
			result += "; "
		}
		result += e
	}
	return result
}
