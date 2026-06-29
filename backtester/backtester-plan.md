# Backtester Implementation Plan

## Purpose

The `backtester` is a Go program that evaluates AlphaNet Intermediate Representation files against historical market data.

The backtester consumes:

```text
compiled/strategy.ir.json
```

It produces:

```text
backtests/YYYY-MM-DD_YYYY-MM-DD.summary.json
backtests/YYYY-MM-DD_YYYY-MM-DD.equity-curve.csv
backtests/YYYY-MM-DD_YYYY-MM-DD.trades.csv
backtests/YYYY-MM-DD_YYYY-MM-DD.decision-trace.jsonl
```

The backtester must be deterministic.

It must not call LLMs or agent engines.

---

## Core Principle

The backtester evaluates.

It does not reason.

It does not compile.

It does not reinterpret strategy intent.

The only strategy input is AIR.

---

## Source Contract

The backtester reads only:

```text
compiled/strategy.ir.json
```

The backtester should not read:

```text
manifest.json
strategy.md
rules.json
```

`manifest.json` is an authoring-time source file.

`rules.json` is an authoring-time seed rule file.

The compiled AIR contains the normalized deterministic portfolio model the backtester needs, including:

- starting capital
- initial allocation
- candidate baskets
- selection policy
- targets
- constraints
- risk budgets
- rebalance settings
- executable rules

---

## Initial CLI

Recommended command:

```bash
alphanet-backtest ./strategies/oil-rates-growth-tech/compiled/strategy.ir.json \
  --start 2018-01-01 \
  --end 2025-12-31
```

Optional flags:

```bash
--data ./data
--data-version local-v0.1
--out ./strategies/oil-rates-growth-tech/backtests
--starting-cash 100000
--backtester-version v0.1.0
--validate-only
--decision-trace
--fail-on-missing-data
--window-trading-days-target 60
--max-memory-mb 256
--max-storage-mb 2048
--prune-after-window=false
```

`--starting-cash` should be treated as an explicit runtime override.

If not provided, the backtester should use `portfolio.starting_cash` from AIR.

---

## Inputs

### Required

```text
compiled/strategy.ir.json
```

### Required CLI Config

- start date
- end date
- output directory

### Required Runtime Config

- data source
- data version
- execution assumptions

### Optional

- starting cash override
- benchmark
- transaction cost override
- slippage override
- window trading days target
- max memory
- max local storage
- prune-after-window policy
- debug mode
- trace verbosity

---

## Outputs

### Summary

```text
YYYY-MM-DD_YYYY-MM-DD.summary.json
```

Contains:

- total return
- annualized return
- volatility
- Sharpe ratio
- Sortino ratio
- max drawdown
- turnover
- win rate
- final value
- exposure summaries
- basket allocation summaries
- candidate selection summaries
- data plan and data resolution statistics

### Equity Curve

```text
YYYY-MM-DD_YYYY-MM-DD.equity-curve.csv
```

Contains:

- date
- portfolio value
- daily return
- cash
- gross exposure
- net exposure
- drawdown

### Trades

```text
YYYY-MM-DD_YYYY-MM-DD.trades.csv
```

Contains:

- date
- symbol
- side
- quantity
- price
- notional
- cost
- slippage
- reason code
- decision trace id
- source rule id
- source basket id, if applicable

### Decision Trace

```text
YYYY-MM-DD_YYYY-MM-DD.decision-trace.jsonl
```

Contains one JSON object per decision day.

Records:

- evaluated signals
- active regimes
- active relations
- triggered rules
- requested actions
- rejected actions
- approved actions
- candidate basket rankings
- portfolio constraints
- portfolio state before/after

---

## Forbidden Behavior

The backtester must not:

- call OpenAI
- call Claude
- call any LLM
- call TradingAgents
- call AI Hedge Fund
- read `strategy.md`
- read `manifest.json` for portfolio state
- read `rules.json`
- reinterpret natural language intent
- mutate `strategy.ir.json`
- silently ignore invalid references
- fetch non-versioned data for verified runs

---

## Proposed Go Package Structure

```text
backtester/
├── go.mod
├── cmd/
│   └── alphanet-backtest/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── backtest.go
│   ├── air/
│   │   ├── load.go
│   │   ├── model.go
│   │   └── validate.go
│   ├── data/
│   │   ├── planner.go
│   │   ├── requirements.go
│   │   ├── resolver.go
│   │   ├── dolt.go
│   │   ├── provider.go
│   │   ├── csv.go
│   │   ├── memory.go
│   │   └── calendar.go
│   ├── signals/
│   │   ├── engine.go
│   │   ├── transforms.go
│   │   └── registry.go
│   ├── regimes/
│   │   └── engine.go
│   ├── relations/
│   │   └── engine.go
│   ├── rules/
│   │   ├── engine.go
│   │   ├── condition.go
│   │   ├── actions.go
│   │   └── conflicts.go
│   ├── portfolio/
│   │   ├── state.go
│   │   ├── engine.go
│   │   ├── initialization.go
│   │   ├── baskets.go
│   │   ├── selection_policy.go
│   │   ├── constraints.go
│   │   ├── rebalance.go
│   │   └── risk.go
│   ├── execution/
│   │   ├── simulator.go
│   │   ├── orders.go
│   │   └── costs.go
│   ├── metrics/
│   │   ├── returns.go
│   │   ├── drawdown.go
│   │   └── summary.go
│   ├── trace/
│   │   └── trace.go
│   └── output/
│       ├── summary.go
│       ├── equity_curve.go
│       ├── trades.go
│       └── decision_trace.go
└── testdata/
    ├── air/
    ├── prices/
    └── expected/
```

---

## Suggested Core Types

### BacktestConfig

```go
type BacktestConfig struct {
    StartDate string
    EndDate string
    DataPath string
    DataVersion string
    OutputPath string
    StartingCash *float64
    WindowTradingDaysTarget int
    MaxMemoryMB int
    MaxStorageMB int
    PruneAfterWindow bool
    Trace bool
}
```

### DataRequirementPlan

The backtester should create a `DataRequirementPlan` before loading data.

The plan is derived from compiled AIR plus runtime date/config options.

It should include:

- required provider sources
- symbols and series
- fields
- transform lookbacks
- processing windows
- lookback-expanded data windows
- estimated rows and storage
- estimated memory
- source storage budgets
- expected access modes

Example:

```go
type DataRequirementPlan struct {
    StartDate time.Time
    EndDate time.Time
    WindowTradingDaysTarget int
    Windows []DataWindow
    Sources []SourceRequirement
    MaxMemoryMB int
    MaxStorageMB int
    SourceBudgets map[string]int
}
```

### DataWindow

```go
type DataWindow struct {
    ProcessingStart time.Time
    ProcessingEnd time.Time
    RequiredStart time.Time
    RequiredEnd time.Time
    MaxLookbackTradingDays int
    EstimatedRows int64
    EstimatedMemoryMB int
    EstimatedStorageMB int
}
```

`ProcessingStart` and `ProcessingEnd` define the dates evaluated by the
backtester.

`RequiredStart` and `RequiredEnd` define the wider data range needed to compute
lookback-dependent signals for that processing window.

### SourceRequirement

```go
type SourceRequirement struct {
    Source string
    Database string
    Dataset string
    Symbols []string
    Series []string
    Fields []string
    RequiredStart time.Time
    RequiredEnd time.Time
    LookbackTradingDays int
    Transforms []string
    EstimatedRows int64
    EstimatedMemoryMB int
    EstimatedStorageMB int
}
```

Source requirements should be granular enough for the resolver to detect and
fill gaps without fetching entire provider databases.

### DataResolutionStats

```go
type DataResolutionStats struct {
    Source string
    Database string
    Commit string
    AccessModesUsed []string
    RowsFromLocal int64
    RowsFromRemote int64
    RowsImported int64
    MissingRows int64
}
```

These stats should be included in summary output.

### PortfolioState

```go
type PortfolioState struct {
    Date string
    Cash float64
    Positions map[string]Position
    TotalValue float64
    Weights map[string]float64
    BasketWeights map[string]float64
}
```

### Position

```go
type Position struct {
    Symbol string
    Quantity float64
    MarketValue float64
    Weight float64
    CostBasis float64
}
```

### InitialAllocation

```go
type InitialAllocation struct {
    Mode string
    Positions []InitialPosition
}
```

Supported modes:

- `cash`
- `weights`
- `dollars`
- `shares`

### CandidateBasket

```go
type CandidateBasket struct {
    BasketID string
    Symbols []string
    AssetClass string
    Sector string
    Theme string
    MinWeight *float64
    MaxWeight *float64
    MinPositionWeight *float64
    MaxPositionWeight *float64
}
```

### BasketScore

```go
type BasketScore struct {
    BasketID string
    Symbol string
    Score float64
    Rank int
    Components map[string]float64
}
```

### SignalValue

```go
type SignalValue struct {
    SignalID string
    Date string
    Value any
    Available bool
}
```

### RequestedAction

```go
type RequestedAction struct {
    RuleID string
    Layer string
    Priority int
    Confidence float64
    Action string
    Target string
    TargetType string
    Amount float64
    Unit string
    Reason string
}
```

### ApprovedOrder

```go
type ApprovedOrder struct {
    Symbol string
    Side string
    Quantity float64
    TargetWeightDelta float64
    SourceRuleID string
    SourceBasketID string
    Reason string
}
```

---

## Backtest Pipeline

Recommended pipeline:

```text
Load AIR
    ↓
Validate AIR
    ↓
Build DataRequirementPlan
    ↓
For each rolling window:
    Expand data requirements by signal lookback
    Resolve provider Dolt clone gaps from local, remote SQL, or importer
    Load normalized runtime records for the processing window
    ↓
Initialize portfolio from AIR initial_allocation, on first window only
    ↓
For each date in processing window:
    Evaluate signals
    Evaluate regimes
    Evaluate relations
    Rank candidate baskets, if needed
    Evaluate rules
    Resolve conflicts
    Apply portfolio constraints
    Apply basket selection policy
    Simulate execution
    Update portfolio
    Record trace
    Record metrics
    Optionally prune provider Dolt clone data no longer needed
    ↓
Write outputs
```

---

## Portfolio Initialization

At the start of a backtest, the portfolio engine should initialize from:

```text
portfolio.starting_cash
portfolio.initial_allocation
```

Supported modes:

### `cash`

Start 100% in cash.

### `weights`

Convert target weights into dollar allocations using starting account value and first available execution prices.

### `dollars`

Convert dollar allocations into initial positions.

### `shares`

Use supplied share counts and calculate remaining cash.

Validation rules:

- weights should sum to 1.0 unless residual cash is explicitly allowed
- cash may be represented as symbol `cash`
- symbols must exist in the AIR universe or candidate universe
- initial positions must not violate hard constraints unless explicitly allowed for historical reproduction
- missing initial allocation should default to 100% cash

---

## Data Provider Interface

The backtester needs a data abstraction.

Example:

```go
type DataProvider interface {
    TradingDays(start, end time.Time) ([]time.Time, error)
    Price(symbol string, date time.Time, field string) (float64, error)
    Series(instrument string, field string, start, end time.Time) ([]DataPoint, error)
    Event(eventID string, date time.Time) (EventValue, error)
}
```

The primary v1 data path should use provider-specific local Dolt clones as the
cache and workspace.

The resolver should satisfy `DataRequirementPlan` gaps using:

```text
local provider Dolt clone
    ↓
public remote Dolt SQL query for missing slices
    ↓
provider importer into local provider Dolt clone
```

CSV/local files may remain useful for unit tests, golden tests, and explicit
user overrides, but they should not be the main cache architecture.

Provider adapters should support:

- Alpaca local Dolt clone
- FRED
- Cboe
- IMF PortWatch
- Alpha Vantage
- Yahoo Finance, optional
- CSV/local override

---

## Data Planning and Resolution

The planner must inspect compiled AIR before simulation starts.

It should collect requirements from:

- signal interests
- executable rule conditions
- candidate basket ranking signals
- initial allocation symbols
- candidate basket symbols
- benchmark symbols
- execution pricing assumptions

For each requirement, the planner should determine provider, dataset, symbol or
series, field, transform, frequency, and lookback.

The planner should then build adaptive rolling windows:

```text
window_trading_days_target = desired processing size
required_start = processing_start - requirement_lookback
required_end = processing_end
```

`window_trading_days_target` is a target. The planner may shrink the actual
window when estimated memory or storage would exceed configured limits.

Remote Dolt SQL result sets count against `max_memory_mb` while the current
rolling window is being resolved and evaluated. After the processing window is
complete, remote result memory should be released unless those rows were
explicitly imported or streamed into a local provider Dolt clone.

Storage budgets should be allocated across provider Dolt clones with the
default `weighted_by_requirement` formula:

```text
provider_weight = required_rows * average_row_width * reuse_factor
provider_budget = max_storage_mb * provider_weight / sum(provider_weights)
```

A market data source with many symbols, fields, rows, or repeated reuse should
usually receive a larger allocation than compact macro series.

For each window, the resolver should:

1. Query the local provider Dolt clone.
2. Detect missing symbols, series, fields, or date ranges.
3. Query public remote Dolt SQL for the missing slices.
4. Run the relevant importer into the local provider Dolt clone if data remains missing.
5. Release remote query result memory after the processing window completes.
6. Record resolution stats and data commits.

Pruning completed window data should be configurable. When enabled, pruning
must actually reclaim enough disk space to keep provider clones within their
storage budgets. The implementation may use row deletion, ephemeral branches,
temporary clones, Dolt garbage collection, compaction, or clone rotation, but it
must verify measured on-disk usage after pruning. If the budget cannot be met,
the planner must shrink future windows, avoid more imports, or fail with a
clear storage-budget error.

AlphaNet should periodically rebalance the public provider Dolt databases based
on public usage and official strategy needs. Commonly used data should be
preloaded so user backtests stay affordable, while less common data can remain
available through remote query or importer fallback.

---

## Signal Engine

The signal engine evaluates signal definitions from AIR.

Example signal:

```json
{
  "signal_id": "wti_change_20d",
  "instrument": "WTI",
  "field": "close",
  "transform": "percent_change",
  "window": "20d"
}
```

The engine should:

1. Resolve data series.
2. Apply transform.
3. Return value for current date.
4. Mark missing values clearly.

Initial transforms:

- level
- change
- percent_change
- moving_average
- z_score
- realized_volatility
- rank
- percentile_rank

---

## Regime Engine

The regime engine evaluates regime conditions.

A regime may be active or inactive.

Example:

```json
{
  "regime_id": "tight_liquidity",
  "conditions": {
    "signal": "ust10y_change_20d",
    "operator": ">",
    "value": 25
  }
}
```

The engine should output:

```json
{
  "regime_id": "tight_liquidity",
  "active": true,
  "confidence": 0.8
}
```

---

## Relation Engine

Relations evaluate cross-asset or cross-domain beliefs.

Example:

```json
{
  "relation_id": "oil_rates_negative_for_growth_tech",
  "drivers": ["wti_change_20d", "ust10y_change_20d"],
  "targets": ["QQQ", "NVDA"],
  "effect": "negative"
}
```

For v0.1, relation evaluation can be simple:

- if relation conditions are true, relation is active
- otherwise inactive

---

## Candidate Basket Engine

Candidate baskets define assets the strategy may buy later, even if they are not in the initial allocation.

For each basket, the backtester should be able to:

1. Resolve basket symbols.
2. Validate required data exists.
3. Evaluate ranking signals.
4. Score each symbol.
5. Select eligible symbols.
6. Enforce basket min/max weights.
7. Emit candidate ranking details into the decision trace.

Example:

```json
{
  "basket_id": "growth_technology",
  "symbols": ["QQQ", "NVDA", "AMD", "MSFT", "AAPL", "AVGO", "SMH"]
}
```

---

## Selection Policy Engine

The selection policy decides how to choose within candidate baskets.

Example behavior:

```text
For growth_technology:
    rank all symbols by weighted score
    select top 5
    enforce min and max position weights
    avoid replacing holdings unless rank change exceeds threshold
```

The engine should support at least:

- `static`
- `ranked`
- `equal_weight`
- `manual`

For v0.1, a simple ranked implementation is enough.

---

## Rule Engine

The rule engine evaluates rules using:

- signal values
- regime states
- relation states
- portfolio state
- basket rankings
- date state
- event state

It emits requested actions.

The rule engine should not enforce portfolio constraints.

Rules may target:

- symbols
- baskets
- asset classes
- sectors
- themes
- cash
- portfolio

Example basket-targeting action:

```json
{
  "action": "decrease_weight",
  "target": "growth_technology",
  "target_type": "basket",
  "amount": 0.10,
  "unit": "weight"
}
```

---

## Conflict Resolution

Conflict resolution order:

```text
Layer Priority
    ↓
Rule Priority
    ↓
Confidence
    ↓
Tie Breaker
```

For v0.1:

1. Group actions by target and target type.
2. Detect conflicting actions.
3. Resolve by decision hierarchy.
4. Pass non-conflicting requested actions to portfolio engine.
5. Record conflicts in decision trace.

---

## Portfolio Engine

The portfolio engine receives requested actions and decides what can be approved.

It should enforce:

- cash minimums
- max single position
- max sector weight
- max asset class weight
- max basket weight
- min/max position weights
- max leverage
- no shorting
- no margin
- risk budgets
- turnover limits

It may:

- approve actions
- scale actions
- reject actions
- substitute safer actions
- force rebalancing

When actions target baskets, the portfolio engine should expand them into symbol-level order intents using the relevant candidate basket and selection policy.

---

## Execution Simulator

The execution simulator turns approved orders into filled trades.

Initial assumptions:

- order timing: `next_open`
- fractional shares allowed
- transaction cost in basis points
- slippage in basis points
- adjusted prices used for corporate actions

For v0.1, keep this simple.

Example fill:

```text
fill_price = next_open_price * (1 + slippage_bps / 10000)
```

For sells:

```text
fill_price = next_open_price * (1 - slippage_bps / 10000)
```

---

## Metrics

Initial metrics:

- total return
- annualized return
- daily returns
- volatility
- Sharpe ratio
- Sortino ratio
- max drawdown
- turnover
- final value
- cash utilization
- average exposure
- basket exposure
- sector exposure
- candidate selection turnover

---

## Decision Trace

Decision trace is central to AlphaNet.

Each decision record should include:

```json
{
  "date": "2026-06-19",
  "portfolio_before": {},
  "signals": {},
  "regimes": {},
  "relations": {},
  "basket_rankings": {},
  "rules_triggered": [],
  "actions_requested": [],
  "conflicts": [],
  "actions_approved": [],
  "actions_rejected": [],
  "portfolio_after": {}
}
```

Trace verbosity can be configurable.

---

## First Implementation Milestone

Build a one-day evaluator.

Milestone 1:

1. Load AIR.
2. Build a minimal `DataRequirementPlan`.
3. Resolve one day of local test data.
4. Initialize portfolio from `initial_allocation`.
5. Evaluate one or two signals.
6. Evaluate one regime.
7. Evaluate one rule.
8. Emit requested actions.
9. Pass through portfolio engine.
10. Emit decision trace.

No real historical backtest yet.

---

## Second Implementation Milestone

Build multi-day replay.

Tasks:

1. Load date range.
2. Build adaptive rolling windows from `window_trading_days_target`.
3. Expand each window by required signal lookback.
4. Resolve local provider Dolt clone gaps.
5. Iterate trading days.
6. Evaluate signals daily.
7. Track portfolio state.
8. Track basket exposures.
9. Generate equity curve.
10. Generate decision trace.

---

## Third Implementation Milestone

Add execution simulation.

Tasks:

1. Convert approved actions into orders.
2. Fill orders using next open or close.
3. Apply costs and slippage.
4. Track positions.
5. Track cash.
6. Emit trades CSV.

---

## Fourth Implementation Milestone

Add candidate basket selection.

Tasks:

1. Load candidate baskets from AIR.
2. Evaluate ranking signals.
3. Score and rank basket symbols.
4. Select holdings based on policy.
5. Expand basket actions into symbol-level orders.
6. Record ranking details in decision trace.

---

## Fifth Implementation Milestone

Add metrics.

Tasks:

1. Daily returns.
2. Total return.
3. Annualized return.
4. Volatility.
5. Sharpe.
6. Sortino.
7. Max drawdown.
8. Turnover.
9. Basket and sector exposure summaries.

---

## Sixth Implementation Milestone

Add schema and semantic validation.

Tasks:

1. Validate AIR schema.
2. Validate references.
3. Validate universe symbols.
4. Validate initial allocation.
5. Validate candidate baskets.
6. Validate selection policies.
7. Validate portfolio constraints.
8. Validate rule targets.
9. Validate signal definitions.

---

## Testing Plan

### Unit Tests

- signal transforms
- condition evaluation
- regime activation
- relation activation
- basket ranking
- selection policy
- rule matching
- conflict resolution
- portfolio initialization
- portfolio constraints
- execution fills
- metrics

### Integration Tests

- example AIR runs without error
- initial allocation creates expected starting positions
- expected rule fires on known date
- basket-targeting rule expands into symbol-level orders
- cash minimum blocks invalid buy
- max position scales order
- output files are generated

### Golden Tests

Use fixed input data and expected outputs.

Golden tests should verify:

- same AIR
- same data
- same date range
- same output hash

---

## Data Strategy for v0.1

Start with provider-specific local Dolt clones.

Example:

```text
data/
└── dolthub/
    ├── alphanet_alpaca/
    ├── alphanet_fred/
    ├── alphanet_cboe/
    ├── alphanet_imf_portwatch/
    ├── alphanet_alpha_vantage/
    └── alphanet_yahoo_finance/
```

The backtester should use the `DataRequirementPlan` to query only the source,
symbol, series, field, and date ranges required for the current rolling window.

For local development and tests, CSV fixtures may still be used.

Market price fixture fields:

```text
date,open,high,low,close,adjusted_close,volume
```

Macro fixture fields:

```text
date,value
```

Candidate basket symbols should use the same market price schema as tradable
symbols.

For production-like runs, missing local Dolt data should be resolved through
remote Dolt SQL first, then provider importer fallback.

Imported rolling-window data should remain uncommitted by default unless local
or public writes are explicitly enabled.

---

## Non-Goals for v0.1

Do not implement yet:

- live trading
- broker integrations
- intraday simulation
- options modeling
- margin accounts
- short borrow costs
- tax-aware accounting
- leaderboard server
- distributed backtesting
- automatic data licensing

---

## Success Criteria

The backtester is successful when:

1. It loads valid AIR.
2. It initializes the portfolio from AIR.
3. It runs over a date range.
4. It evaluates signals deterministically.
5. It evaluates rules deterministically.
6. It evaluates candidate basket selections deterministically.
7. It enforces portfolio constraints.
8. It emits trades, equity curve, summary, and decision trace.
9. The same inputs produce the same outputs.

# Backtester Plan Update: Benchmarks and AlphaNet Scoring

This update is intended to be added to:

```text
backtester/backtester-plan.md
```

It gives the backtester implementer detailed guidance for calculating benchmark results and AlphaNet Score.

---

## Benchmark and Scoring Requirements

The backtester must produce two related but separate outputs:

1. **Benchmark returns**
2. **AlphaNet Score**

Benchmarks are for comparison.

AlphaNet Score is a composite score for ranking and evaluating strategies.

Benchmarks should not directly change the v0.1 score yet. They should be included in the summary output so users can interpret strategy performance against common alternatives.

---

## Required Summary Output Fields

The backtest summary should continue to include the core performance fields:

```json
{
  "strategy_name": "Oil Rates Growth Tech",
  "strategy_ir_sha256": "sha256:example",
  "spec_version": "v0.1",
  "backtester_version": "v0.1.0",
  "data_version": "example-data-v0.1",
  "date_range": {
    "start": "2018-01-01",
    "end": "2025-12-31"
  },
  "starting_cash": 100000,
  "ending_value": 137250.42,
  "total_return": 0.3725042,
  "annualized_return": 0.0412,
  "volatility": 0.142,
  "sharpe": 0.61,
  "sortino": 0.84,
  "max_drawdown": -0.168,
  "turnover": 1.94,
  "data_plan": {}
}
```

The backtester should add:

```json
{
  "benchmarks": [],
  "score": {}
}
```

---

## Benchmark Configuration

Benchmarks should be configured from the strategy manifest or compiled AIR.

Recommended default benchmark set:

```json
[
  {
    "symbol": "SPY",
    "name": "SPDR S&P 500 ETF Trust",
    "type": "fund"
  },
  {
    "symbol": "QQQ",
    "name": "Invesco QQQ Trust",
    "type": "fund"
  },
  {
    "symbol": "IWM",
    "name": "iShares Russell 2000 ETF",
    "type": "fund"
  },
  {
    "symbol": "TLT",
    "name": "iShares 20+ Year Treasury Bond ETF",
    "type": "fund"
  },
  {
    "symbol": "AGG",
    "name": "iShares Core U.S. Aggregate Bond ETF",
    "type": "fund"
  },
  {
    "symbol": "XAUUSD",
    "name": "Gold Spot Price USD",
    "type": "commodity_spot",
    "currency": "USD",
    "data_field": "close"
  },
  {
    "symbol": "DBC",
    "name": "Invesco DB Commodity Index Tracking Fund",
    "type": "fund"
  },
  {
    "symbol": "CASH",
    "name": "Cash / risk-free baseline",
    "type": "cash"
  }
]
```

### Benchmark Type Semantics

| Type | Meaning | Return calculation |
|---|---|---|
| `fund` | ETF or mutual fund style instrument | adjusted close return |
| `index` | index level | close or level return |
| `commodity_spot` | spot commodity series such as `XAUUSD` | spot close return |
| `cash` | simple cash baseline | v0.1 defaults to 0% |
| `custom` | user-defined provider logic | adapter-defined |

---

## Benchmark Result Shape

Each benchmark result should use this shape:

```json
{
  "benchmark": "XAUUSD",
  "name": "Gold Spot Price USD",
  "type": "commodity_spot",
  "total_return": 0.62,
  "annualized_return": 0.064,
  "start_value": 1825.50,
  "end_value": 2957.31,
  "currency": "USD",
  "data_field": "close"
}
```

Required fields:

```text
benchmark
name
type
total_return
annualized_return
```

Optional fields:

```text
start_value
end_value
currency
data_field
notes
```

---

## Benchmark Calculation Rules

Benchmarks must be calculated over the exact same date range as the strategy backtest.

Use:

```text
strategy_start = summary.date_range.start
strategy_end = summary.date_range.end
```

### Fund Benchmarks

For `fund` benchmarks:

```text
benchmark_total_return =
  ending_adjusted_close / starting_adjusted_close - 1
```

Use `adjusted_close` by default.

If adjusted close is unavailable, the backtester may use `close`, but the benchmark result should include a note:

```json
{
  "notes": "adjusted_close unavailable; close used"
}
```

### Index Benchmarks

For `index` benchmarks:

```text
benchmark_total_return =
  ending_level / starting_level - 1
```

### Commodity Spot Benchmarks

For `commodity_spot`, such as `XAUUSD`:

```text
benchmark_total_return =
  ending_spot_price / starting_spot_price - 1
```

Use `close` by default unless the benchmark config specifies another field.

### Cash Benchmark

For v0.1, cash may be treated as zero return:

```text
total_return = 0
annualized_return = 0
```

Later, this can be replaced by a Treasury bill, money market, or risk-free rate series.

---

## Annualized Return Calculation

For the strategy and all benchmarks, annualized return should use the same function.

```text
annualized_return =
  (1 + total_return) ^ (1 / years) - 1
```

Where:

```text
years = number_of_days_between_start_and_end / 365.25
```

If `years <= 0`, return an error.

If `1 + total_return <= 0`, annualized return should be set to `-1.0` or the run should fail validation, depending on implementation preference. The safer v0.1 behavior is to set it to `-1.0` and add a warning.

---

## AlphaNet Score Overview

AlphaNet Score is a 0-100 composite score.

It is designed to reward strategies that:

- create value
- produce positive expected value
- have good risk-adjusted returns
- avoid large drawdowns
- work across longer periods
- work across multiple windows
- avoid unrealistic turnover

The v0.1 formula is:

```text
AlphaNet Score =
100 * (
  0.25 * return_score
+ 0.25 * risk_adjusted_score
+ 0.20 * drawdown_score
+ 0.15 * consistency_score
+ 0.10 * duration_score
+ 0.05 * turnover_score
)
```

---

## Raw Score vs Official Score

The backtester should support two score concepts.

### Raw Score

`raw_score` is calculated for any backtest range.

This is useful for local experimentation.

```json
{
  "raw_score": 52.7
}
```

### Official Score

`official_score` is only calculated for official AlphaNet evaluation windows.

Official score should require:

- AlphaNet-defined date windows
- sufficient historical duration
- no user-defined excluded ranges
- daily valuation when required
- declared recompilation schedule if the strategy is recompiled during the backtest
- no future data leakage

If a run is not eligible for official scoring, set:

```json
{
  "official_score": null,
  "official_status": "not_official_window"
}
```

---

## Score Output Shape

The summary should include:

```json
{
  "score": {
    "raw_score": 52.7,
    "official_score": null,
    "official_status": "not_official_window",
    "reason": "This backtest was run on a user-selected date range.",
    "description": "AlphaNet Score is a 0-100 composite score based on return, risk-adjusted return, drawdown control, consistency, duration, and turnover realism.",
    "components": {
      "return_score": 0.31,
      "risk_adjusted_score": 0.28,
      "drawdown_score": 0.66,
      "consistency_score": 0.72,
      "duration_score": 0.80,
      "turnover_score": 0.81
    },
    "weights": {
      "return_score": 0.25,
      "risk_adjusted_score": 0.25,
      "drawdown_score": 0.20,
      "consistency_score": 0.15,
      "duration_score": 0.10,
      "turnover_score": 0.05
    }
  }
}
```

---

## Component Normalization

All component scores should be normalized to the range:

```text
0.0 to 1.0
```

Use a helper:

```go
func Clamp01(v float64) float64 {
    if v < 0 {
        return 0
    }
    if v > 1 {
        return 1
    }
    return v
}
```

---

## Return Score

Return score blends annualized return and total return.

Annualized return makes different date ranges comparable.

Total return rewards longer compounding periods.

Recommended v0.1 formula:

```text
annualized_return_score = clamp(annualized_return / 0.20, 0, 1)

total_return_score =
  clamp(log(ending_value / starting_cash) / log(4), 0, 1)

return_score =
  0.70 * annualized_return_score
+ 0.30 * total_return_score
```

Interpretation:

- 20% annualized return is excellent.
- 4x total account growth is excellent.
- negative or zero account growth scores poorly.

Implementation notes:

- if `starting_cash <= 0`, fail validation.
- if `ending_value <= 0`, set `total_return_score = 0`.
- use natural log.

---

## Risk-Adjusted Score

Risk-adjusted score blends Sharpe and Sortino.

Recommended v0.1 formula:

```text
sharpe_score = clamp(sharpe / 2.0, 0, 1)

sortino_score = clamp(sortino / 3.0, 0, 1)

risk_adjusted_score =
  0.45 * sharpe_score
+ 0.55 * sortino_score
```

Interpretation:

- Sharpe of 2.0 is excellent.
- Sortino of 3.0 is excellent.
- Sortino is weighted slightly more because downside volatility matters more than upside volatility.

If Sharpe or Sortino is NaN or infinite, treat that component as zero and emit a warning.

---

## Drawdown Score

Drawdown score penalizes large losses.

Use absolute max drawdown:

```text
drawdown = abs(max_drawdown)
```

Recommended formula:

```text
drawdown_score = clamp(1 - drawdown / 0.50, 0, 1)
```

Interpretation:

| Max Drawdown | Drawdown Score |
|---:|---:|
| 0% | 1.00 |
| 10% | 0.80 |
| 20% | 0.60 |
| 30% | 0.40 |
| 50%+ | 0.00 |

---

## Consistency Score

Consistency score should reward strategies that work across multiple periods.

For a single local backtest, if rolling-window details are not available yet, v0.1 may use:

```text
consistency_score = 0.5
```

and emit:

```json
{
  "reason": "Rolling-window consistency not available; neutral consistency score used."
}
```

However, the intended implementation should calculate rolling-window consistency.

Recommended approach:

1. Split the full backtest into rolling windows.
2. Suggested default: 3-year windows stepped annually.
3. Calculate each window's total return and max drawdown.
4. Compute positive return rate.
5. Compute acceptable drawdown rate.
6. Blend both.

Formula:

```text
positive_window_rate =
  number_of_windows_with_total_return_above_0 / total_windows

acceptable_drawdown_window_rate =
  number_of_windows_with_max_drawdown_better_than_-25_percent / total_windows

consistency_score =
  0.60 * positive_window_rate
+ 0.40 * acceptable_drawdown_window_rate
```

If fewer than two rolling windows are available:

```text
consistency_score = 0.5
```

and mark the score as less reliable.

---

## Duration Score

Duration score rewards longer evaluated periods.

Recommended formula:

```text
duration_score = clamp(years_tested / 10, 0, 1)
```

Interpretation:

| Years Tested | Duration Score |
|---:|---:|
| 1 | 0.10 |
| 3 | 0.30 |
| 5 | 0.50 |
| 10+ | 1.00 |

For official scoring, the recommended minimum history is:

```text
3 years
```

If the run has less than 3 years:

```json
{
  "official_score": null,
  "official_status": "insufficient_history"
}
```

The raw score may still be calculated.

---

## Turnover Score

Turnover score penalizes unrealistic trading intensity.

The scoring formula should use annualized turnover.

If the summary field `turnover` is already annualized, use it directly.

If `turnover` is total turnover over the whole test, convert it:

```text
annual_turnover = total_turnover / years_tested
```

Recommended formula:

```text
turnover_score = clamp(1 - annual_turnover / 10, 0, 1)
```

Interpretation:

| Annual Turnover | Turnover Score |
|---:|---:|
| 0x | 1.00 |
| 1x | 0.90 |
| 3x | 0.70 |
| 5x | 0.50 |
| 10x+ | 0.00 |

Do not over-penalize active strategies, but very high turnover should reduce the score.

---

## Score Calculation Pseudocode

```go
func CalculateAlphaNetScore(summary Summary, windows []WindowResult) Score {
    years := YearsBetween(summary.StartDate, summary.EndDate)

    annualizedReturnScore := Clamp01(summary.AnnualizedReturn / 0.20)

    totalGrowth := summary.EndingValue / summary.StartingCash
    totalReturnScore := 0.0
    if totalGrowth > 0 {
        totalReturnScore = Clamp01(math.Log(totalGrowth) / math.Log(4))
    }

    returnScore := 0.70*annualizedReturnScore + 0.30*totalReturnScore

    sharpeScore := Clamp01(summary.Sharpe / 2.0)
    sortinoScore := Clamp01(summary.Sortino / 3.0)
    riskAdjustedScore := 0.45*sharpeScore + 0.55*sortinoScore

    drawdownScore := Clamp01(1 - math.Abs(summary.MaxDrawdown)/0.50)

    consistencyScore := CalculateConsistencyScore(windows)

    durationScore := Clamp01(years / 10.0)

    annualTurnover := summary.Turnover
    if !summary.TurnoverIsAnnualized {
        annualTurnover = summary.Turnover / years
    }
    turnoverScore := Clamp01(1 - annualTurnover/10.0)

    raw := 100 * (
        0.25*returnScore +
        0.25*riskAdjustedScore +
        0.20*drawdownScore +
        0.15*consistencyScore +
        0.10*durationScore +
        0.05*turnoverScore
    )

    return Score{
        RawScore: raw,
        Components: ScoreComponents{
            ReturnScore: returnScore,
            RiskAdjustedScore: riskAdjustedScore,
            DrawdownScore: drawdownScore,
            ConsistencyScore: consistencyScore,
            DurationScore: durationScore,
            TurnoverScore: turnoverScore,
        },
    }
}
```

---

## Official Scoring Eligibility

The backtester should determine whether the run is eligible for official scoring.

Recommended statuses:

```text
official
not_official_window
insufficient_history
uses_excluded_ranges
missing_required_valuation
invalid
not_applicable
```

### Official Score Requirements

A run can receive `official_score` only if:

1. The date range matches an AlphaNet official evaluation window or suite.
2. The strategy was evaluated for the minimum required duration.
3. The run did not use user-defined excluded ranges.
4. Daily valuation was available when required.
5. The data version is accepted for official scoring.
6. The backtester version is accepted for official scoring.
7. If the strategy was recompiled during the test, the recomputation schedule was declared ahead of time.
8. Each compiled AIR version only used training data available up to its effective date.

If any requirement fails, set `official_score = null` and set the proper `official_status`.

---

## Recompiled Strategy Handling

If a strategy is recompiled on a schedule, the backtester may receive multiple AIR versions.

Example:

```json
{
  "strategy_versions": [
    {
      "effective_from": "2020-01-01",
      "strategy_ir_sha256": "sha256:a"
    },
    {
      "effective_from": "2020-02-01",
      "strategy_ir_sha256": "sha256:b"
    },
    {
      "effective_from": "2020-03-01",
      "strategy_ir_sha256": "sha256:c"
    }
  ]
}
```

For scoring, this should be treated as one stitched backtest.

The score is calculated on the full stitched equity curve.

Important rule:

```text
No AIR version may be compiled using data after its effective_from date.
```

This prevents lookahead bias.

---

## Benchmark Calculation Pseudocode

```go
func CalculateBenchmarkResult(
    benchmark BenchmarkConfig,
    provider DataProvider,
    start time.Time,
    end time.Time,
) (BenchmarkResult, error) {
    if benchmark.Type == "cash" {
        return BenchmarkResult{
            Benchmark: benchmark.Symbol,
            Name: benchmark.Name,
            Type: benchmark.Type,
            TotalReturn: 0,
            AnnualizedReturn: 0,
        }, nil
    }

    field := benchmark.DataField
    if field == "" {
        if benchmark.Type == "commodity_spot" {
            field = "close"
        } else {
            field = "adjusted_close"
        }
    }

    startValue, err := provider.ValueOnOrNear(benchmark.Symbol, field, start, "nearest_next")
    if err != nil {
        return BenchmarkResult{}, err
    }

    endValue, err := provider.ValueOnOrNear(benchmark.Symbol, field, end, "nearest_previous")
    if err != nil {
        return BenchmarkResult{}, err
    }

    if startValue <= 0 {
        return BenchmarkResult{}, fmt.Errorf("invalid benchmark start value")
    }

    totalReturn := endValue/startValue - 1
    years := YearsBetween(start, end)
    annualizedReturn := AnnualizeReturn(totalReturn, years)

    return BenchmarkResult{
        Benchmark: benchmark.Symbol,
        Name: benchmark.Name,
        Type: benchmark.Type,
        TotalReturn: totalReturn,
        AnnualizedReturn: annualizedReturn,
        StartValue: startValue,
        EndValue: endValue,
        Currency: benchmark.Currency,
        DataField: field,
    }, nil
}
```

---

## Data Requirements

The backtester data layer must support benchmark series.

Minimum v0.1 benchmark data requirements:

| Benchmark | Required field |
|---|---|
| SPY | adjusted_close |
| QQQ | adjusted_close |
| IWM | adjusted_close |
| TLT | adjusted_close |
| AGG | adjusted_close |
| XAUUSD | close |
| DBC | adjusted_close |
| CASH | none |

If benchmark data is missing:

- local raw score may still be calculated
- benchmark result should include an error or warning
- official scoring should fail if benchmarks are required for the official run

---

## Summary Output Example

```json
{
  "strategy_name": "Oil Rates Growth Tech",
  "starting_cash": 100000,
  "ending_value": 137250.42,
  "total_return": 0.3725042,
  "annualized_return": 0.0412,
  "volatility": 0.142,
  "sharpe": 0.61,
  "sortino": 0.84,
  "max_drawdown": -0.168,
  "turnover": 1.94,
  "benchmarks": [
    {
      "benchmark": "SPY",
      "name": "SPDR S&P 500 ETF Trust",
      "type": "fund",
      "total_return": 0.95,
      "annualized_return": 0.085
    },
    {
      "benchmark": "XAUUSD",
      "name": "Gold Spot Price USD",
      "type": "commodity_spot",
      "total_return": 0.62,
      "annualized_return": 0.064
    },
    {
      "benchmark": "CASH",
      "name": "Cash / risk-free baseline",
      "type": "cash",
      "total_return": 0,
      "annualized_return": 0
    }
  ],
  "score": {
    "raw_score": 52.7,
    "official_score": null,
    "official_status": "not_official_window",
    "reason": "This backtest was run on a user-selected date range.",
    "components": {
      "return_score": 0.31,
      "risk_adjusted_score": 0.28,
      "drawdown_score": 0.66,
      "consistency_score": 0.72,
      "duration_score": 0.80,
      "turnover_score": 0.81
    },
    "weights": {
      "return_score": 0.25,
      "risk_adjusted_score": 0.25,
      "drawdown_score": 0.20,
      "consistency_score": 0.15,
      "duration_score": 0.10,
      "turnover_score": 0.05
    }
  }
}
```

---

## Acceptance Criteria

The implementation is complete when:

1. Backtest summary includes `benchmarks`.
2. Backtest summary includes `score`.
3. `XAUUSD` is supported as a `commodity_spot` benchmark.
4. Fund benchmarks use adjusted close by default.
5. Spot commodity benchmarks use close by default.
6. Cash benchmark returns 0% in v0.1.
7. Raw AlphaNet Score is calculated for any valid backtest.
8. Official score is null unless official eligibility checks pass.
9. Score components are individually emitted.
10. Benchmark and score logic is deterministic.
11. Missing benchmark data produces a clear warning or error.
12. Unit tests cover benchmark return calculation.
13. Unit tests cover score component normalization.
14. Unit tests cover official scoring eligibility statuses.

---

## Suggested Unit Tests

### Benchmark Tests

- `SPY` fund benchmark calculates adjusted-close total return.
- `XAUUSD` commodity benchmark calculates close-to-close total return.
- `CASH` benchmark returns zero total and annualized return.
- missing benchmark data returns a clear error.
- invalid starting value returns a clear error.

### Score Tests

- negative return produces low return score.
- high annualized return caps return score at 1.
- Sharpe and Sortino scores clamp to 1.
- max drawdown of 50% produces drawdown score 0.
- 10+ year run produces duration score 1.
- high annual turnover reduces turnover score.
- NaN Sharpe or Sortino is handled safely.
- raw score is always between 0 and 100.

### Official Eligibility Tests

- user-selected date range returns `not_official_window`.
- less than 3 years returns `insufficient_history`.
- excluded ranges return `uses_excluded_ranges`.
- non-daily valuation returns `missing_required_valuation` when daily valuation is required.
- accepted official suite returns `official`.

---

## Per-Provider Dolt Databases

AlphaNet v1 uses one DoltHub database per public data source.

Do not implement one combined database for all sources.

Provider databases:

```text
alphanet_alpaca
alphanet_fred
alphanet_cboe
alphanet_imf_portwatch
alphanet_alpha_vantage
alphanet_yahoo_finance
```

Each source has a separate connection and a source-specific schema under:

```text
backtester/data-providers/<provider>/schema.sql
```

The backtester normalizes provider-specific rows in Go code through provider adapters and `CompositeProvider`.

See:

```text
docs/data-sources.md
backtester/data-access-plan.md
backtester/data-providers/
```

---

## Provider-Collocated Data Schemas

Each public data source has a single provider folder:

```text
backtester/data-providers/<source>/
```

Each folder contains:

```text
README.md
schema.sql
```

The `schema.sql` is a seed/maintainer schema for initially creating the public DoltHub database. Normal users usually clone the public database and do not run table creation statements.

Public DoltHub writes are disabled by default and must be enabled explicitly by config or environment variables.

Use of a local Dolt clone is configurable per source so local and cloud runs can choose the lowest-cost access mode.
