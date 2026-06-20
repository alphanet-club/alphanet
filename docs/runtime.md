# AlphaNet Runtime

## Overview

AlphaNet runtime behavior is split into two distinct runtimes:

1. Compilation runtime
2. Backtest runtime

These runtimes have different responsibilities and different determinism requirements.

---

## Runtime Separation

```text
Compilation Runtime
    - may use LLMs
    - may use agent engines
    - may use external data
    - may be expensive
    - may be non-deterministic

Backtest Runtime
    - must not use LLMs
    - must not use agent engines
    - must be deterministic
    - must be reproducible
    - must consume only AIR and market data
```

This separation is the core performance optimization in AlphaNet.

---

## Compilation Runtime

The compilation runtime is responsible for producing:

```text
compiled/strategy.ir.json
```

from:

```text
manifest.json
strategy.md
rules.json
```

The compiler may also produce:

```text
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
compiled/compile-report.json
```

---

## Compiler Inputs

### `manifest.json`

Defines:

- strategy name
- spec version
- compiler mode
- compiler engines
- training window
- schedule
- hosted/local compute preferences

### `strategy.md`

Defines human-readable strategy intent.

### `rules.json`

Defines user-provided seed rules.

---

## Compiler Modes

### `none`

No compilation is required.

A valid `compiled/strategy.ir.json` already exists and should be used directly.

This is useful when a user compiles locally and provides the final AIR artifact.

---

### `manual`

The compiler validates and normalizes user-provided rules without calling external agent engines.

This mode is useful for deterministic hand-authored strategies.

---

### `single`

The compiler uses one configured engine.

Example:

```json
{
  "compiler": {
    "mode": "single",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5"
      }
    ]
  }
}
```

---

### `ensemble`

The compiler uses multiple engines.

Example:

```json
{
  "compiler": {
    "mode": "ensemble",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "weight": 0.5
      },
      {
        "name": "virattt/ai-hedge-fund",
        "version": "2026.6.17",
        "weight": 0.5
      }
    ],
    "ensemble_method": "weighted_vote"
  }
}
```

Supported ensemble methods may include:

- union
- intersection
- weighted vote
- priority order
- LLM judge
- human review

---

## Training Windows

A training window defines what historical data the compiler may inspect while generating AIR.

Examples:

```json
{
  "training_window": {
    "lookback_days": 365
  }
}
```

or:

```json
{
  "training_window": {
    "start": "2020-01-01",
    "end": "2025-12-31"
  }
}
```

The training window constrains the agent reasoning phase.

The backtest date range may be different.

---

## Local Compilation

Users may compile strategies locally.

Example command:

```bash
alphanet compile strategies/oil-rates-growth-tech
```

Local compilation may use:

- user API keys
- local models
- local agent installations
- local data caches
- user-selected training windows

If local compilation succeeds, the user can provide:

```text
compiled/strategy.ir.json
```

directly.

---

## Hosted Compilation

A hosted AlphaNet system may compile strategies for users.

Hosted compilation may consume compute credits based on:

- selected engines
- number of symbols
- training window size
- LLM usage
- external agent runtime costs
- schedule frequency

The output is still:

```text
compiled/strategy.ir.json
```

Hosted compilation does not change the backtester contract.

---

## Scheduled Compilation

Strategies may be recompiled periodically.

Supported schedule types may include:

- manual
- daily
- weekly
- monthly
- quarterly
- event-driven

Event-driven examples:

- FOMC
- CPI
- earnings season
- volatility spike
- drawdown threshold
- regime change

Scheduled recompilation updates AIR.

Backtesting of a given AIR remains deterministic.

---

## Backtest Runtime

The backtest runtime evaluates AIR over a date range.

Example command:

```bash
alphanet backtest strategies/oil-rates-growth-tech/compiled/strategy.ir.json \
  --start 2018-01-01 \
  --end 2025-12-31
```

The backtester should:

1. Validate AIR.
2. Load historical data.
3. Initialize portfolio state.
4. Evaluate signals.
5. Evaluate regimes.
6. Evaluate relations.
7. Evaluate rules.
8. Resolve conflicts.
9. Enforce portfolio constraints.
10. Simulate execution.
11. Record decision trace.
12. Emit results.

---

## Forbidden Backtester Behavior

The backtester must not:

- call OpenAI
- call Claude
- call local LLMs
- call TradingAgents
- call AI Hedge Fund
- read `strategy.md`
- reinterpret user intent
- mutate `strategy.ir.json`
- silently ignore invalid references
- use non-versioned data for verified results

---

## Runtime Data Dependencies

The backtester needs market data for all signal references.

Example signal:

```json
{
  "signal_id": "wti_change_20d",
  "family": "macro",
  "instrument": "WTI",
  "transform": "percent_change",
  "window": "20d"
}
```

The backtester must resolve that signal using configured data adapters.

Potential adapters:

- FRED
- Yahoo Finance
- FMP
- OECD
- IMF PortWatch
- local CSV
- local Parquet
- custom data providers

---

## Backtest Outputs

Recommended output files:

```text
backtests/
├── YYYY-MM-DD_YYYY-MM-DD.summary.json
├── YYYY-MM-DD_YYYY-MM-DD.equity-curve.csv
├── YYYY-MM-DD_YYYY-MM-DD.trades.csv
└── YYYY-MM-DD_YYYY-MM-DD.decision-trace.jsonl
```

---

## Determinism Requirements

The same inputs should produce the same results.

Required inputs:

- AIR file
- AIR hash
- spec version
- backtester version
- data version
- date range
- execution assumptions

The compiler version is useful for provenance, but the compiler does not need to be rerun to reproduce a backtest.

---

## Runtime Error Handling

The runtime should fail clearly when:

- AIR schema validation fails
- a signal reference cannot be resolved
- a regime references a missing signal
- a rule references a missing regime
- an action target is not found
- portfolio constraints are impossible
- data is missing for required dates
- execution assumptions are invalid

Silent failure should be avoided.

---

## Repository Placement

This document belongs at:

```text
docs/runtime.md
```


---

## Runtime Portfolio Initialization

At backtest startup, the runtime must initialize the portfolio from `portfolio.initial_allocation`.

Startup sequence:

```text
1. Read starting_cash.
2. Read initial_allocation.
3. Convert weights, dollars, or shares into starting positions.
4. Validate starting positions against universe and constraints.
5. Begin evaluating signals, regimes, relations, and rules.
```

If no `initial_allocation` is provided, the runtime should default to 100% cash.

---

## Runtime Candidate Selection

Candidate baskets define assets the strategy may buy later.

The runtime should not assume that the initial allocation is the complete tradable universe.

Instead:

```text
initial_allocation = current starting holdings
candidate_baskets = securities available for future selection
selection_policy = deterministic logic for choosing among candidates
```

A compiled AIR file may therefore start with `SPY`, `QQQ`, `TLT`, and cash, while still allowing later purchases of `NVDA`, `AMD`, `XLU`, `XLP`, or other basket members.

---

## Runtime Recompilation Model

If AIR is regenerated on a higher frequency, the compiler can update candidate baskets and selection policies using agent intelligence.

The backtester remains deterministic because each run consumes a fixed `strategy.ir.json`.

For live or rolling simulations, a future runtime may support dated AIR versions:

```text
2026-01-01 strategy.ir.json
2026-02-01 strategy.ir.json
2026-03-01 strategy.ir.json
```

Each AIR version is deterministic over its own effective period.

---

## Runtime Source Boundaries

At compile time, the compiler reads:

```text
manifest.json
strategy.md
rules.json
```

`manifest.json` provides strategy configuration, including the portfolio model and candidate universe. `strategy.md` provides narrative strategy intent. `rules.json` provides seed rules.

At backtest time, the backtester reads:

```text
compiled/strategy.ir.json
```

The compiled AIR file contains the normalized portfolio model, including initial allocation, candidate baskets, selection policy, targets, and constraints. The backtester should not inspect source files to reconstruct portfolio configuration.

---

## Sampling Policies

AlphaNet runtime supports two separate sampling concepts:

```text
decision_sampling = when rules and trades may happen
valuation_frequency = how often portfolio value and risk are measured
```

A strategy may trade weekly while still being valued daily.

Example:

```json
{
  "decision_sampling": {
    "frequency": "weekly",
    "anchor": "friday",
    "calendar": "trading",
    "missing_date_policy": "nearest_previous"
  },
  "valuation_frequency": "daily"
}
```

Supported frequencies:

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

The compiler can also use sampling in its training window so that a large historical range can be analyzed more quickly.

---

## Benchmark Results and Scoring

The backtester should calculate configured benchmark returns over the same backtest period as the strategy.

For fund and index benchmarks:

```text
total_return = ending_adjusted_close / starting_adjusted_close - 1
```

For spot commodities such as `XAUUSD`:

```text
total_return = ending_spot_price / starting_spot_price - 1
```

For `CASH`, v0.1 may use:

```text
total_return = 0
annualized_return = 0
```

AlphaNet Score should be emitted in the backtest summary as a 0-100 composite score. It should include both a local `raw_score` and, when eligible, an `official_score`.

The score is based on return, Sharpe/Sortino, max drawdown, consistency across windows, duration tested, and turnover realism.
