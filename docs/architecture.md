# AlphaNet Architecture

## Overview

AlphaNet is a market intelligence runtime for compiling human-authored strategy intent and optional agent feedback into a deterministic strategy artifact.

The core architectural idea is:

```text
Reasoning happens during compilation.
Evaluation happens during deterministic backtesting.
```

AlphaNet separates:

1. Strategy authoring
2. Agent-assisted compilation
3. Deterministic strategy evaluation
4. Portfolio-aware execution simulation

The primary compiled artifact is:

```text
compiled/strategy.ir.json
```

This file contains the AlphaNet Intermediate Representation, or **AIR**.

---

## High-Level Flow

```text
manifest.json
strategy.md
rules.json
        │
        ▼
┌─────────────────┐
│ Rules Compiler  │
│                 │
│ LLMs            │
│ TradingAgents   │
│ AI Hedge Fund   │
│ User Rules      │
└─────────────────┘
        │
        ▼
compiled/strategy.ir.json
        │
        ▼
┌─────────────────┐
│   Backtester    │
│                 │
│ No LLMs         │
│ No Agents       │
│ Deterministic   │
└─────────────────┘
        │
        ▼
Backtest Results
Decision Trace
Performance Metrics
```

---

## Major Components

### 1. Strategy Source

Strategy source files are human-authored or user-provided inputs.

```text
manifest.json
strategy.md
rules.json
```

They describe what the strategy is trying to do, what engines may be used during compilation, what user-defined rules should be considered, and what portfolio assumptions should be respected.

The source files are not consumed directly by the backtester.

---

### 2. Rules Compiler

The `rules-compiler` is a Go program that converts source strategy files into AIR.

It may use:

- user-provided `rules.json`
- natural-language strategy intent from `strategy.md`
- compiler settings from `manifest.json`
- LLMs
- local models
- agent engines
- historical training windows
- external data adapters

Example agent engines:

- `TauricResearch/TradingAgents` version `0.2.5`
- `virattt/ai-hedge-fund` version `2026.6.17`

The compiler emits:

```text
compiled/strategy.ir.json
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
```

---

### 3. AlphaNet Intermediate Representation

AIR is the deterministic strategy representation.

A compiled AIR file may include:

- metadata
- universe
- signal definitions
- relation definitions
- regime definitions
- portfolio targets
- constraints
- decision hierarchy
- rules
- execution assumptions
- provenance

AIR is designed to be portable and reproducible.

The backtester does not care whether AIR came from:

- a human
- a local compiler
- hosted compute
- TradingAgents
- AI Hedge Fund
- another future agent engine

If the AIR is valid, it can be evaluated.

---

### 4. Backtester

The `backtester` is a Go program that evaluates AIR against historical market data.

The backtester must not call:

- LLMs
- TradingAgents
- AI Hedge Fund
- arbitrary external agent engines

The backtester consumes only:

```text
compiled/strategy.ir.json
```

and a versioned market data source.

The output should be deterministic given the same:

- AIR
- backtester version
- data version
- date range
- execution assumptions

---

## Architectural Boundary

The most important boundary in AlphaNet is this:

```text
Compiler = may reason
Backtester = must evaluate deterministically
```

This allows AlphaNet to use powerful but expensive reasoning systems without making backtesting slow, non-reproducible, or cost-prohibitive.

---

## Authoring Layer

The authoring layer contains:

```text
manifest.json
strategy.md
rules.json
```

### `manifest.json`

Defines strategy identity and compiler behavior.

It may include:

- name
- description
- author
- spec version
- compiler mode
- engine list
- training window
- schedule
- universe
- default backtest settings

### `strategy.md`

Human-readable strategy intent.

This file should explain:

- why the strategy exists
- what it is trying to capture
- what market conditions it reacts to
- what assets or asset classes it considers
- risk assumptions
- expected behavior

### `rules.json`

User-provided seed rules.

These rules are not necessarily final. The compiler may refine, reject, merge, or expand them.

---

## Compilation Layer

The compilation layer converts source strategy files into AIR.

Compilation may be:

- local
- hosted
- manual
- single-engine
- ensemble
- scheduled
- event-driven

The compiler is allowed to be expensive because it does not run for every backtest day.

---

## Evaluation Layer

The evaluation layer is deterministic.

It performs:

1. AIR validation
2. market data loading
3. signal evaluation
4. regime evaluation
5. relation evaluation
6. rule evaluation
7. conflict resolution
8. portfolio constraint enforcement
9. execution simulation
10. performance calculation
11. decision trace recording

---

## Data Flow

```text
Historical Market Data
        │
        ▼
Signal Engine
        │
        ▼
Regime Engine
        │
        ▼
Relation Engine
        │
        ▼
Rule Engine
        │
        ▼
Portfolio Engine
        │
        ▼
Execution Simulator
        │
        ▼
Backtest Output
```

---

## Versioning

AlphaNet has several versioned surfaces.

### Spec Version

Defines AIR structure and semantics.

Example:

```json
{
  "spec_version": "v0.1"
}
```

### Compiler Version

Defines how source files are compiled into AIR.

Example:

```json
{
  "compiler_version": "v0.1.0"
}
```

### Backtester Version

Defines how AIR is evaluated.

Example:

```json
{
  "backtester_version": "v0.1.0"
}
```

### Data Version

Defines the market data snapshot used for evaluation.

A reproducible result requires all of these to be known.

---

## Reproducibility Model

A future verified result should be reproducible from:

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
+
execution assumptions
```

The compiler does not need to be rerun to reproduce a backtest result.

This is a key design feature.

---

## Decision Traceability

Every trade or rebalance should be explainable.

A decision trace should record:

- date
- evaluated signals
- active regimes
- active relations
- triggered rules
- conflicts
- portfolio constraints
- requested actions
- approved actions
- rejected actions
- final portfolio state

This makes AlphaNet strategies inspectable and debuggable.

---

## Repository Placement

This document belongs at:

```text
docs/architecture.md
```

The formal specification belongs at:

```text
specs/v0.1/SPEC.md
```

---

## Sampling Architecture

Sampling is part of both compilation and backtesting.

Compiler sampling controls which timestamps inside a training window are analyzed by expensive reasoning engines.

Backtest decision sampling controls which timestamps are eligible for strategy decisions and trades.

Portfolio valuation frequency is separate. This allows a strategy to make decisions weekly or monthly while still measuring performance and drawdown daily.
