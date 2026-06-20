# AlphaNet v0.1 Changelog

## v0.1.0 — Draft Foundation

Initial AlphaNet specification draft.

### Added

- Defined AlphaNet as a market intelligence runtime for strategies, agents, and portfolio management.
- Introduced AlphaNet Intermediate Representation, or AIR.
- Standardized the compiled strategy artifact as `compiled/strategy.ir.json`.
- Defined the separation between:
  - strategy authoring
  - rules compilation
  - deterministic backtesting
- Defined strategy source files:
  - `manifest.json`
  - `strategy.md`
  - `rules.json`
- Defined compiler outputs:
  - `compiled/strategy.ir.json`
  - `compiled/provenance.json`
  - `compiled/reasoning.md`
  - `compiled/validation-report.json`
- Defined deterministic backtester expectations.
- Defined versioning requirements for:
  - spec version
  - compiler version
  - backtester version
  - data version
- Added support for local and hosted compilation.
- Added support for multiple compiler engines per strategy.
- Added example engine references:
  - `TauricResearch/TradingAgents` version `0.2.5`
  - `virattt/ai-hedge-fund` version `2026.6.17`
- Added compiler training windows.
- Added portfolio targets and constraints.
- Added hard constraints and soft targets.
- Added decision hierarchy with layer and rule priorities.
- Added signal families:
  - valuation
  - macro
  - global macro
  - fed
  - shipping
  - market flow
  - events
  - sentiment
  - volatility
  - portfolio
  - custom
- Added relation and regime concepts.
- Added decision trace concept for explainable backtesting.
- Added initial repository hierarchy.

### Schemas Added

- `alphanet.schema.json`
- `manifest.schema.json`
- `portfolio.schema.json`
- `signal.schema.json`
- `rule.schema.json`
- `regime.schema.json`

### Notes

This version is intended as the first stable design target for repository bootstrapping. It is not yet a production trading system or financial advice framework.


## Unreleased — Portfolio Initialization and Candidate Baskets

### Added

- Added `portfolio.initial_allocation` for explicit starting holdings.
- Added initial allocation modes: `cash`, `weights`, `dollars`, and `shares`.
- Added `portfolio.candidate_baskets` for securities the strategy may buy or rotate into later.
- Added `portfolio.selection_policy` for deterministic basket selection.
- Added basket-targeting rule actions using `target_type`.
- Added rotation-style actions such as `rotate`, `allocate_to_basket`, `reduce_basket`, `rank_and_select`, and `replace_low_ranked`.

### Updated

- Updated portfolio schema to distinguish starting holdings from desired portfolio targets.
- Updated AIR schema to carry initial allocations and candidate baskets into the compiled artifact.
- Updated rule schema so actions can target symbols, baskets, sectors, themes, cash, or the whole portfolio.
- Updated documentation for runtime initialization, rule targeting, and portfolio basket selection.
- Updated the oil/rates/growth-tech example strategy with explicit initial allocation and candidate baskets.

### Clarified

- Clarified source-file responsibilities for `manifest.json`, `strategy.md`, `rules.json`, and `compiled/strategy.ir.json`.
- Clarified that portfolio initialization, candidate baskets, selection policy, targets, constraints, and risk budgets are portfolio configuration carried from `manifest.json` into compiled AIR.
- Clarified that `rules.json` contains user-authored seed rules.
- Updated `rule.schema.json` to explicitly support basket-targeted and rotation actions.

---

## Sampling Policy Update

Added sampling policy support for AlphaNet v0.1.

### Added

- `training_window.sampling`
- `training_window.include_ranges`
- `training_window.exclude_ranges`
- `backtest.decision_sampling`
- `backtest.valuation_frequency`
- `backtest.include_ranges`
- `backtest.exclude_ranges`
- AIR `execution.decision_sampling`
- AIR `execution.valuation_frequency`

Supported sampling frequencies:

```text
hourly
daily
weekly
biweekly
monthly
quarterly
semiannual
annual
```

This allows large training windows and backtest ranges to be evaluated at lower decision cadence while preserving deterministic behavior.
