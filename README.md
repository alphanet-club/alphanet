# AlphaNet

**A Market Intelligence Runtime for Strategies, Agents, and Portfolio Management**

AlphaNet is an open strategy platform for compiling human-authored strategy intent and optional agent feedback into a deterministic, portable trading strategy artifact.

At its core, AlphaNet separates **market reasoning** from **market evaluation**.

LLMs and agent frameworks may be used during strategy compilation, but backtesting and execution operate only on a deterministic Intermediate Representation called **AIR** (AlphaNet Intermediate Representation). This makes strategies portable, reproducible, explainable, and suitable for deterministic backtesting.

## Table of Contents

- [Vision](#vision)
- [Why AlphaNet Exists](#why-alphanet-exists)
- [Core Principles](#core-principles)
  - [1. Separate Reasoning From Evaluation](#1-separate-reasoning-from-evaluation)
  - [2. Backtesting Must Be Deterministic](#2-backtesting-must-be-deterministic)
  - [3. The IR Is the Contract](#3-the-ir-is-the-contract)
  - [4. Portfolio Constraints Always Win](#4-portfolio-constraints-always-win)
  - [5. Decisions Should Be Explainable](#5-decisions-should-be-explainable)
- [Architecture](#architecture)
  - [Authoring](#authoring)
  - [Compilation](#compilation)
  - [Evaluation](#evaluation)
- [AlphaNet IR](#alphanet-ir)
- [Signal Families](#signal-families)
  - [Valuation Signals](#valuation-signals)
  - [Macro Signals](#macro-signals)
  - [Global Macro Signals](#global-macro-signals)
  - [Fed and Monetary Plumbing Signals](#fed-and-monetary-plumbing-signals)
  - [Shipping and Real Economy Signals](#shipping-and-real-economy-signals)
  - [Market Flow Signals](#market-flow-signals)
  - [Event Signals](#event-signals)
  - [News and Social Sentiment Signals](#news-and-social-sentiment-signals)
  - [Volatility Signals](#volatility-signals)
- [Relations](#relations)
- [Regimes](#regimes)
- [Portfolio Engine](#portfolio-engine)
  - [Portfolio Targets](#portfolio-targets)
  - [Hard Constraints](#hard-constraints)
  - [Soft Constraints](#soft-constraints)
  - [Risk Budgets](#risk-budgets)
- [Decision Hierarchy](#decision-hierarchy)
  - [Layer Priorities](#layer-priorities)
  - [Rule Priorities](#rule-priorities)
  - [Conflict Resolution](#conflict-resolution)
- [Rules](#rules)
- [Runtime Modes](#runtime-modes)
  - [Local Compilation](#local-compilation)
  - [Hosted Compilation](#hosted-compilation)
  - [Scheduled Compilation](#scheduled-compilation)
  - [Single Engine Compilation](#single-engine-compilation)
  - [Ensemble Compilation](#ensemble-compilation)
- [Training Windows](#training-windows)
- [Versioning and Reproducibility](#versioning-and-reproducibility)
  - [Spec Version](#spec-version)
  - [Compiler Version](#compiler-version)
  - [Backtester Version](#backtester-version)
  - [Provenance](#provenance)
- [Repository Layout](#repository-layout)
  - [Root Files](#root-files)
  - [`specs/v0.1/`](#specsv01)
  - [`rules-compiler/`](#rules-compiler)
  - [`backtester/`](#backtester)
  - [`docs/`](#docs)
  - [`strategies/`](#strategies)
- [Strategy Folder Structure](#strategy-folder-structure)
  - [`manifest.json`](#manifestjson)
  - [`strategy.md`](#strategymd)
  - [`rules.json`](#rulesjson)
  - [`compiled/strategy.ir.json`](#compiledstrategyirjson)
  - [`compiled/provenance.json`](#compiledprovenancejson)
  - [`compiled/reasoning.md`](#compiledreasoningmd)
  - [`compiled/validation-report.json`](#compiledvalidation-reportjson)
  - [`backtests/`](#backtests)
- [Example Workflow](#example-workflow)
  - [1. Author a Strategy](#1-author-a-strategy)
  - [2. Compile the Strategy](#2-compile-the-strategy)
  - [3. Backtest the Strategy](#3-backtest-the-strategy)
  - [4. Inspect Results](#4-inspect-results)
- [End-to-End Example](#end-to-end-example)
  - [Market Conditions](#market-conditions)
  - [Relation](#relation)
  - [Rule Files](#rule-files)
  - [Portfolio Engine Applies Constraints](#portfolio-engine-applies-constraints)
  - [Backtester Records Decision Trace](#backtester-records-decision-trace)
- [Backtesting Philosophy](#backtesting-philosophy)
- [Roadmap](#roadmap)
  - [v0.1](#v01)
  - [v0.2](#v02)
  - [v0.3](#v03)
  - [Future](#future)
- [Summary](#summary)

## Vision

AlphaNet is designed to let users create, compile, backtest, share, and eventually compete with investment strategies.

Unlike traditional trading bots that directly execute indicator logic, AlphaNet models the market as a structured state machine consisting of:

- valuation signals
- macroeconomic signals
- global macro signals
- monetary policy and liquidity signals
- shipping and real economy signals
- market flow signals
- event risk signals
- news and social sentiment signals
- volatility signals
- portfolio constraints
- decision rules
- cross-asset relationships
- regime inference

The goal is not simply to generate trades.

The goal is to compile a strategy into a durable, explainable market model that can be evaluated across history.

---

## Why AlphaNet Exists

Many AI trading systems are expensive to run because the agent is called repeatedly during backtesting.

For example, a naive system might do this:

```text
Day 1 market data
        ↓
Trading agent / LLM
        ↓
Trade decision

Day 2 market data
        ↓
Trading agent / LLM
        ↓
Trade decision

Day 3 market data
        ↓
Trading agent / LLM
        ↓
Trade decision
```

That quickly becomes expensive, slow, and difficult to reproduce.

AlphaNet instead uses a compilation model:

```text
strategy.md
rules.json
manifest.json
        ↓
rules-compiler
        ↓
strategy.ir.json
        ↓
backtester
        ↓
deterministic results
```

The expensive reasoning step happens during compilation.

The backtester then evaluates the compiled strategy against historical data without calling an LLM or external agent.

---

## Core Principles

### 1. Separate Reasoning From Evaluation

The rules compiler may use:

- LLMs
- agent frameworks
- TradingAgents
- AI Hedge Fund
- news analysis
- social sentiment
- macro interpretation
- user-provided rules
- strategy descriptions

The backtester does not.

The backtester consumes only:

```text
strategy.ir.json
```

and historical market data.

---

### 2. Backtesting Must Be Deterministic

Given the same:

- `strategy.ir.json`
- backtester version
- data source version
- date range
- execution assumptions

The backtester should produce the same results.

This is critical for future leaderboard verification.

---

### 3. The IR Is the Contract

The compiled strategy artifact is the contract between:

- authoring
- compilation
- backtesting
- execution
- future leaderboard systems

The strategy may have been written by a human, assisted by an LLM, generated by an agent, or compiled locally.

Once it becomes `strategy.ir.json`, it is treated as a deterministic AlphaNet strategy.

---

### 4. Portfolio Constraints Always Win

Strategies may propose trades.

Rules may produce buy, sell, reduce, increase, or rebalance decisions.

But the portfolio engine is the final authority.

Portfolio safety rules, risk limits, allocation constraints, and concentration controls are evaluated before execution.

---

### 5. Decisions Should Be Explainable

Every trade should be traceable back to:

- triggered signals
- inferred regimes
- matched relations
- fired rules
- active portfolio constraints
- final execution decision

AlphaNet is designed so a decision can produce a trace like:

```json
{
  "decision": "reduce_position",
  "asset": "QQQ",
  "reasoning": [
    "tight_liquidity_regime",
    "wti_change_20d > 10%",
    "ust10y_change_20d > 25bps",
    "growth_tech_negative_relation",
    "equity_allocation_above_target"
  ],
  "rules_triggered": [
    "reduce_growth_tech_when_oil_and_rates_rise"
  ]
}
```

---

## Architecture

AlphaNet has three major phases:

```text
Authoring
    ↓
Compilation
    ↓
Evaluation
```

### Authoring

A strategy author provides:

```text
manifest.json
strategy.md
rules.json
```

These files describe:

- the strategy name and metadata
- the AlphaNet spec version
- the user’s intent
- optional user-defined rules
- compiler engines
- training window
- runtime preferences
- portfolio targets and constraints

---

### Compilation

The Go-based `rules-compiler` reads:

```text
manifest.json
strategy.md
rules.json
```

It may call one or more configured agent engines.

Examples:

- `TauricResearch/TradingAgents` version `0.2.5`
- `virattt/ai-hedge-fund` version `2026.6.17`

The compiler may analyze a training window and produce a finalized AlphaNet IR artifact:

```text
strategy.ir.json
```

It may also produce supporting artifacts:

```text
provenance.json
reasoning.md
validation-report.json
compile-report.json
```

---

### Evaluation

The Go-based `backtester` consumes only:

```text
strategy.ir.json
```

The backtester evaluates the compiled strategy against historical market data over a specified date range.

The backtester should not read `strategy.md`, call an LLM, call TradingAgents, or call AI Hedge Fund.

This boundary is intentional.

It keeps backtesting fast, deterministic, reproducible, and cheap.

---

## AlphaNet IR

The AlphaNet Intermediate Representation, or **AIR**, is the compiled strategy format.

A strategy IR may contain:

```json
{
  "metadata": {},
  "universe": {},
  "signals": {},
  "relations": {},
  "regimes": {},
  "portfolio": {},
  "decision_hierarchy": {},
  "rules": {},
  "execution": {}
}
```

The IR does not need to store every market observation.

Instead, it references signal definitions that the backtester can evaluate against historical data.

For example, the IR should generally express:

```json
{
  "signal": "wti_change_20d",
  "operator": ">",
  "value": 0.10
}
```

rather than embedding a single observed WTI value from one date.

This allows the same strategy to be backtested across many historical periods.

---

## Signal Families

AlphaNet models market state through signal families.

Signals are observations or derived measurements.

Signals do not directly execute trades.

---

### Valuation Signals

Examples:

- S&P 500 P/E
- Shiller CAPE
- earnings yield
- dividend yield
- equity risk premium

Purpose:

- long-term valuation context
- expected return pressure
- market fragility detection
- structural equity risk assessment

Example:

```json
{
  "signal_id": "sp500_cape",
  "family": "valuation",
  "type": "level",
  "source": "multpl",
  "value_type": "number"
}
```

---

### Macro Signals

Examples:

- Fed Funds Rate
- 2Y Treasury yield
- 10Y Treasury yield
- 10Y–2Y spread
- CPI YoY
- unemployment rate
- initial jobless claims
- nonfarm payrolls
- WTI crude
- dollar index
- M2 money supply
- consumer sentiment
- bank tightening

Purpose:

- inflation state
- growth state
- labor state
- liquidity state
- rate pressure
- commodity pressure

Example:

```json
{
  "signal_id": "ust10y_change_20d",
  "family": "macro",
  "type": "rate_change",
  "asset": "UST10Y",
  "window": "20d",
  "unit": "basis_points"
}
```

---

### Global Macro Signals

Examples:

- country CPI YoY
- short rates
- composite leading indicators
- house price indices
- equity indices
- relative country growth
- relative country inflation
- relative country rates

Purpose:

- capital flow inference
- global risk appetite
- relative macro strength
- regional liquidity conditions

---

### Fed and Monetary Plumbing Signals

Examples:

- Fed total assets
- Treasuries held
- MBS held
- dealer net positions
- dealer Treasury positions
- FOMC statements
- FOMC projections
- FOMC minutes

Purpose:

- liquidity transmission
- monetary stance
- policy regime
- market support
- dealer balance sheet pressure

---

### Shipping and Real Economy Signals

Examples:

- Suez Canal vessel count
- Suez Canal tonnage
- Panama Canal vessel count
- Strait of Hormuz flows
- Malacca Strait flows
- chokepoint disruptions
- shipping volume trends

Purpose:

- trade intensity
- supply chain stress
- commodity flow risk
- real economy nowcasting

---

### Market Flow Signals

Examples:

- gainers
- losers
- most active
- relative volume
- dollar volume
- sector rotation
- factor rotation
- growth vs value
- small cap vs large cap

Purpose:

- capital movement
- momentum
- liquidity clustering
- crowding
- market breadth
- thematic pressure

---

### Event Signals

Examples:

- earnings dates
- EPS estimates
- revenue estimates
- IPO calendar
- ex-dividend dates

Purpose:

- event risk
- volatility scheduling
- position sizing
- risk reduction windows

---

### News and Social Sentiment Signals

Examples:

- news sentiment by symbol
- social sentiment by symbol
- topic sentiment
- article count
- post volume
- narrative acceleration
- disagreement across sources

Possible sources:

- news APIs
- RSS feeds
- Reddit
- X/Twitter
- StockTwits
- earnings transcripts
- company filings

Purpose:

- narrative detection
- reflexivity modeling
- retail attention
- sentiment divergence
- momentum acceleration

Example:

```json
{
  "signal_id": "nvda_news_sentiment_7d",
  "family": "sentiment",
  "symbol": "NVDA",
  "source_type": "news",
  "window": "7d",
  "value_range": [-1, 1]
}
```

---

### Volatility Signals

Examples:

- VIX
- VIX term structure
- realized volatility
- implied volatility
- asset-level volatility
- event volatility
- earnings implied move
- volatility regime

Purpose:

- position sizing
- risk limits
- leverage limits
- event risk
- drawdown control

Example:

```json
{
  "signal_id": "nvda_realized_vol_20d",
  "family": "volatility",
  "symbol": "NVDA",
  "window": "20d",
  "type": "realized_volatility"
}
```

---

## Relations

Relations define cross-asset or cross-domain relationships.

A relation allows one market condition to influence the interpretation of another asset or asset class.

Example:

```text
Oil ↑
Rates ↑
    ↓
Growth technology pressure ↓
```

In IR form:

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
    "NVDA",
    "AMD"
  ],
  "effect": "negative",
  "confidence": 0.72
}
```

Other examples:

```text
Dollar ↑
    ↓
Commodities ↓
```

```text
Shipping volume ↓
    ↓
Global growth pressure ↓
```

```text
Fed liquidity ↑
    ↓
Risk assets ↑
```

```text
Credit tightening ↑
    ↓
Small caps ↓
```

Relations are not trades.

They are structured beliefs that rules can consume.

---

## Regimes

Regimes summarize the market environment.

Examples:

- Risk On
- Risk Off
- Tight Liquidity
- Loose Liquidity
- Inflationary
- Disinflationary
- Growth Expansion
- Growth Slowdown
- Stagflation
- Recovery
- High Volatility
- Low Volatility
- Credit Tightening
- Liquidity Expansion

A regime may be inferred from multiple signals and relations.

Example:

```json
{
  "regime_id": "tight_liquidity",
  "conditions": [
    {
      "signal": "fed_funds_rate",
      "operator": ">",
      "value": 3.5
    },
    {
      "signal": "dxy_change_60d",
      "operator": ">",
      "value": 0.03
    },
    {
      "signal": "bank_tightening_sloos",
      "operator": ">",
      "value": 5.0
    }
  ],
  "confidence": 0.81
}
```

Strategies primarily consume regimes rather than raw data whenever possible.

---

## Portfolio Engine

The portfolio engine is a first-class part of AlphaNet.

A strategy cannot be evaluated correctly without portfolio state.

The portfolio engine tracks:

- starting capital
- cash
- positions
- asset weights
- asset class weights
- sector weights
- realized and unrealized PnL
- risk exposure
- drawdown
- volatility
- allocation drift

Default starting capital for examples:

```json
{
  "starting_cash": 100000
}
```

---

### Portfolio Targets

A strategy may define allocation targets.

Example:

```json
{
  "targets": {
    "cash": 0.25,
    "equities": 0.25,
    "bonds": 0.25,
    "commodities": 0.25
  }
}
```

Targets may be hard or soft.

---

### Hard Constraints

Hard constraints must never be violated.

Examples:

```json
{
  "cash_min": 0.10,
  "max_single_position": 0.10,
  "max_sector_weight": {
    "technology": 0.30
  },
  "max_leverage": 1.0
}
```

---

### Soft Constraints

Soft constraints allow controlled drift.

Example:

```json
{
  "asset_class": "equities",
  "target": 0.50,
  "tolerance": 0.10
}
```

This means equities may drift between:

```text
40% and 60%
```

before the portfolio engine forces a correction.

---

### Risk Budgets

Risk budgets may include:

- maximum portfolio volatility
- maximum drawdown
- maximum beta
- maximum concentration
- maximum sector exposure
- maximum turnover
- maximum daily trade size

Example:

```json
{
  "max_portfolio_volatility": 0.20,
  "max_drawdown": 0.15,
  "max_daily_turnover": 0.10
}
```

---

## Decision Hierarchy

AlphaNet supports priorities at two levels:

1. **Layer priority**
2. **Individual rule priority inside each layer**

This allows every major subsystem to have a default priority while still allowing individual rules to override other rules in the same layer.

---

### Layer Priorities

Example default layer priorities:

| Layer | Priority |
|---|---:|
| Portfolio Safety | 100 |
| Risk Management | 90 |
| Regime Rules | 80 |
| Cross-Asset Relation Rules | 70 |
| Strategy Rules | 60 |
| Tactical Rules | 50 |
| Experimental Overlays | 25 |

Layer priorities determine which subsystem wins when conflicts occur across layers.

---

### Rule Priorities

Individual rules may also define a priority.

Example:

```json
{
  "rule_id": "reduce_growth_tech_when_oil_and_rates_rise",
  "layer": "strategy",
  "priority": 90,
  "confidence": 0.76
}
```

Another rule in the same layer may have lower priority:

```json
{
  "rule_id": "buy_semis_on_positive_ai_sentiment",
  "layer": "strategy",
  "priority": 45,
  "confidence": 0.70
}
```

If both rules fire and conflict, the higher-priority rule wins within that layer.

---

### Conflict Resolution

The engine resolves conflicts in this order:

```text
Layer Priority
    ↓
Rule Priority
    ↓
Confidence
    ↓
Tie Break Logic
```

For example:

```text
Strategy Rule:
BUY NVDA
priority 90

Risk Rule:
REDUCE NVDA
priority 40

Result:
REDUCE NVDA
```

Even though the strategy rule has a higher individual priority, the risk layer outranks the strategy layer.

Layer priority wins first.

---

## Rules

Rules consume:

- signals
- relations
- regimes
- portfolio state

Rules produce desired actions.

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

## Runtime Modes

AlphaNet supports multiple compilation and runtime approaches.

---

### Local Compilation

A user can run the rules compiler locally.

Example:

```bash
alphanet compile strategies/oil_rates_strategy
```

Local compilation allows users to use:

- their own API keys
- their own local LLM
- their own TradingAgents installation
- their own AI Hedge Fund installation
- their own training window

The result is a precompiled:

```text
strategy.ir.json
```

If a valid `strategy.ir.json` is provided, AlphaNet can use it directly.

---

### Hosted Compilation

A hosted AlphaNet system may compile strategies using platform compute.

This consumes compute credits based on:

- selected agent engines
- number of symbols
- training window size
- schedule frequency
- LLM usage
- external agent runtime cost

The output is still the same:

```text
strategy.ir.json
```

---

### Scheduled Compilation

Some strategies may recompile periodically.

Examples:

- daily
- weekly
- monthly
- quarterly
- event-driven

Event-driven triggers may include:

- FOMC decision
- CPI release
- earnings season
- volatility spike
- drawdown threshold
- regime change

Scheduled compilation should be defined in `manifest.json`.

---

### Single Engine Compilation

A strategy may use one agent engine.

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

### Ensemble Compilation

A strategy may use multiple engines.

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

Possible ensemble methods:

- union
- intersection
- weighted vote
- priority order
- LLM judge
- human approval

---

## Training Windows

A training window tells the compiler what historical period it may inspect when producing `strategy.ir.json`.

The backtester may later test the compiled strategy across a different date range.

Example:

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

This is important because agent reasoning should be bounded.

The compiler may analyze certain dates, market episodes, earnings events, macro environments, or historical regimes while constructing rules.

The backtester then evaluates the compiled rules over the requested backtest range without using the LLM.

---

## Versioning and Reproducibility

AlphaNet uses multiple versioned components.

---

### Spec Version

The spec version defines schema and semantic rules.

Example:

```json
{
  "spec_version": "v0.1"
}
```

---

### Compiler Version

The compiler version defines how strategy files are transformed into AIR.

Example:

```json
{
  "compiler_version": "v0.1.0"
}
```

---

### Backtester Version

The backtester version defines how AIR is evaluated.

Example:

```json
{
  "backtester_version": "v0.1.0"
}
```

---

### Provenance

Compiled strategies should include provenance metadata.

Example:

```json
{
  "strategy_name": "Oil Rates Strategy",
  "spec_version": "v0.1",
  "compiler_version": "v0.1.0",
  "generated_at": "2026-06-19",
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
    "start": "2020-01-01",
    "end": "2025-12-31"
  }
}
```

A future leaderboard can verify:

```text
strategy.ir.json
+
spec version
+
backtester version
+
data version
+
date range
=
reproducible result
```

---

## Repository Layout

Proposed AlphaNet repository layout:

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

### Root Files

#### `README.md`

High-level project overview.

Describes:

- vision
- architecture
- AlphaNet IR
- signals
- regimes
- portfolio engine
- decision hierarchy
- runtime modes
- repository layout
- roadmap

---

### `specs/v0.1/`

Versioned AlphaNet specification files.

#### `SPEC.md`

Formal specification for AlphaNet v0.1.

#### `CHANGELOG.md`

Spec-level changelog.

#### `alphanet.schema.json`

Top-level schema for `strategy.ir.json`.

#### `manifest.schema.json`

Schema for strategy `manifest.json`.

#### `portfolio.schema.json`

Schema for portfolio targets, constraints, and risk budgets.

#### `signal.schema.json`

Schema for signal definitions and signal references.

#### `rule.schema.json`

Schema for rule definitions, conditions, actions, priorities, and conflict resolution.

#### `regime.schema.json`

Schema for regime definitions.

---

### `rules-compiler/`

Go program that compiles source strategy files into AIR.

Inputs:

```text
manifest.json
strategy.md
rules.json
```

Outputs:

```text
compiled/strategy.ir.json
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
```

The compiler may call configured agent engines.

---

### `backtester/`

Go program that evaluates AIR against historical data.

Input:

```text
compiled/strategy.ir.json
```

Outputs:

- trades
- daily portfolio values
- returns
- drawdowns
- volatility
- Sharpe ratio
- Sortino ratio
- turnover
- final summary

The backtester should be deterministic.

---

### `docs/`

Supporting documentation.

Expected files:

```text
docs/architecture.md
docs/runtime.md
docs/portfolio-engine.md
docs/rule-engine.md
docs/examples/v0.1/
```

Examples are versioned by AlphaNet spec version.

---

### `strategies/`

Example or contributed strategy folders.

Each strategy may contain:

```text
manifest.json
strategy.md
rules.json
compiled/
backtests/
```

---

## Strategy Folder Structure

A single strategy should look like:

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

---

### `manifest.json`

Defines metadata and compilation behavior.

Example:

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

---

### `strategy.md`

Human-readable strategy description.

This file should explain:

- intent
- rationale
- target market conditions
- expected behavior
- risks
- assumptions
- assets or asset classes involved

The rules compiler may use this file as context.

---

### `rules.json`

User-provided initial rules.

These are not necessarily final.

The compiler reads this file, validates it, reconciles it with `strategy.md`, optionally incorporates agent feedback, and emits `strategy.ir.json`.

---

### `compiled/strategy.ir.json`

Final compiled AlphaNet IR.

This is the only strategy artifact required by the backtester.

---

### `compiled/provenance.json`

Records how the IR was produced.

---

### `compiled/reasoning.md`

Human-readable explanation of the compile process and major rule decisions.

---

### `compiled/validation-report.json`

Validation results against the AlphaNet schema.

---

### `backtests/`

Backtest outputs for specific date ranges and backtester versions.

---

## Example Workflow

### 1. Author a Strategy

Create:

```text
strategies/oil_rates_strategy/
├── manifest.json
├── strategy.md
└── rules.json
```

The human strategy intent might be:

```text
When oil and long rates rise together, reduce growth technology exposure,
increase cash, and optionally increase bond exposure if volatility is elevated.
```

---

### 2. Compile the Strategy

Run:

```bash
alphanet compile strategies/oil_rates_strategy
```

The compiler reads:

```text
manifest.json
strategy.md
rules.json
```

It may call configured agent engines.

It writes:

```text
compiled/strategy.ir.json
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
```

---

### 3. Backtest the Strategy

Run:

```bash
alphanet backtest strategies/oil_rates_strategy/compiled/strategy.ir.json \
  --start 2018-01-01 \
  --end 2025-12-31
```

The backtester evaluates the strategy deterministically.

---

### 4. Inspect Results

Backtest output may include:

```text
backtests/
├── 2018-01-01_2025-12-31.summary.json
├── 2018-01-01_2025-12-31.equity-curve.csv
├── 2018-01-01_2025-12-31.trades.csv
└── 2018-01-01_2025-12-31.decision-trace.jsonl
```

---

## End-to-End Example

### Market Conditions

Assume the backtester evaluates a historical day where:

```json
{
  "wti_change_20d": 0.12,
  "ust10y_change_20d_bps": 32,
  "vix": 25,
  "regime": "tight_liquidity"
}
```

---

### Relation

```json
{
  "relation_id": "oil_rates_negative_for_growth_tech",
  "effect": "negative",
  "targets": ["QQQ", "NVDA", "AMD"]
}
```

---

### Rule Files

```json
{
  "rule_id": "reduce_growth_tech_when_oil_and_rates_rise",
  "layer": "strategy",
  "priority": 70,
  "then": [
    {
      "action": "decrease_weight",
      "target": "QQQ",
      "amount": 0.05
    },
    {
      "action": "decrease_weight",
      "target": "NVDA",
      "amount": 0.03
    },
    {
      "action": "increase_weight",
      "target": "cash",
      "amount": 0.08
    }
  ]
}
```

---

### Portfolio Engine Applies Constraints

Current portfolio:

```json
{
  "cash": 0.18,
  "equities": 0.62,
  "bonds": 0.15,
  "commodities": 0.05
}
```

Target portfolio:

```json
{
  "cash": 0.25,
  "equities": 0.50,
  "bonds": 0.20,
  "commodities": 0.05
}
```

The rule is allowed because it:

- reduces overweight equities
- increases cash toward target
- does not violate max single-position limits
- does not violate minimum cash constraints

---

### Backtester Records Decision Trace

```json
{
  "date": "2026-06-19",
  "decision": "rebalance",
  "rules_triggered": [
    "reduce_growth_tech_when_oil_and_rates_rise"
  ],
  "signals": [
    "wti_change_20d",
    "ust10y_change_20d",
    "vix"
  ],
  "regimes": [
    "tight_liquidity"
  ],
  "actions": [
    {
      "action": "sell",
      "asset": "QQQ",
      "weight_delta": -0.05
    },
    {
      "action": "sell",
      "asset": "NVDA",
      "weight_delta": -0.03
    },
    {
      "action": "increase_cash",
      "weight_delta": 0.08
    }
  ]
}
```

---

## Backtesting Philosophy

Backtesting should be:

- deterministic
- reproducible
- LLM-free
- versioned
- explainable
- data-source aware

The backtester should support:

- daily replay
- portfolio state tracking
- rule evaluation
- conflict resolution
- transaction cost modeling
- slippage modeling
- rebalancing rules
- decision tracing
- performance metrics

Potential metrics:

- total return
- annualized return
- volatility
- Sharpe ratio
- Sortino ratio
- max drawdown
- turnover
- win rate
- beta
- exposure by asset class
- exposure by sector

---

## Roadmap

### v0.1

- Define AlphaNet IR
- Define repository layout
- Implement initial JSON schemas
- Implement Go rules compiler skeleton
- Implement Go backtester skeleton
- Support deterministic rule evaluation
- Support portfolio targets and constraints
- Support decision hierarchy
- Support strategy examples

---

### v0.2

Potential additions:

- richer regime inference
- sentiment adapters
- volatility surface support
- event calendar adapters
- FRED data ingestion
- Yahoo Finance data ingestion
- FMP data ingestion
- IMF PortWatch data ingestion
- OECD data ingestion

---

### v0.3

Potential additions:

- multi-engine compilation
- ensemble compiler modes
- compile reports
- validation reports
- provenance enforcement
- deterministic data snapshots
- strategy submission packaging

---

### Future

Potential future extensions:

- options positioning
- CFTC positioning
- gamma exposure
- insider transaction signals
- alternative data
- factor model support
- custom agent engines
- live paper trading
- separate leaderboard repository
- hosted compilation credits
- verified reproducibility framework

---

## Summary

AlphaNet is a market intelligence runtime built around a deterministic Intermediate Representation.

It allows expensive market reasoning to happen during compilation, while evaluation remains fast, cheap, and reproducible.

The central artifact is:

```text
strategy.ir.json
```

This file represents the compiled strategy.

It can be generated locally, generated through hosted compute, or produced by future agent engines.

Once compiled, it can be backtested deterministically against historical data.

AlphaNet’s long-term goal is to make strategies:

- portable
- inspectable
- explainable
- reproducible
- comparable
- agent-compatible
- portfolio-aware
