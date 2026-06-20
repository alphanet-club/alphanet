# AlphaNet Portfolio Engine

## Overview

The portfolio engine is the final authority in AlphaNet.

Rules may request trades.

Relations may imply pressure.

Regimes may alter risk appetite.

But the portfolio engine decides what can actually be executed.

The portfolio engine exists to ensure that strategies remain:

- portfolio-aware
- constraint-aware
- risk-aware
- reproducible
- explainable

---

## Responsibilities

The portfolio engine is responsible for:

1. Tracking portfolio state.
2. Evaluating allocation targets.
3. Enforcing hard constraints.
4. Managing soft target drift.
5. Applying risk budgets.
6. Scaling or rejecting requested actions.
7. Producing approved orders.
8. Recording portfolio decisions in the decision trace.

---

## Portfolio State

At any point in a backtest, the portfolio engine should know:

- cash
- positions
- market value
- total portfolio value
- asset weights
- asset class weights
- sector weights
- theme weights
- leverage
- turnover
- realized PnL
- unrealized PnL
- drawdown
- volatility estimate

Example state:

```json
{
  "date": "2026-06-19",
  "portfolio_value": 103125.42,
  "cash": 25781.35,
  "cash_weight": 0.25,
  "positions": {
    "QQQ": {
      "market_value": 21000,
      "weight": 0.2036
    },
    "TLT": {
      "market_value": 20500,
      "weight": 0.1988
    }
  }
}
```

---

## Starting Capital

AlphaNet examples default to:

```json
{
  "starting_cash": 100000,
  "base_currency": "USD"
}
```

A backtest configuration may override starting capital if allowed.

---

## Allocation Targets

Targets describe desired portfolio allocation.

Example:

```json
{
  "targets": {
    "cash": 0.25,
    "equities": 0.50,
    "bonds": 0.20,
    "commodities": 0.05
  }
}
```

Targets may apply to:

- asset classes
- symbols
- sectors
- themes
- factor exposures
- custom buckets

Targets are not necessarily hard constraints.

---

## Hard Constraints

Hard constraints must not be violated.

Examples:

```json
{
  "cash_min": 0.10,
  "max_single_position": 0.15,
  "max_leverage": 1.0,
  "allow_shorting": false,
  "allow_margin": false
}
```

Hard constraints can block, scale, or modify requested actions.

---

## Soft Targets

Soft targets allow drift.

Example:

```json
{
  "target": "equities",
  "weight": 0.50,
  "tolerance": 0.10
}
```

This means equities are allowed to drift between:

```text
40% and 60%
```

before a corrective rebalance is required.

---

## Risk Budgets

Risk budgets define portfolio-level risk limits.

Examples:

```json
{
  "max_portfolio_volatility": 0.20,
  "max_drawdown": 0.15,
  "max_daily_turnover": 0.10,
  "max_trade_weight": 0.05
}
```

Risk budgets can cause the portfolio engine to:

- reduce order sizes
- reject new risk
- raise cash
- rebalance toward targets
- block trades
- enter a cooldown period

---

## Portfolio Safety Layer

Portfolio safety rules should usually have the highest layer priority.

Default priority:

```text
100
```

Portfolio safety may override:

- risk management rules
- regime rules
- strategy rules
- tactical rules
- experimental overlays

Example:

```text
Strategy rule wants to buy NVDA.
Portfolio safety rule says max technology weight is already exceeded.
Result: buy is rejected or scaled down.
```

---

## Action Review Process

The portfolio engine should process requested actions like this:

```text
Requested Actions
        ↓
Normalize Actions
        ↓
Check Hard Constraints
        ↓
Check Risk Budgets
        ↓
Check Soft Targets
        ↓
Scale / Reject / Approve
        ↓
Generate Orders
        ↓
Record Decision Trace
```

---

## Action Normalization

Rules may emit abstract actions:

```json
{
  "action": "decrease_weight",
  "target": "QQQ",
  "amount": 0.05,
  "unit": "weight"
}
```

The portfolio engine converts these into executable order intent:

```json
{
  "side": "sell",
  "symbol": "QQQ",
  "target_weight_delta": -0.05
}
```

Actual shares and prices are determined by the execution simulator.

---

## Scaling Actions

If a requested action partially violates constraints, the engine may scale it.

Example:

Requested:

```json
{
  "action": "increase_weight",
  "target": "NVDA",
  "amount": 0.10
}
```

Constraint:

```json
{
  "max_single_position": 0.15
}
```

Current NVDA weight:

```text
12%
```

Maximum allowed increase:

```text
3%
```

Approved action:

```json
{
  "action": "increase_weight",
  "target": "NVDA",
  "amount": 0.03,
  "scaled": true
}
```

---

## Rejecting Actions

An action should be rejected when it cannot be safely scaled or reconciled.

Example:

```json
{
  "action": "buy",
  "target": "NVDA",
  "reason_rejected": "Technology sector max weight exceeded."
}
```

Rejected actions should be included in the decision trace.

---

## Rebalancing

Rebalancing may be:

- daily
- weekly
- monthly
- quarterly
- event-driven
- threshold-driven

Example rebalance policy:

```json
{
  "frequency": "daily",
  "threshold": 0.05,
  "allow_partial_rebalance": true,
  "min_trade_weight": 0.005
}
```

---

## Cash Handling

Cash is a first-class allocation.

Rules can request:

```json
{
  "action": "increase_weight",
  "target": "cash",
  "amount": 0.05
}
```

The portfolio engine should treat cash as both:

- an asset class
- a safety buffer

Minimum cash constraints should be enforced before new buys.

---

## Drawdown Handling

Drawdown-based policies may trigger de-risking.

Example:

```json
{
  "max_drawdown": 0.15,
  "de_risking_policy": {
    "enabled": true,
    "trigger": "drawdown",
    "action": "raise_cash",
    "amount": 0.10,
    "cooldown_days": 10
  }
}
```

If max drawdown is breached, the portfolio engine may raise cash and prevent new risk for a cooldown period.

---

## Decision Trace

The portfolio engine should emit trace information.

Example:

```json
{
  "date": "2026-06-19",
  "requested_actions": [
    {
      "action": "increase_weight",
      "target": "NVDA",
      "amount": 0.10
    }
  ],
  "approved_actions": [
    {
      "action": "increase_weight",
      "target": "NVDA",
      "amount": 0.03,
      "scaled": true
    }
  ],
  "rejected_actions": [],
  "constraints_checked": [
    "max_single_position",
    "max_sector_weight",
    "cash_min"
  ]
}
```

---

## Implementation Guidance

The portfolio engine should be deterministic.

Given the same:

- portfolio state
- requested actions
- constraints
- prices
- execution assumptions

it should produce the same approved actions.

Avoid hidden randomness.

Avoid using LLMs.

Avoid implicit state not recorded in backtest output.

---

## Repository Placement

This document belongs at:

```text
docs/portfolio-engine.md
```
