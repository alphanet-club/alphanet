# AlphaNet Rule Engine

## Overview

The rule engine evaluates compiled AlphaNet rules against market state, regime state, relation state, and portfolio state.

Rules do not execute trades directly.

Rules emit requested actions.

The portfolio engine then approves, scales, rejects, or modifies those actions.

---

## Rule Evaluation Flow

```text
Signals
  ↓
Regimes
  ↓
Relations
  ↓
Rules
  ↓
Requested Actions
  ↓
Decision Hierarchy
  ↓
Portfolio Engine
  ↓
Approved Orders
```

---

## Rule Structure

A rule contains:

- `rule_id`
- `layer`
- `priority`
- `confidence`
- `when`
- `then`
- optional validity windows
- optional cooldowns
- optional limits
- optional explanation metadata

Example:

```json
{
  "rule_id": "reduce_growth_tech_when_oil_and_rates_rise",
  "layer": "strategy",
  "priority": 70,
  "confidence": 0.76,
  "when": {
    "all": [
      {
        "signal": "wti_change_20d",
        "operator": ">",
        "value": 0.10
      },
      {
        "signal": "ust10y_change_20d",
        "operator": ">",
        "value": 25,
        "unit": "basis_points"
      },
      {
        "regime": "tight_liquidity",
        "operator": "active"
      }
    ]
  },
  "then": [
    {
      "action": "decrease_weight",
      "target": "QQQ",
      "amount": 0.05
    },
    {
      "action": "increase_weight",
      "target": "cash",
      "amount": 0.05
    }
  ]
}
```

---

## Conditions

Rules may reference:

- signals
- regimes
- relations
- portfolio metrics
- events
- dates

---

## Logical Operators

Supported logical operators:

- `all`
- `any`
- `not`

Example:

```json
{
  "all": [
    {
      "signal": "wti_change_20d",
      "operator": ">",
      "value": 0.10
    },
    {
      "regime": "tight_liquidity",
      "operator": "active"
    }
  ]
}
```

---

## Comparison Operators

Supported comparison operators include:

- `>`
- `>=`
- `<`
- `<=`
- `==`
- `!=`
- `between`
- `in`
- `not_in`
- `active`
- `inactive`
- `exists`
- `missing`
- `crosses_above`
- `crosses_below`

---

## Rule Layers

Rules belong to layers.

Default layers:

| Layer | Default Priority |
|---|---:|
| Portfolio Safety | 100 |
| Risk Management | 90 |
| Regime Rules | 80 |
| Cross-Asset Relation Rules | 70 |
| Strategy Rules | 60 |
| Tactical Rules | 50 |
| Experimental Overlays | 25 |

Layer priority is evaluated before individual rule priority.

---

## Individual Rule Priority

Every rule may define its own priority within its layer.

Example:

```json
{
  "rule_id": "buy_semis_on_positive_ai_sentiment",
  "layer": "strategy",
  "priority": 45
}
```

Another strategy-layer rule may have higher priority:

```json
{
  "rule_id": "reduce_growth_tech_when_oil_and_rates_rise",
  "layer": "strategy",
  "priority": 70
}
```

If both conflict within the same layer, the higher-priority rule wins.

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

Example:

```text
Risk Management rule:
REDUCE NVDA
layer priority 90
rule priority 40

Strategy rule:
BUY NVDA
layer priority 60
rule priority 90

Result:
REDUCE NVDA
```

Layer priority wins first.

---

## Confidence

Rules may define confidence.

Confidence is used after layer priority and rule priority.

Example:

```json
{
  "rule_id": "reduce_growth_tech",
  "priority": 70,
  "confidence": 0.76
}
```

Confidence is not a substitute for priority.

---

## Tie Breakers

If layer priority, rule priority, and confidence are equal, the engine must use a deterministic tie breaker.

Possible tie breakers:

- rule order in AIR
- lexicographic rule id
- most conservative action
- explicit tie-break field

The selected tie breaker should be declared in AIR.

---

## Actions

Rules emit requested actions.

Supported action types include:

- `increase_weight`
- `decrease_weight`
- `set_weight`
- `buy`
- `sell`
- `hold`
- `rebalance`
- `raise_cash`
- `reduce_risk`
- `block_trade`
- `allow_trade`
- `set_regime_bias`
- `emit_signal`
- `custom`

Example:

```json
{
  "action": "decrease_weight",
  "target": "QQQ",
  "amount": 0.05,
  "unit": "weight"
}
```

---

## Action Targets

Targets may be:

- symbols
- asset classes
- sectors
- themes
- factors
- exposure buckets
- portfolio
- cash

Examples:

```json
{
  "target": "QQQ"
}
```

```json
{
  "target": "cash"
}
```

```json
{
  "target": "growth_technology"
}
```

---

## Requested Actions vs Approved Actions

The rule engine produces requested actions.

The portfolio engine produces approved actions.

Example requested action:

```json
{
  "action": "increase_weight",
  "target": "NVDA",
  "amount": 0.10
}
```

The portfolio engine may approve:

```json
{
  "action": "increase_weight",
  "target": "NVDA",
  "amount": 0.03,
  "scaled": true
}
```

or reject:

```json
{
  "action": "increase_weight",
  "target": "NVDA",
  "amount": 0.10,
  "approved": false,
  "reason": "max_single_position exceeded"
}
```

---

## Cooldowns

Rules may define cooldowns to prevent repeated firing.

Example:

```json
{
  "cooldown": {
    "enabled": true,
    "days": 5,
    "scope": "rule"
  }
}
```

Cooldown scopes:

- rule
- target
- portfolio

---

## Rule Limits

Rules may define limits.

Examples:

```json
{
  "limits": {
    "max_fire_count_per_year": 12,
    "max_single_action_weight": 0.05,
    "requires_portfolio_approval": true
  }
}
```

---

## Validity Windows

Rules may be valid only during certain dates.

Example:

```json
{
  "valid_from": "2020-01-01",
  "valid_to": "2025-12-31"
}
```

If omitted, the rule is valid for the full backtest period.

---

## Rule Evaluation Algorithm

A simple deterministic evaluation loop:

```text
For each date:
    Evaluate all signals
    Evaluate regimes
    Evaluate relations
    For each enabled rule:
        Check validity window
        Check cooldown
        Evaluate condition
        If condition is true:
            Emit requested actions
    Resolve action conflicts
    Send actions to portfolio engine
```

---

## Conflict Examples

### Buy vs Sell Same Asset

Rule A requests:

```text
BUY NVDA
```

Rule B requests:

```text
SELL NVDA
```

Resolution uses:

1. layer priority
2. rule priority
3. confidence
4. tie breaker

---

### Increase vs Decrease Same Weight

Rule A:

```text
increase QQQ by 5%
```

Rule B:

```text
decrease QQQ by 3%
```

Possible deterministic approaches:

- net the actions
- choose higher-priority action
- choose most conservative action
- pass both to portfolio engine with conflict metadata

For v0.1, the recommended approach is:

```text
Resolve by priority first.
If same priority, use tie breaker.
```

---

## Decision Trace

The rule engine should record:

- evaluated rules
- matched rules
- unmatched rules, optionally
- requested actions
- conflicts
- resolution path

Example:

```json
{
  "date": "2026-06-19",
  "rules_evaluated": [
    "reduce_growth_tech_when_oil_and_rates_rise",
    "high_volatility_reduce_risk"
  ],
  "rules_triggered": [
    "reduce_growth_tech_when_oil_and_rates_rise"
  ],
  "actions_requested": [
    {
      "action": "decrease_weight",
      "target": "QQQ",
      "amount": 0.05
    }
  ]
}
```

---

## Implementation Guidance

The rule engine should be:

- deterministic
- schema-driven
- reference-safe
- explainable
- independent of LLMs
- independent of source strategy files

The rule engine should fail clearly when:

- a signal reference is missing
- a regime reference is missing
- an operator is unsupported
- an action is invalid
- a rule id is duplicated

---

## Repository Placement

This document belongs at:

```text
docs/rule-engine.md
```
