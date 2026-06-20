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
    Trace bool
}
```

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
Load market data
    ↓
Initialize portfolio from AIR initial_allocation
    ↓
For each date:
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

Start with CSV/local files.

Future adapters can support:

- FRED
- Yahoo Finance
- FMP
- OECD
- IMF PortWatch
- Parquet
- DuckDB
- SQLite

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
2. Initialize portfolio from `initial_allocation`.
3. Load one day of mock data.
4. Evaluate one or two signals.
5. Evaluate one regime.
6. Evaluate one rule.
7. Emit requested actions.
8. Pass through portfolio engine.
9. Emit decision trace.

No real historical backtest yet.

---

## Second Implementation Milestone

Build multi-day replay.

Tasks:

1. Load date range.
2. Iterate trading days.
3. Evaluate signals daily.
4. Track portfolio state.
5. Track basket exposures.
6. Generate equity curve.
7. Generate decision trace.

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

Start with local CSV files.

Example:

```text
data/
├── prices/
│   ├── QQQ.csv
│   ├── NVDA.csv
│   ├── TLT.csv
│   └── USO.csv
├── macro/
│   ├── WTI.csv
│   ├── UST10Y.csv
│   └── DXY.csv
└── volatility/
    └── VIX.csv
```

CSV fields:

```text
date,open,high,low,close,adjusted_close,volume
```

Macro CSV fields:

```text
date,value
```

Candidate basket symbols should use the same price file format as tradable symbols.

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
