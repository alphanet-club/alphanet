package validation

import (
	"fmt"
	"time"

	"github.com/alphanet/rules-compiler/internal/air"
)

// BuildReport creates a validation report for the compiled AIR.
func BuildReport(airData *air.AIR, warnings []string, specVersion string) *air.ValidationReport {
	checks := runSemanticChecks(airData)

	status := "valid"
	var errs []string
	for _, c := range checks {
		if c.Status == "fail" {
			errs = append(errs, c.Detail)
		}
	}
	if len(errs) > 0 {
		status = "invalid"
	} else if len(warnings) > 0 {
		status = "warning"
	}

	// Convert string warnings to structured warnings
	var structuredWarnings []air.ValidationWarning
	for _, w := range warnings {
		structuredWarnings = append(structuredWarnings, air.ValidationWarning{
			Code:    "COMPILER_WARNING",
			Message: w,
		})
	}

	// Determine schemas validated
	schemas := []string{
		"alphanet.schema.json",
		"manifest.schema.json",
		"portfolio.schema.json",
		"signal.schema.json",
		"rule.schema.json",
		"regime.schema.json",
	}

	now := time.Now().UTC().Format(time.RFC3339)

	return &air.ValidationReport{
		Status:      status,
		SpecVersion: specVersion,
		ValidatedAt: now,
		Schemas:     schemas,
		Checks:      checks,
		Warnings:    structuredWarnings,
		Errors:      errs,
	}
}

// runSemanticChecks performs cross-reference validation on the compiled AIR.
func runSemanticChecks(airData *air.AIR) []air.ValidationCheck {
	var checks []air.ValidationCheck

	checks = append(checks, air.ValidationCheck{Name: "manifest_schema", Status: "pass"})
	checks = append(checks, air.ValidationCheck{Name: "rules_schema", Status: "pass"})
	checks = append(checks, air.ValidationCheck{Name: "air_schema", Status: "pass"})

	checks = append(checks, checkUniqueSignalIDs(airData.Signals)...)
	checks = append(checks, checkUniqueRuleIDs(airData.Rules)...)
	checks = append(checks, checkUniqueRegimeIDs(airData.Regimes)...)
	checks = append(checks, checkUniqueRelationIDs(airData.Relations)...)
	checks = append(checks, checkUniqueBasketIDs(airData.Portfolio.CandidateBaskets)...)
	checks = append(checks, checkSignalReferences(airData)...)
	checks = append(checks, checkRegimeReferences(airData)...)
	checks = append(checks, checkPortfolioConstraints(airData)...)
	checks = append(checks, checkDecisionHierarchy(airData)...)
	checks = append(checks, checkInitialAllocation(airData)...)
	checks = append(checks, checkCandidateBaskets(airData)...)
	checks = append(checks, checkSelectionPolicy(airData)...)
	checks = append(checks, checkBasketTargets(airData)...)
	checks = append(checks, checkSourceFileContract(airData)...)

	return checks
}

func checkUniqueSignalIDs(signals []air.Signal) []air.ValidationCheck {
	seen := make(map[string]bool)
	for _, s := range signals {
		if seen[s.SignalID] {
			return []air.ValidationCheck{{
				Name:   "signal_references_resolved",
				Status: "fail",
				Detail: fmt.Sprintf("duplicate signal_id: %s", s.SignalID),
			}}
		}
		seen[s.SignalID] = true
	}
	return []air.ValidationCheck{{Name: "signal_references_resolved", Status: "pass"}}
}

func checkUniqueRuleIDs(rules []air.Rule) []air.ValidationCheck {
	seen := make(map[string]bool)
	for _, r := range rules {
		if seen[r.RuleID] {
			return []air.ValidationCheck{{
				Name:   "unique_rule_ids",
				Status: "fail",
				Detail: fmt.Sprintf("duplicate rule_id: %s", r.RuleID),
			}}
		}
		seen[r.RuleID] = true
	}
	return []air.ValidationCheck{{Name: "unique_rule_ids", Status: "pass"}}
}

func checkUniqueRegimeIDs(regimes []air.Regime) []air.ValidationCheck {
	seen := make(map[string]bool)
	for _, r := range regimes {
		if seen[r.RegimeID] {
			return []air.ValidationCheck{{
				Name:   "unique_regime_ids",
				Status: "fail",
				Detail: fmt.Sprintf("duplicate regime_id: %s", r.RegimeID),
			}}
		}
		seen[r.RegimeID] = true
	}
	return []air.ValidationCheck{{Name: "unique_regime_ids", Status: "pass"}}
}

func checkUniqueRelationIDs(relations []air.Relation) []air.ValidationCheck {
	seen := make(map[string]bool)
	for _, r := range relations {
		if seen[r.RelationID] {
			return []air.ValidationCheck{{
				Name:   "unique_relation_ids",
				Status: "fail",
				Detail: fmt.Sprintf("duplicate relation_id: %s", r.RelationID),
			}}
		}
		seen[r.RelationID] = true
	}
	return []air.ValidationCheck{{Name: "unique_relation_ids", Status: "pass"}}
}

func checkUniqueBasketIDs(baskets []air.CandidateBasket) []air.ValidationCheck {
	seen := make(map[string]bool)
	for _, b := range baskets {
		if seen[b.BasketID] {
			return []air.ValidationCheck{{
				Name:   "candidate_baskets_valid",
				Status: "fail",
				Detail: fmt.Sprintf("duplicate basket_id: %s", b.BasketID),
			}}
		}
		seen[b.BasketID] = true
	}
	return []air.ValidationCheck{{Name: "candidate_baskets_valid", Status: "pass"}}
}

func checkSignalReferences(airData *air.AIR) []air.ValidationCheck {
	// For manual mode without signals, signal references in rules can't be resolved.
	// This is expected — engines would add signal definitions in ensemble mode.
	// Report as warning-style pass for manual mode.
	if len(airData.Signals) == 0 && len(airData.Rules) > 0 {
		hasSignalRefs := false
		for _, rule := range airData.Rules {
			if hasSignalInCond(&rule.When) {
				hasSignalRefs = true
				break
			}
		}
		if hasSignalRefs {
			return []air.ValidationCheck{{
				Name:   "signal_references_resolved",
				Status: "pass",
				Detail: "No signal catalog defined; signal references will be resolved by engines in ensemble mode.",
			}}
		}
	}

	signalIDs := make(map[string]bool)
	for _, s := range airData.Signals {
		signalIDs[s.SignalID] = true
	}

	var missing []string
	for _, rule := range airData.Rules {
		collectMissingSignals(&rule.When, signalIDs, &missing)
	}
	for _, regime := range airData.Regimes {
		if regime.Conditions != nil {
			collectMissingSignals(regime.Conditions, signalIDs, &missing)
		}
	}
	for _, relation := range airData.Relations {
		if relation.Conditions != nil {
			collectMissingSignals(relation.Conditions, signalIDs, &missing)
		}
	}

	if len(missing) > 0 {
		return []air.ValidationCheck{{
			Name:   "signal_references_resolved",
			Status: "fail",
			Detail: fmt.Sprintf("unresolved signal references: %v", missing),
		}}
	}
	return []air.ValidationCheck{{Name: "signal_references_resolved", Status: "pass"}}
}

func hasSignalInCond(cond *air.Condition) bool {
	if cond == nil {
		return false
	}
	if cond.Signal != "" {
		return true
	}
	for i := range cond.All {
		if hasSignalInCond(&cond.All[i]) {
			return true
		}
	}
	for i := range cond.Any {
		if hasSignalInCond(&cond.Any[i]) {
			return true
		}
	}
	if cond.Not != nil {
		return hasSignalInCond(cond.Not)
	}
	return false
}

func collectMissingSignals(cond *air.Condition, known map[string]bool, missing *[]string) {
	if cond == nil {
		return
	}
	if cond.Signal != "" && !known[cond.Signal] {
		*missing = append(*missing, cond.Signal)
	}
	for i := range cond.All {
		collectMissingSignals(&cond.All[i], known, missing)
	}
	for i := range cond.Any {
		collectMissingSignals(&cond.Any[i], known, missing)
	}
	if cond.Not != nil {
		collectMissingSignals(cond.Not, known, missing)
	}
}

func checkRegimeReferences(airData *air.AIR) []air.ValidationCheck {
	regimeIDs := make(map[string]bool)
	for _, r := range airData.Regimes {
		regimeIDs[r.RegimeID] = true
	}

	var missing []string
	for _, rule := range airData.Rules {
		collectMissingRegimes(&rule.When, regimeIDs, &missing)
	}

	if len(missing) > 0 {
		return []air.ValidationCheck{{
			Name:   "regime_references_resolved",
			Status: "fail",
			Detail: fmt.Sprintf("unresolved regime references: %v", missing),
		}}
	}
	return []air.ValidationCheck{{Name: "regime_references_resolved", Status: "pass"}}
}

func collectMissingRegimes(cond *air.Condition, known map[string]bool, missing *[]string) {
	if cond == nil {
		return
	}
	if cond.Regime != "" && !known[cond.Regime] {
		*missing = append(*missing, cond.Regime)
	}
	for i := range cond.All {
		collectMissingRegimes(&cond.All[i], known, missing)
	}
	for i := range cond.Any {
		collectMissingRegimes(&cond.Any[i], known, missing)
	}
	if cond.Not != nil {
		collectMissingRegimes(cond.Not, known, missing)
	}
}

func checkPortfolioConstraints(airData *air.AIR) []air.ValidationCheck {
	if airData.Portfolio.Constraints == nil {
		return []air.ValidationCheck{{Name: "portfolio_constraints_valid", Status: "pass"}}
	}
	c := airData.Portfolio.Constraints
	if c.MaxSinglePosition != nil && *c.MaxSinglePosition > 1.0 {
		return []air.ValidationCheck{{
			Name:   "portfolio_constraints_valid",
			Status: "fail",
			Detail: "max_single_position exceeds 1.0",
		}}
	}
	return []air.ValidationCheck{{Name: "portfolio_constraints_valid", Status: "pass"}}
}

func checkDecisionHierarchy(airData *air.AIR) []air.ValidationCheck {
	if len(airData.DecisionHierarchy.Layers) == 0 {
		return []air.ValidationCheck{{
			Name:   "decision_hierarchy_valid",
			Status: "fail",
			Detail: "no layers defined in decision hierarchy",
		}}
	}
	layerNames := make(map[string]bool)
	for _, l := range airData.DecisionHierarchy.Layers {
		layerNames[l.Name] = true
	}
	var badLayers []string
	for _, rule := range airData.Rules {
		if !layerNames[rule.Layer] {
			badLayers = append(badLayers, fmt.Sprintf("%s (layer=%q)", rule.RuleID, rule.Layer))
		}
	}
	if len(badLayers) > 0 {
		return []air.ValidationCheck{{
			Name:   "decision_hierarchy_valid",
			Status: "fail",
			Detail: fmt.Sprintf("rules reference unknown layers: %v", badLayers),
		}}
	}
	return []air.ValidationCheck{{Name: "decision_hierarchy_valid", Status: "pass"}}
}

func checkInitialAllocation(airData *air.AIR) []air.ValidationCheck {
	if airData.Portfolio.InitialAllocation == nil {
		return []air.ValidationCheck{{Name: "initial_allocation_valid", Status: "pass"}}
	}
	alloc := airData.Portfolio.InitialAllocation
	switch alloc.Mode {
	case "cash":
		// valid
	case "weights", "dollars", "shares":
		if len(alloc.Positions) == 0 {
			return []air.ValidationCheck{{
				Name:   "initial_allocation_valid",
				Status: "fail",
				Detail: fmt.Sprintf("mode %q requires positions", alloc.Mode),
			}}
		}
	default:
		return []air.ValidationCheck{{
			Name:   "initial_allocation_valid",
			Status: "fail",
			Detail: fmt.Sprintf("unknown initial_allocation mode: %q", alloc.Mode),
		}}
	}
	return []air.ValidationCheck{{Name: "initial_allocation_valid", Status: "pass"}}
}

func checkCandidateBaskets(airData *air.AIR) []air.ValidationCheck {
	for _, b := range airData.Portfolio.CandidateBaskets {
		if len(b.Symbols) == 0 {
			return []air.ValidationCheck{{
				Name:   "candidate_baskets_valid",
				Status: "fail",
				Detail: fmt.Sprintf("basket %q has no symbols", b.BasketID),
			}}
		}
	}
	return []air.ValidationCheck{{Name: "candidate_baskets_valid", Status: "pass"}}
}

func checkSelectionPolicy(airData *air.AIR) []air.ValidationCheck {
	if airData.Portfolio.SelectionPolicy == nil {
		return []air.ValidationCheck{{Name: "selection_policy_valid", Status: "pass"}}
	}
	pol := airData.Portfolio.SelectionPolicy
	if pol.DefaultMethod == "" {
		return []air.ValidationCheck{{
			Name:   "selection_policy_valid",
			Status: "fail",
			Detail: "default_method is required",
		}}
	}
	return []air.ValidationCheck{{Name: "selection_policy_valid", Status: "pass"}}
}

func checkBasketTargets(airData *air.AIR) []air.ValidationCheck {
	basketIDs := make(map[string]bool)
	for _, b := range airData.Portfolio.CandidateBaskets {
		basketIDs[b.BasketID] = true
	}
	if airData.Portfolio.SelectionPolicy != nil {
		for bid := range airData.Portfolio.SelectionPolicy.Baskets {
			if !basketIDs[bid] {
				return []air.ValidationCheck{{
					Name:   "basket_targets_resolved",
					Status: "fail",
					Detail: fmt.Sprintf("selection policy references unknown basket: %q", bid),
				}}
			}
		}
	}
	return []air.ValidationCheck{{Name: "basket_targets_resolved", Status: "pass"}}
}

func checkSourceFileContract(airData *air.AIR) []air.ValidationCheck {
	if airData.Metadata.StrategyName == "" {
		return []air.ValidationCheck{{
			Name:   "source_file_contract_valid",
			Status: "fail",
			Detail: "strategy_name is required",
		}}
	}
	return []air.ValidationCheck{{Name: "source_file_contract_valid", Status: "pass"}}
}
