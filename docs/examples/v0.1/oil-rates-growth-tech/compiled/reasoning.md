# Compile Reasoning

## Strategy

`Oil Rates Growth Tech`

## Summary

The compiler converted the user-provided strategy description and seed rules into a portfolio-aware AlphaNet Intermediate Representation.

The core compiled belief is:

> This strategy reduces growth technology exposure when oil prices and long-term interest rates rise together.

## Inputs Reviewed

- `manifest.json`
- `strategy.md`
- `rules.json`
- Compiler engine configuration
- Training window definition

## Agent Engines

Configured engines:

- `TauricResearch/TradingAgents` version `0.2.5`
- `priley86/ai-hedge-fund` version `2026.6.17`

This example does not include real agent logs. In a real compile, this file would summarize relevant agent feedback and explain why rules were accepted, modified, or rejected.

## Major Compiler Decisions

### 1. Preserved Core User Intent

The user wanted to reduce growth technology exposure when oil and rates rise together.

The compiler preserved this logic as:

- `wti_change_20d > 10%`
- `ust10y_change_20d > 25 basis points`
- `tight_liquidity` active

### 2. Added Explicit Relation

The compiler added:

```text
oil_rates_negative_for_growth_tech
```

This relation connects the macro drivers to the target assets and theme.

### 3. Added Regime Layer

The compiler added:

```text
tight_liquidity
high_volatility
```

These regimes allow the strategy to reason at a higher level than raw signals.

### 4. Added Portfolio Safety Rule

The compiler added a portfolio safety rule blocking trades that would leave cash below the minimum.

### 5. Added Risk Management Rule

The compiler added a high-volatility risk reduction rule that can override ordinary strategy rules because it belongs to a higher-priority layer.

## Final AIR Output

The final compiled strategy is:

```text
compiled/strategy.ir.json
```

This is the only file required by the backtester.

## Portfolio Initialization Update

The compiled AIR now explicitly separates:

- initial allocation
- candidate baskets
- long-term targets
- hard constraints
- selection policy

This allows the backtester to start from a known portfolio while still allowing later deterministic rotation into candidate baskets.

## Basket Selection Update

The compiler added candidate baskets for:

- growth technology
- defensive equities
- duration
- commodities and energy

A basket rotation rule was added so oil/rates pressure can reduce growth technology exposure and rotate into defensive equities when appropriate.

## Source File Contract

The source strategy keeps portfolio configuration in `manifest.json`, human strategy intent in `strategy.md`, and user-authored seed rules in `rules.json`.

The compiler normalizes those inputs into `compiled/strategy.ir.json`, which is the only strategy artifact required by the backtester.

