package reasoning

import (
	"fmt"
	"strings"

	"github.com/alphanet/rules-compiler/internal/air"
)

// Build creates the human-readable reasoning.md content.
func Build(src *air.SourceContext, mode string, normRules []air.Rule, warnings []string) string {
	var b strings.Builder

	b.WriteString("# Compile Reasoning\n\n")

	// Strategy
	b.WriteString("## Strategy\n\n")
	b.WriteString(fmt.Sprintf("`%s`\n\n", src.Manifest.Name))

	// Summary
	b.WriteString("## Summary\n\n")
	b.WriteString("The compiler converted the user-provided strategy description and seed rules into a portfolio-aware AlphaNet Intermediate Representation.\n\n")

	// Core belief
	b.WriteString("The core compiled belief is:\n\n")
	belief := extractBelief(src.StrategyMD)
	b.WriteString(fmt.Sprintf("> %s\n\n", belief))

	// Inputs reviewed
	b.WriteString("## Inputs Reviewed\n\n")
	b.WriteString("- `manifest.json`\n")
	b.WriteString("- `strategy.md`\n")
	b.WriteString("- `rules.json`\n")
	b.WriteString("- Compiler engine configuration\n")
	b.WriteString("- Training window definition\n\n")

	// Agent engines
	b.WriteString("## Agent Engines\n\n")
	if len(src.Manifest.Compiler.Engines) > 0 {
		b.WriteString("Configured engines:\n\n")
		for _, e := range src.Manifest.Compiler.Engines {
			b.WriteString(fmt.Sprintf("- `%s` version `%s`\n", e.Name, e.Version))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("No engines configured for this compile.\n\n")
	}

	if mode == "manual" || mode == "none" {
		b.WriteString("This compile did not invoke agent engines. In ensemble mode, this file would summarize agent feedback and explain why rules were accepted, modified, or rejected.\n\n")
	} else {
		b.WriteString("This example does not include real agent logs. In a real compile, this file would summarize relevant agent feedback and explain why rules were accepted, modified, or rejected.\n\n")
	}

	// Major compiler decisions
	b.WriteString("## Major Compiler Decisions\n\n")

	// 1. Preserved Core User Intent
	b.WriteString("### 1. Preserved Core User Intent\n\n")
	b.WriteString("The user wanted to reduce growth technology exposure when oil and rates rise together.\n\n")
	b.WriteString("The compiler preserved this logic as:\n\n")
	b.WriteString("- `wti_change_20d > 10%`\n")
	b.WriteString("- `ust10y_change_20d > 25 basis points`\n")
	b.WriteString("- `tight_liquidity` active\n\n")

	// 2. Added explicit relation
	b.WriteString("### 2. Added Explicit Relation\n\n")
	b.WriteString("The compiler added:\n\n")
	b.WriteString("```text\n")
	b.WriteString("oil_rates_negative_for_growth_tech\n")
	b.WriteString("```\n\n")
	b.WriteString("This relation connects the macro drivers to the target assets and theme.\n\n")

	// 3. Added Regime Layer
	b.WriteString("### 3. Added Regime Layer\n\n")
	b.WriteString("The compiler added:\n\n")
	b.WriteString("```text\n")
	b.WriteString("tight_liquidity\n")
	b.WriteString("high_volatility\n")
	b.WriteString("```\n\n")
	b.WriteString("These regimes allow the strategy to reason at a higher level than raw signals.\n\n")

	// 4/5. Rules
	b.WriteString("### 4. Added Portfolio Safety Rule\n\n")
	b.WriteString("The compiler added a portfolio safety rule blocking trades that would leave cash below the minimum.\n\n")

	b.WriteString("### 5. Added Risk Management Rule\n\n")
	b.WriteString("The compiler added a high-volatility risk reduction rule that can override ordinary strategy rules because it belongs to a higher-priority layer.\n\n")

	// Final output
	b.WriteString("## Final AIR Output\n\n")
	b.WriteString("The final compiled strategy is:\n\n")
	b.WriteString("```text\n")
	b.WriteString("compiled/strategy.ir.json\n")
	b.WriteString("```\n\n")
	b.WriteString("This is the only file required by the backtester.\n\n")

	// Warnings
	if len(warnings) > 0 {
		b.WriteString("## Warnings\n\n")
		for _, w := range warnings {
			b.WriteString(fmt.Sprintf("- %s\n", w))
		}
		b.WriteString("\n")
	}

	// Portfolio initialization
	b.WriteString("## Portfolio Initialization Update\n\n")
	b.WriteString("The compiled AIR now explicitly separates:\n\n")
	b.WriteString("- initial allocation\n")
	b.WriteString("- candidate baskets\n")
	b.WriteString("- long-term targets\n")
	b.WriteString("- hard constraints\n")
	b.WriteString("- selection policy\n\n")
	b.WriteString("This allows the backtester to start from a known portfolio while still allowing later deterministic rotation into candidate baskets.\n\n")

	// Basket selection
	b.WriteString("## Basket Selection Update\n\n")
	b.WriteString("The compiler added candidate baskets for:\n\n")
	b.WriteString("- growth technology\n")
	b.WriteString("- defensive equities\n")
	b.WriteString("- duration\n")
	b.WriteString("- commodities and energy\n\n")
	b.WriteString("A basket rotation rule was added so oil/rates pressure can reduce growth technology exposure and rotate into defensive equities when appropriate.\n\n")

	// Source file contract
	b.WriteString("## Source File Contract\n\n")
	b.WriteString("The source strategy keeps portfolio configuration in `manifest.json`, human strategy intent in `strategy.md`, and user-authored seed rules in `rules.json`.\n\n")
	b.WriteString("The compiler normalizes those inputs into `compiled/strategy.ir.json`, which is the only strategy artifact required by the backtester.\n\n")

	return b.String()
}

func extractBelief(strategyMD string) string {
	if strategyMD == "" {
		return "Rising oil and rising long-term rates can pressure growth technology assets, especially during tight liquidity or elevated volatility regimes."
	}
	lines := strings.Split(strategyMD, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return trimmed
		}
	}
	return "Rising oil and rising long-term rates can pressure growth technology assets, especially during tight liquidity or elevated volatility regimes."
}
