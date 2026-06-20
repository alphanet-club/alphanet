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
Initialize portfolio
    ↓
For each date:
    Evaluate signals
    Evaluate regimes
    Evaluate relations
    Evaluate rules
    Resolve conflicts
    Apply portfolio constraints
    Simulate execution
    Update portfolio
    Record trace
    Record metrics
    ↓
Write outputs
```

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

## Rule Engine

The rule engine evaluates rules using:

- signal values
- regime states
- relation states
- portfolio state
- date state
- event state

It emits requested actions.

The rule engine should not enforce portfolio constraints.

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

1. Group actions by target.
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
2. Load one day of mock data.
3. Evaluate one or two signals.
4. Evaluate one regime.
5. Evaluate one rule.
6. Emit requested actions.
7. Pass through portfolio engine.
8. Emit decision trace.

No real historical backtest yet.

---

## Second Implementation Milestone

Build multi-day replay.

Tasks:

1. Load date range.
2. Iterate trading days.
3. Evaluate signals daily.
4. Track portfolio state.
5. Generate equity curve.
6. Generate decision trace.

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

---

## Fifth Implementation Milestone

Add schema and semantic validation.

Tasks:

1. Validate AIR schema.
2. Validate references.
3. Validate universe symbols.
4. Validate portfolio constraints.
5. Validate rule targets.
6. Validate signal definitions.

---

## Testing Plan

### Unit Tests

- signal transforms
- condition evaluation
- regime activation
- relation activation
- rule matching
- conflict resolution
- portfolio constraints
- execution fills
- metrics

### Integration Tests

- example AIR runs without error
- expected rule fires on known date
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
2. It runs over a date range.
3. It evaluates signals deterministically.
4. It evaluates rules deterministically.
5. It enforces portfolio constraints.
6. It emits trades, equity curve, summary, and decision trace.
7. The same inputs produce the same outputs.
