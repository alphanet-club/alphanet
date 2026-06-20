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
