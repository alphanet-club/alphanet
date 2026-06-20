# AlphaNet Specification v0.1

**Status:** Draft
**Spec Version:** v0.1
**Primary Artifact:** `strategy.ir.json`
**IR Name:** AlphaNet Intermediate Representation (AIR)

---

## 1. Purpose

AlphaNet is a market intelligence runtime for compiling human-authored strategy intent and optional agent feedback into a deterministic strategy artifact.

This specification defines AlphaNet v0.1, including:

- strategy authoring inputs
- rules compiler behavior
- AlphaNet Intermediate Representation (AIR)
- portfolio semantics
- signal references
- regime definitions
- relations
- decision hierarchy
- rule evaluation
- backtester requirements
- reproducibility requirements
- versioning expectations

The core AlphaNet design principle is:

> Market reasoning may be expensive and agent-assisted, but market evaluation must be deterministic.

---

## 2. Scope

This document defines the v0.1 AlphaNet contract.

It covers:

- `manifest.json`
- `strategy.md`
- `rules.json`
- `compiled/strategy.ir.json`
- `compiled/provenance.json`
- `compiled/reasoning.md`
- `compiled/validation-report.json`
- deterministic backtesting semantics
- versioning and reproducibility rules

It does not fully define:

- live trading integration
- broker APIs
- leaderboard website behavior
- hosted compute billing
- production security model
- complete market data adapters

Those are future extensions.

---

## 3. Terminology

### AlphaNet

The overall system for compiling, validating, and backtesting strategies.

### AIR

AlphaNet Intermediate Representation.

AIR is the compiled strategy representation stored as:

```text
compiled/strategy.ir.json
```

### Rules Compiler

A Go program that reads source strategy files and emits AIR.

The compiler may call LLMs or agent engines.

### Backtester

A Go program that evaluates AIR against historical market data.

The backtester must not call LLMs or agent engines.

### Strategy Source

The human-authored strategy inputs:

```text
manifest.json
strategy.md
rules.json
```

### Strategy Artifact

The compiled deterministic strategy output:

```text
compiled/strategy.ir.json
```

### Signal

A named observation, derived metric, or market state input.

### Relation

A structured belief linking one or more drivers to targets.

### Regime

A summarized market environment inferred from signals and relations.

### Rule

A conditional decision object that consumes signals, relations, regimes, and portfolio state to propose actions.

### Portfolio Engine

The subsystem that tracks state, enforces constraints, and approves or modifies actions.

### Decision Trace

A record explaining why the engine made a decision on a given date.

---

## 4. System Architecture

AlphaNet has three phases:

```text
Authoring
    ↓
Compilation
    ↓
Evaluation
```

### 4.1 Authoring

A strategy author creates:

```text
manifest.json
strategy.md
rules.json
```

These files define:

- metadata
- strategy intent
- user-provided rules
- compiler configuration
- training window
- agent engines
- portfolio targets
- constraints
- execution preferences

### 4.2 Compilation

The rules compiler reads source files and emits:

```text
compiled/strategy.ir.json
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
```

The compiler may call:

- LLMs
- local models
- TradingAgents
- AI Hedge Fund
- custom agent engines
- data analysis tools

### 4.3 Evaluation

The backtester reads only:

```text
compiled/strategy.ir.json
```

and historical market data.

The backtester emits:

- trades
- equity curve
- daily portfolio states
- performance summary
- decision trace
- validation output

The backtester must be deterministic.

---

## 5. Repository Layout

The recommended AlphaNet repository layout is:

```text
AlphaNet/
│
├── README.md
│
├── specs/
│   └── v0.1/
│       ├── SPEC.md
│       ├── CHANGELOG.md
│       ├── alphanet.schema.json
│       ├── manifest.schema.json
│       ├── portfolio.schema.json
│       ├── signal.schema.json
│       ├── rule.schema.json
│       └── regime.schema.json
│
├── rules-compiler/
│
├── backtester/
│
├── docs/
│   ├── architecture.md
│   ├── runtime.md
│   ├── portfolio-engine.md
│   ├── rule-engine.md
│   └── examples/
│       └── v0.1/
│
└── strategies/
    └── example_strategy/
        ├── manifest.json
        ├── strategy.md
        ├── rules.json
        │
        ├── compiled/
        │   ├── strategy.ir.json
        │   ├── provenance.json
        │   ├── reasoning.md
        │   └── validation-report.json
        │
        └── backtests/
```

---

## 6. Strategy Folder Structure

A strategy folder should contain:

```text
strategy/
│
├── manifest.json
├── strategy.md
├── rules.json
│
├── compiled/
│   ├── strategy.ir.json
│   ├── provenance.json
│   ├── reasoning.md
│   └── validation-report.json
│
└── backtests/
```

### 6.1 `manifest.json`

Defines strategy metadata and compiler configuration.

### 6.2 `strategy.md`

Human-readable strategy intent.

The compiler may use this as context.

### 6.3 `rules.json`

User-provided initial rules.

These are not necessarily final.

The compiler may validate, refine, merge, reject, or expand them.

### 6.4 `compiled/strategy.ir.json`

The canonical compiled AlphaNet strategy artifact.

The backtester consumes this file.

### 6.5 `compiled/provenance.json`

Describes how the IR was generated.

### 6.6 `compiled/reasoning.md`

Human-readable compilation explanation.

### 6.7 `compiled/validation-report.json`

Schema and semantic validation results.

### 6.8 `backtests/`

Stores deterministic backtest outputs.

---

## 7. Manifest Contract

The manifest defines strategy metadata and compilation behavior.

A v0.1 manifest should support:

```json
{
  "name": "Oil Rates Strategy",
  "description": "Reduce growth exposure when oil and long rates rise together.",
  "author": "example",
  "spec_version": "v0.1",
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
    "ensemble_method": "weighted_vote",
    "training_window": {
      "lookback_days": 365
    }
  }
}
```

### 7.1 Required Metadata

A manifest should include:

- `name`
- `spec_version`

Recommended fields:

- `description`
- `author`
- `tags`
- `created_at`
- `updated_at`
- `license`

### 7.2 Strategy Identity

The manifest name and metadata may later be used by leaderboard or strategy registry systems.

### 7.3 Compiler Mode

Supported v0.1 compiler modes:

- `none`
- `single`
- `ensemble`
- `manual`

#### `none`

No compilation should be performed because a valid `strategy.ir.json` already exists.

#### `single`

One engine is used.

#### `ensemble`

Multiple engines may be used.

#### `manual`

The compiler validates and normalizes human-provided rules without calling external agents.

### 7.4 Compiler Engines

Each engine should define:

- `name`
- `version`

Optional:

- `weight`
- `enabled`
- `config`
- `symbols`
- `notes`

Example engines:

```json
[
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
]
```

### 7.5 Ensemble Methods

Possible v0.1 ensemble methods:

- `union`
- `intersection`
- `weighted_vote`
- `priority_order`
- `llm_judge`
- `human_review`

### 7.6 Training Window

The training window bounds what the compiler may analyze while generating AIR.

Supported forms:

```json
{
  "lookback_days": 365
}
```

or:

```json
{
  "start": "2020-01-01",
  "end": "2025-12-31"
}
```

The compiler may use this window to analyze:

- selected market days
- historical regimes
- earnings periods
- volatility regimes
- macro events
- agent outputs
- user-provided rules

The backtester may later evaluate the compiled strategy against a different date range.

---

## 8. Compilation Contract

### 8.1 Inputs

The rules compiler accepts:

```text
manifest.json
strategy.md
rules.json
```

It may also accept optional market data, local cache files, and configured credentials.

### 8.2 Outputs

The compiler writes:

```text
compiled/strategy.ir.json
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
```

Optional outputs:

```text
compiled/compile-report.json
compiled/agent-output/
compiled/debug/
```

### 8.3 Compiler Responsibilities

The compiler should:

1. Read the manifest.
2. Validate spec version.
3. Read strategy description.
4. Read user-provided rules.
5. Resolve compiler mode.
6. Load or query configured engines.
7. Analyze the training window.
8. Generate AIR.
9. Validate AIR against schemas.
10. Produce provenance.
11. Produce reasoning summary.
12. Produce validation report.

### 8.4 Compiler Freedom

The compiler may use non-deterministic tools.

It may call:

- LLMs
- agent systems
- external data APIs
- custom heuristics

However, once it emits AIR, evaluation must become deterministic.

### 8.5 Compiler Output Is Not a Trade Log

The compiler does not produce final historical trades.

It produces reusable strategy logic.

---

## 9. Backtester Contract

### 9.1 Inputs

The backtester accepts:

```text
compiled/strategy.ir.json
```

and a backtest configuration including:

- start date
- end date
- data source
- data version
- starting capital override, if allowed
- execution assumptions
- slippage assumptions
- transaction cost assumptions

### 9.2 Forbidden Behavior

The backtester must not:

- call LLMs
- call TradingAgents
- call AI Hedge Fund
- read `strategy.md`
- reinterpret strategy intent
- modify AIR
- silently ignore invalid rules

### 9.3 Required Behavior

The backtester must:

1. Validate AIR.
2. Load historical data.
3. Initialize portfolio state.
4. Iterate through each backtest period.
5. Evaluate signal references.
6. Infer regimes if defined.
7. Evaluate relations if defined.
8. Evaluate rules.
9. Resolve conflicts.
10. Apply portfolio constraints.
11. Generate approved orders.
12. Simulate execution.
13. Update portfolio state.
14. Record decision trace.
15. Emit performance results.

### 9.4 Determinism

Given identical:

- AIR
- backtester version
- data version
- date range
- execution assumptions

the result should be identical.

---

## 10. AlphaNet Intermediate Representation

AIR is the compiled strategy format.

A v0.1 AIR file should contain:

```json
{
  "metadata": {},
  "universe": {},
  "signals": [],
  "relations": [],
  "regimes": [],
  "portfolio": {},
  "decision_hierarchy": {},
  "rules": [],
  "execution": {}
}
```

### 10.1 Metadata

Metadata should include:

```json
{
  "strategy_name": "Oil Rates Strategy",
  "spec_version": "v0.1",
  "compiler_version": "v0.1.0",
  "generated_at": "2026-06-19T00:00:00Z"
}
```

Recommended metadata:

- strategy id
- description
- author
- tags
- source hash
- IR hash
- compiler mode
- engines used
- training window

### 10.2 Universe

The universe defines assets and asset classes referenced by the strategy.

Example:

```json
{
  "assets": [
    {
      "symbol": "QQQ",
      "asset_class": "equities",
      "sector": "technology"
    },
    {
      "symbol": "TLT",
      "asset_class": "bonds",
      "sector": "duration"
    },
    {
      "symbol": "USO",
      "asset_class": "commodities",
      "sector": "energy"
    }
  ]
}
```

### 10.3 Signals

Signals define measurable inputs.

AIR should generally reference signal definitions rather than embed one-time observed market values.

Example:

```json
{
  "signal_id": "wti_change_20d",
  "family": "macro",
  "source": "market_data",
  "instrument": "WTI",
  "transform": "percent_change",
  "window": "20d"
}
```

### 10.4 Relations

Relations encode structured beliefs.

Example:

```json
{
  "relation_id": "oil_rates_negative_for_growth_tech",
  "drivers": [
    "wti_change_20d",
    "ust10y_change_20d"
  ],
  "targets": [
    "QQQ",
    "XLK",
    "NVDA"
  ],
  "effect": "negative",
  "confidence": 0.72
}
```

### 10.5 Regimes

Regimes summarize market state.

Example:

```json
{
  "regime_id": "tight_liquidity",
  "conditions": {
    "all": [
      {
        "signal": "fed_funds_rate",
        "operator": ">",
        "value": 3.5
      },
      {
        "signal": "dxy_change_60d",
        "operator": ">",
        "value": 0.03
      }
    ]
  },
  "confidence": 0.81
}
```

### 10.6 Portfolio

Portfolio configuration defines starting capital, targets, and constraints.

Example:

```json
{
  "base_currency": "USD",
  "starting_cash": 100000,
  "targets": {
    "cash": 0.25,
    "equities": 0.25,
    "bonds": 0.25,
    "commodities": 0.25
  },
  "constraints": {
    "cash_min": 0.10,
    "max_single_position": 0.10,
    "max_leverage": 1.0
  }
}
```

### 10.7 Decision Hierarchy

The decision hierarchy defines layer priorities.

Example:

```json
{
  "layers": [
    {
      "name": "portfolio_safety",
      "priority": 100
    },
    {
      "name": "risk_management",
      "priority": 90
    },
    {
      "name": "regime",
      "priority": 80
    },
    {
      "name": "cross_asset_relations",
      "priority": 70
    },
    {
      "name": "strategy",
      "priority": 60
    },
    {
      "name": "tactical",
      "priority": 50
    },
    {
      "name": "experimental",
      "priority": 25
    }
  ]
}
```

### 10.8 Rules

Rules define conditional actions.

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

### 10.9 Execution

Execution settings define how trades are simulated.

Example:

```json
{
  "rebalance_frequency": "daily",
  "order_timing": "next_open",
  "transaction_cost_bps": 1,
  "slippage_bps": 2,
  "allow_fractional_shares": true
}
```

---

## 11. Signal Families

AlphaNet v0.1 recognizes these signal families:

- `valuation`
- `macro`
- `global_macro`
- `fed`
- `shipping`
- `market_flow`
- `events`
- `sentiment`
- `volatility`
- `portfolio`
- `custom`

### 11.1 Valuation

Examples:

- P/E
- CAPE
- earnings yield
- dividend yield
- equity risk premium

### 11.2 Macro

Examples:

- rates
- inflation
- labor
- oil
- dollar
- M2
- SLOOS

### 11.3 Global Macro

Examples:

- country CPI
- short rates
- CLI
- house price indices
- equity indices

### 11.4 Fed

Examples:

- balance sheet
- Treasuries held
- MBS held
- dealer positions
- FOMC documents

### 11.5 Shipping

Examples:

- Suez
- Panama
- Hormuz
- Malacca
- tonnage
- vessel count

### 11.6 Market Flow

Examples:

- gainers
- losers
- most active
- relative volume
- dollar volume
- rotation

### 11.7 Events

Examples:

- earnings
- IPOs
- ex-dividend dates

### 11.8 Sentiment

Examples:

- news sentiment
- social sentiment
- topic sentiment
- narrative acceleration

### 11.9 Volatility

Examples:

- VIX
- realized volatility
- implied volatility
- earnings implied move
- term structure

---

## 12. Conditions

Conditions may reference:

- signals
- regimes
- relations
- portfolio state
- dates
- events

### 12.1 Logical Operators

Supported v0.1 logical operators:

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

### 12.2 Comparison Operators

Supported comparison operators:

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

### 12.3 Portfolio Conditions

Examples:

```json
{
  "portfolio_metric": "equities_weight",
  "operator": ">",
  "value": 0.60
}
```

```json
{
  "portfolio_metric": "cash_weight",
  "operator": "<",
  "value": 0.10
}
```

---

## 13. Actions

Rules produce desired actions.

Supported v0.1 actions:

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

### 13.1 Weight Actions

Example:

```json
{
  "action": "decrease_weight",
  "target": "QQQ",
  "amount": 0.05
}
```

### 13.2 Cash Actions

Example:

```json
{
  "action": "increase_weight",
  "target": "cash",
  "amount": 0.05
}
```

### 13.3 Risk Actions

Example:

```json
{
  "action": "reduce_risk",
  "target": "portfolio",
  "amount": 0.20
}
```

---

## 14. Portfolio Semantics

### 14.1 Starting Cash

If not overridden by the backtest configuration, the portfolio starts with:

```json
{
  "starting_cash": 100000
}
```

### 14.2 Asset Classes

Default asset classes:

- `cash`
- `equities`
- `bonds`
- `commodities`
- `crypto`
- `real_assets`
- `alternatives`
- `custom`

### 14.3 Targets

Targets define desired allocation.

Example:

```json
{
  "cash": 0.25,
  "equities": 0.25,
  "bonds": 0.25,
  "commodities": 0.25
}
```

### 14.4 Hard Constraints

Hard constraints must not be violated.

Examples:

- minimum cash
- maximum leverage
- maximum single position
- maximum sector weight
- maximum asset class weight
- no shorting
- no margin

### 14.5 Soft Constraints

Soft constraints define target ranges.

Example:

```json
{
  "asset_class": "equities",
  "target": 0.50,
  "tolerance": 0.10
}
```

### 14.6 Risk Budgets

Risk budgets may include:

- maximum volatility
- maximum drawdown
- maximum beta
- maximum turnover
- maximum daily trade size

### 14.7 Portfolio Engine Authority

The portfolio engine may:

- approve actions
- reject actions
- scale actions down
- substitute actions
- delay actions
- force rebalancing

The portfolio engine must record its decision in the decision trace.

---

## 15. Decision Hierarchy

AlphaNet supports two priority dimensions:

1. Layer priority
2. Rule priority within a layer

### 15.1 Layer Priorities

Default v0.1 priorities:

| Layer | Priority |
|---|---:|
| Portfolio Safety | 100 |
| Risk Management | 90 |
| Regime Rules | 80 |
| Cross-Asset Relation Rules | 70 |
| Strategy Rules | 60 |
| Tactical Rules | 50 |
| Experimental Overlays | 25 |

### 15.2 Rule Priorities

Rules may define an individual priority.

Example:

```json
{
  "rule_id": "reduce_growth_tech",
  "layer": "strategy",
  "priority": 90
}
```

### 15.3 Conflict Resolution

Conflicts are resolved by:

```text
Layer Priority
    ↓
Rule Priority
    ↓
Confidence
    ↓
Tie Break Logic
```

### 15.4 Layer Priority Wins First

A lower-priority rule inside a higher-priority layer can override a higher-priority rule inside a lower-priority layer.

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

### 15.5 Confidence

Rules may include `confidence`.

Confidence is not a substitute for priority.

It is used after layer priority and rule priority.

### 15.6 Tie Break Logic

If priority and confidence are equal, the engine should use deterministic tie breaking.

Possible tie breakers:

- lexicographic rule id
- rule order in AIR
- explicit `tie_breaker` field
- most conservative action

The selected method must be deterministic.

---

## 16. Regime Semantics

A regime may be:

- active
- inactive
- scored
- probabilistic

v0.1 should support active/inactive regimes and optional confidence scores.

Example:

```json
{
  "regime_id": "risk_off",
  "state": "active",
  "confidence": 0.78
}
```

A regime can be inferred from signals or directly defined by rules.

---

## 17. Relation Semantics

Relations encode directional pressure.

Supported effects:

- `positive`
- `negative`
- `neutral`
- `mixed`
- `risk_on`
- `risk_off`
- `increase_volatility`
- `decrease_volatility`

Relations may target:

- symbols
- asset classes
- sectors
- themes
- regimes
- portfolio exposures

Example:

```json
{
  "relation_id": "strong_dollar_negative_for_commodities",
  "drivers": ["dxy_change_60d"],
  "targets": ["commodities"],
  "effect": "negative"
}
```

---

## 18. Execution Semantics

v0.1 backtesting should define:

- rebalance frequency
- order timing
- transaction costs
- slippage
- fractional shares
- cash handling
- dividend handling
- corporate action assumptions

### 18.1 Rebalance Frequency

Supported values:

- `daily`
- `weekly`
- `monthly`
- `quarterly`
- `event_driven`

### 18.2 Order Timing

Supported values:

- `same_close`
- `next_open`
- `next_close`

Default recommendation:

```json
{
  "order_timing": "next_open"
}
```

### 18.3 Transaction Costs

Transaction costs may be modeled in basis points.

Example:

```json
{
  "transaction_cost_bps": 1
}
```

### 18.4 Slippage

Slippage may be modeled in basis points.

Example:

```json
{
  "slippage_bps": 2
}
```

---

## 19. Decision Trace

The backtester should emit a decision trace.

A decision trace should record:

- date
- portfolio state before decision
- evaluated signals
- active regimes
- triggered relations
- triggered rules
- conflicts
- portfolio constraints
- final approved actions
- rejected actions
- portfolio state after decision

Example:

```json
{
  "date": "2026-06-19",
  "decision": "rebalance",
  "signals": [
    "wti_change_20d",
    "ust10y_change_20d"
  ],
  "regimes": [
    "tight_liquidity"
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
  ],
  "actions_approved": [
    {
      "action": "sell",
      "asset": "QQQ",
      "weight_delta": -0.05
    }
  ]
}
```

---

## 20. Provenance

AIR should have provenance.

The compiler should emit:

```text
compiled/provenance.json
```

Provenance should include:

- strategy name
- spec version
- compiler version
- generated timestamp
- source file hashes
- engine list
- engine versions
- training window
- compiler mode
- validation result
- IR hash

Example:

```json
{
  "strategy_name": "Oil Rates Strategy",
  "spec_version": "v0.1",
  "compiler_version": "v0.1.0",
  "generated_at": "2026-06-19T00:00:00Z",
  "engines": [
    {
      "name": "TauricResearch/TradingAgents",
      "version": "0.2.5"
    },
    {
      "name": "virattt/ai-hedge-fund",
      "version": "2026.6.17"
    }
  ],
  "training_window": {
    "lookback_days": 365
  },
  "ir_sha256": "TODO"
}
```

---

## 21. Validation

The compiler should validate:

- manifest schema
- user rules schema
- AIR schema
- portfolio constraints
- signal references
- rule references
- regime references
- relation references
- action validity
- priority validity
- duplicate ids
- unsupported operators

Validation output should be written to:

```text
compiled/validation-report.json
```

The backtester should also validate AIR before execution.

---

## 22. Backtest Outputs

A backtest should produce:

```text
backtests/
├── YYYY-MM-DD_YYYY-MM-DD.summary.json
├── YYYY-MM-DD_YYYY-MM-DD.equity-curve.csv
├── YYYY-MM-DD_YYYY-MM-DD.trades.csv
└── YYYY-MM-DD_YYYY-MM-DD.decision-trace.jsonl
```

### 22.1 Summary Metrics

Recommended metrics:

- total return
- annualized return
- volatility
- Sharpe ratio
- Sortino ratio
- max drawdown
- turnover
- win rate
- beta
- cash utilization
- average exposure
- exposure by asset class
- exposure by sector

### 22.2 Trades

Trades should include:

- date
- symbol
- side
- quantity
- price
- notional
- transaction cost
- slippage
- reason code
- decision trace id

### 22.3 Equity Curve

Equity curve should include:

- date
- portfolio value
- daily return
- cash
- gross exposure
- net exposure
- drawdown

---

## 23. Reproducibility Requirements

A reproducible result requires:

- AIR hash
- spec version
- backtester version
- data version
- date range
- execution assumptions

A future leaderboard submission may include:

```json
{
  "strategy_name": "Oil Rates Strategy",
  "strategy_ir_sha256": "TODO",
  "spec_version": "v0.1",
  "backtester_version": "v0.1.0",
  "data_version": "TODO",
  "date_range": {
    "start": "2018-01-01",
    "end": "2025-12-31"
  }
}
```

---

## 24. Security and Safety Notes

AlphaNet v0.1 is intended for research, backtesting, and strategy evaluation.

It should not be assumed to provide financial advice.

Implementations should avoid:

- executing untrusted code from strategies
- allowing strategies to call arbitrary APIs during backtesting
- silently accepting invalid AIR
- treating LLM output as deterministic
- using non-versioned datasets for verified results

---

## 25. Future Extensions

Potential future extensions:

- richer JSON schemas
- plugin-based data adapters
- deterministic data snapshots
- sentiment adapters
- volatility surface adapters
- FRED data adapters
- Yahoo Finance adapters
- FMP adapters
- OECD adapters
- IMF PortWatch adapters
- options positioning
- CFTC futures positioning
- gamma exposure
- paper trading
- hosted compilation credits
- strategy packaging
- separate leaderboard repository

---

## 26. Summary

AlphaNet v0.1 defines a clean separation between strategy reasoning and deterministic evaluation.

The rules compiler may use agents, LLMs, and expensive analysis to generate AIR.

The backtester evaluates AIR without LLMs.

The primary artifact is:

```text
compiled/strategy.ir.json
```

This artifact is portable, reproducible, explainable, and suitable for deterministic backtesting.

AlphaNet’s long-term goal is to make strategies:

- portable
- inspectable
- explainable
- reproducible
- portfolio-aware
- agent-compatible
- comparable across a future leaderboard


---

## Portfolio Initialization and Candidate Baskets

AlphaNet distinguishes between three portfolio concepts:

```text
1. Initial allocation
   What the portfolio holds at the start of a backtest or runtime simulation.

2. Candidate baskets
   What the strategy is allowed or interested in buying later.

3. Portfolio targets and constraints
   What the portfolio should look like over time and what it must never violate.
```

### Initial Allocation

`portfolio.initial_allocation` defines the starting holdings.

Example:

```json
{
  "initial_allocation": {
    "mode": "weights",
    "positions": [
      { "symbol": "SPY", "weight": 0.40 },
      { "symbol": "QQQ", "weight": 0.25 },
      { "symbol": "TLT", "weight": 0.20 },
      { "symbol": "cash", "weight": 0.15 }
    ]
  }
}
```

Supported modes:

- `cash`
- `weights`
- `dollars`
- `shares`

If omitted, the runtime should default to 100% cash.

### Candidate Baskets

`portfolio.candidate_baskets` define groups of securities the strategy may buy, sell, rank, or rotate into over time.

Example:

```json
{
  "candidate_baskets": [
    {
      "basket_id": "growth_technology",
      "asset_class": "equities",
      "sector": "technology",
      "symbols": ["QQQ", "NVDA", "AMD", "MSFT", "AAPL", "AVGO", "SMH"],
      "max_weight": 0.35,
      "max_position_weight": 0.10
    }
  ]
}
```

Candidate baskets are part of the compiled AIR contract and are available to the backtester without agent calls.

### Selection Policy

`portfolio.selection_policy` defines deterministic rules for selecting securities inside candidate baskets.

Example:

```json
{
  "selection_policy": {
    "default_method": "ranked",
    "rebalance_threshold": 0.05,
    "baskets": {
      "growth_technology": {
        "method": "ranked",
        "max_positions": 5,
        "ranking": [
          { "signal": "relative_strength_60d", "weight": 0.40, "direction": "higher_is_better" },
          { "signal": "news_sentiment_7d", "weight": 0.30, "direction": "higher_is_better" },
          { "signal": "realized_volatility_20d", "weight": 0.30, "direction": "lower_is_better" }
        ]
      }
    }
  }
}
```

### Basket-Targeting Actions

Rules may target baskets using `target_type`.

Example:

```json
{
  "action": "decrease_weight",
  "target": "growth_technology",
  "target_type": "basket",
  "amount": 0.10,
  "unit": "weight"
}
```

Rules may also request rotation between baskets:

```json
{
  "action": "rotate",
  "from": "growth_technology",
  "to": "defensive_equities",
  "target": "portfolio",
  "target_type": "portfolio",
  "amount": 0.10,
  "unit": "weight"
}
```

The rule engine emits requested actions.
The portfolio engine determines the exact symbols to buy or sell.

---

## Source File Responsibilities

AlphaNet v0.1 uses a clear source-file boundary.

```text
manifest.json
  strategy metadata, compiler configuration, universe, portfolio configuration, and backtest defaults

strategy.md
  human-readable strategy intent and rationale

rules.json
  user-authored seed rules only

compiled/strategy.ir.json
  normalized deterministic AIR consumed by the backtester
```

Portfolio configuration belongs in `manifest.json` during authoring and in `compiled/strategy.ir.json` after compilation. This includes `starting_cash`, `initial_allocation`, `candidate_baskets`, `selection_policy`, `targets`, `constraints`, and `risk_budgets`.

The compiler carries the authoring portfolio model into AIR, validates it, and normalizes it for deterministic backtesting.

`rules.json` should remain focused on conditional decision rules. Rules may target symbols, baskets, asset classes, sectors, themes, cash, or the whole portfolio, but they do not define the portfolio model itself.

## Portfolio Initialization and Candidate Baskets

A portfolio should define both the starting state and the investable candidate set.

`initial_allocation` defines what the portfolio holds at the beginning of a backtest.

`candidate_baskets` define assets the strategy may buy or rotate into later, even when those assets are not part of the initial allocation.

`selection_policy` defines how the portfolio engine chooses among basket members when a rule targets a basket.

Example:

```json
{
  "portfolio": {
    "base_currency": "USD",
    "starting_cash": 100000,
    "initial_allocation": {
      "mode": "weights",
      "positions": [
        { "symbol": "SPY", "weight": 0.40 },
        { "symbol": "QQQ", "weight": 0.25 },
        { "symbol": "TLT", "weight": 0.20 },
        { "symbol": "cash", "weight": 0.15 }
      ]
    },
    "candidate_baskets": [
      {
        "basket_id": "growth_technology",
        "asset_class": "equities",
        "sector": "technology",
        "symbols": ["QQQ", "NVDA", "AMD", "MSFT", "AAPL", "AVGO", "SMH"]
      }
    ],
    "selection_policy": {
      "default_method": "ranked",
      "rebalance_threshold": 0.05
    }
  }
}
```
