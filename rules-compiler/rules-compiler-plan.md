# Rules Compiler Implementation Plan

## Purpose

The `rules-compiler` is a Go program that converts AlphaNet strategy source files into a compiled AlphaNet Intermediate Representation, or AIR.

The compiler reads:

```text
manifest.json
strategy.md
rules.json
```

and writes:

```text
compiled/strategy.ir.json
compiled/provenance.json
compiled/reasoning.md
compiled/validation-report.json
```

The compiler is allowed to use LLMs, agent engines, local models, historical data, and other expensive reasoning tools.

The output must be deterministic enough to be validated, stored, hashed, and later consumed by the backtester.

---

## Source Contract

The compiler should treat the source files as having distinct responsibilities.

```text
manifest.json
  = strategy metadata, compiler configuration, universe, and portfolio configuration

strategy.md
  = human-readable strategy intent

rules.json
  = user-authored seed rules only

compiled/strategy.ir.json
  = normalized deterministic AIR consumed by the backtester
```

### `manifest.json`

`manifest.json` is the source of truth for portfolio configuration during authoring.

It should define:

- strategy identity
- spec version
- compiler mode
- compiler engines
- training window
- schedule
- universe
- starting capital
- initial allocation
- candidate baskets
- portfolio targets
- constraints
- risk budgets
- selection policy

### `strategy.md`

`strategy.md` explains the strategy in natural language.

It may describe desired behavior, market logic, risk assumptions, and examples of how the strategy should behave.

### `rules.json`

`rules.json` contains user-authored seed rules only.

It should define rules that the compiler may preserve, normalize, merge, refine, reject, or expand.

The compiler should not require `rules.json` to define portfolio configuration.

### `compiled/strategy.ir.json`

`compiled/strategy.ir.json` contains the normalized deterministic strategy.

The backtester should read portfolio configuration only from AIR.

---

## Initial CLI

Recommended command:

```bash
alphanet-compile ./strategies/oil-rates-growth-tech
```

Optional flags:

```bash
alphanet-compile ./strategies/oil-rates-growth-tech \
  --spec specs/v0.1 \
  --out ./strategies/oil-rates-growth-tech/compiled \
  --mode manual \
  --dry-run
```

Future flags:

```bash
--engine TradingAgents
--engine ai-hedge-fund
--training-start 2020-01-01
--training-end 2025-12-31
--lookback-days 365
--allow-network
--no-network
--hosted
--local
--validate-only
--emit-reasoning
```

---

## Inputs

### `manifest.json`

Defines:

- strategy name
- strategy id
- spec version
- compiler mode
- compiler engines
- training window
- schedule
- universe
- portfolio configuration
- hosted/local compute preferences

The portfolio section may include:

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
    "targets": {
      "cash": 0.20,
      "equities": 0.55,
      "bonds": 0.20,
      "commodities": 0.05
    }
  }
}
```

### `strategy.md`

Human-readable intent.

Used as natural language context by the compiler and optional engines.

### `rules.json`

User-provided seed rules.

Rules may be:

- preserved
- normalized
- merged
- refined
- rejected
- expanded

Rules may target symbols, baskets, asset classes, sectors, themes, cash, or the whole portfolio if supported by the schema.

---

## Outputs

### `compiled/strategy.ir.json`

The canonical compiled AIR artifact.

This is the only required input to the backtester.

The compiled AIR should include the final normalized portfolio object, including:

- `starting_cash`
- `initial_allocation`
- `candidate_baskets`
- `targets`
- `constraints`
- `selection_policy`
- `risk_budgets`
- `rebalance`

### `compiled/provenance.json`

Records how AIR was generated.

Should include:

- strategy name
- spec version
- compiler version
- generated timestamp
- source file hashes
- engines used
- engine versions
- training window
- output hash

### `compiled/reasoning.md`

Human-readable explanation of compilation.

Should summarize:

- interpreted strategy intent
- accepted rules
- rejected rules
- added signals
- added regimes
- added relations
- added or normalized candidate baskets
- added or normalized selection policies
- agent feedback summary
- major design decisions

### `compiled/validation-report.json`

Machine-readable validation result.

Should include:

- schema validation
- semantic validation
- warnings
- errors

### Optional

```text
compiled/compile-report.json
compiled/agent-output/
compiled/debug/
```

---

## Compiler Modes

### `none`

No compilation is performed.

A valid `compiled/strategy.ir.json` already exists.

Compiler may optionally validate it.

### `manual`

No external agent calls.

The compiler validates, normalizes, and compiles user-provided files.

### `single`

One configured engine is used.

### `ensemble`

Multiple engines are used and reconciled.

---

## Agent Engine Support

Initial engine interface should support adapters like:

- `TauricResearch/TradingAgents@0.2.5`
- `virattt/ai-hedge-fund@2026.6.17`

The first implementation can stub these engines.

The interface should allow real implementations later.

When agents are used, they may suggest:

- new signals
- updated regimes
- cross-asset relations
- preferred candidate baskets
- basket ranking signals
- updated selection policy weights
- rule refinements
- portfolio risk posture adjustments

The compiler is responsible for converting those suggestions into valid AIR.

---

## Proposed Go Package Structure

```text
rules-compiler/
├── go.mod
├── cmd/
│   └── alphanet-compile/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── compile.go
│   ├── manifest/
│   │   ├── load.go
│   │   └── validate.go
│   ├── strategy/
│   │   ├── load.go
│   │   └── source.go
│   ├── rules/
│   │   ├── load.go
│   │   └── normalize.go
│   ├── portfolio/
│   │   ├── normalize.go
│   │   ├── baskets.go
│   │   └── selection_policy.go
│   ├── compiler/
│   │   ├── compiler.go
│   │   ├── manual.go
│   │   ├── single.go
│   │   └── ensemble.go
│   ├── engines/
│   │   ├── engine.go
│   │   ├── tradingagents.go
│   │   ├── aihedgefund.go
│   │   └── stub.go
│   ├── air/
│   │   ├── model.go
│   │   ├── build.go
│   │   ├── hash.go
│   │   └── write.go
│   ├── validation/
│   │   ├── schema.go
│   │   ├── semantic.go
│   │   └── report.go
│   ├── provenance/
│   │   └── provenance.go
│   └── reasoning/
│       └── reasoning.go
└── testdata/
    ├── valid/
    └── invalid/
```

---

## Suggested Core Types

### Manifest

```go
type Manifest struct {
    Name        string
    StrategyID  string
    Description string
    Author      string
    SpecVersion string
    Version     string
    Tags        []string
    Compiler    CompilerConfig
    Universe    UniverseConfig
    Portfolio   PortfolioConfig
}
```

### PortfolioConfig

```go
type PortfolioConfig struct {
    BaseCurrency      string
    StartingCash      float64
    InitialAllocation InitialAllocation
    CandidateBaskets  []CandidateBasket
    Targets           map[string]float64
    Constraints       PortfolioConstraints
    SelectionPolicy   SelectionPolicy
    RiskBudgets       RiskBudgets
    Rebalance         RebalancePolicy
}
```

### InitialAllocation

```go
type InitialAllocation struct {
    Mode      string
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
    BasketID          string
    Description       string
    AssetClass        string
    Sector            string
    Sectors           []string
    Theme             string
    Symbols           []string
    Role              string
    MinWeight         *float64
    MaxWeight         *float64
    MinPositionWeight *float64
    MaxPositionWeight *float64
}
```

### SelectionPolicy

```go
type SelectionPolicy struct {
    DefaultMethod      string
    RebalanceThreshold float64
    Baskets            map[string]BasketSelectionPolicy
}
```

### CompilerConfig

```go
type CompilerConfig struct {
    Mode               string
    Engines            []EngineConfig
    EnsembleMethod     string
    TrainingWindow     TrainingWindow
    AllowNetwork       bool
    AllowHostedCompute bool
}
```

### Engine

```go
type Engine interface {
    Name() string
    Version() string
    Analyze(ctx context.Context, input EngineInput) (EngineOutput, error)
}
```

### EngineInput

```go
type EngineInput struct {
    Manifest       Manifest
    StrategyMD     string
    RulesJSON      []byte
    TrainingWindow TrainingWindow
    Universe       []string
    Portfolio      PortfolioConfig
}
```

### EngineOutput

```go
type EngineOutput struct {
    Signals         []SignalSuggestion
    Relations       []RelationSuggestion
    Regimes         []RegimeSuggestion
    Rules           []RuleSuggestion
    CandidateBaskets []CandidateBasketSuggestion
    SelectionPolicy *SelectionPolicySuggestion
    Portfolio       *PortfolioSuggestion
    Notes           string
}
```

### AIR

AIR should mirror `alphanet.schema.json`.

A practical first implementation can represent AIR with explicit Go structs or generic maps.

Recommendation:

Start with structs for core fields and allow `map[string]any` for extension metadata.

---

## Compilation Pipeline

Recommended pipeline:

```text
Load source files
    ↓
Validate manifest
    ↓
Validate seed rules
    ↓
Normalize portfolio configuration from manifest
    ↓
Resolve compiler mode
    ↓
Load engines
    ↓
Collect engine feedback
    ↓
Normalize suggestions
    ↓
Merge user rules and agent suggestions
    ↓
Merge candidate basket and selection policy suggestions
    ↓
Build AIR
    ↓
Semantic validation
    ↓
Schema validation
    ↓
Write outputs
```

---

## Portfolio Normalization

The compiler should normalize the authored portfolio configuration before writing AIR.

Tasks:

1. Validate `starting_cash`.
2. Validate `initial_allocation`.
3. Default missing initial allocation to 100% cash.
4. Validate candidate basket ids are unique.
5. Validate candidate basket symbols exist in the universe or add them to the normalized AIR universe.
6. Validate selection policies reference known baskets.
7. Validate targets sum or intentionally allow residual cash.
8. Validate constraints are internally consistent.
9. Normalize cash as a first-class allocation target.
10. Emit warnings for unused candidate baskets.

---

## Semantic Validation

Schema validation is not enough.

The compiler should also verify:

- unique signal ids
- unique rule ids
- unique regime ids
- unique relation ids
- unique candidate basket ids
- every signal reference resolves
- every regime reference resolves
- every relation reference resolves
- every selection policy basket reference resolves
- every rule layer exists in decision hierarchy
- every rule action target is known or allowed
- every basket-targeting action references a known basket
- portfolio constraints are internally valid
- initial allocation is valid
- initial allocation symbols resolve
- compiler engines are version-pinned
- training window is valid
- no unsupported operators are used

---

## Rule Merge Strategy

For v0.1, keep merging simple.

Suggested behavior:

1. Preserve user rules when valid.
2. Add compiler-generated signals required by rules.
3. Add regimes only when referenced or useful.
4. Add relations when cross-asset logic is present.
5. Add portfolio safety rules if missing.
6. Preserve explicit manifest portfolio constraints unless an agent suggestion is accepted by the compiler.
7. Emit warnings when agent suggestions conflict with user rules or portfolio configuration.

---

## Candidate Basket Strategy

For v0.1, candidate baskets should be deterministic data structures.

The compiler may:

- preserve authored baskets
- add metadata to baskets
- add symbols suggested by engines
- remove duplicate symbols
- normalize sectors and asset classes
- attach ranking signals through `selection_policy`

The compiler should avoid silently replacing the user's candidate universe.

If an agent suggests removing symbols from a basket, the compiler should record that decision in `reasoning.md`.

---

## Ensemble Strategy

For v0.1, ensemble support can be mostly structural.

Supported methods:

### `union`

Include all non-conflicting suggestions.

### `intersection`

Include only suggestions supported by multiple engines.

### `weighted_vote`

Assign scores based on engine weights.

### `priority_order`

Earlier engines win conflicts.

### `human_review`

Emit unresolved suggestions for manual review.

### `llm_judge`

Reserved for future.

---

## Hashing and Provenance

Compiler should compute hashes for:

- `manifest.json`
- `strategy.md`
- `rules.json`
- `strategy.ir.json`

Recommended:

- SHA-256
- canonical JSON for AIR
- record hashes in `provenance.json`

Example:

```json
{
  "source_hashes": {
    "manifest.json": "sha256:...",
    "strategy.md": "sha256:...",
    "rules.json": "sha256:..."
  },
  "ir_sha256": "sha256:..."
}
```

---

## Validation Report

Example:

```json
{
  "status": "valid",
  "schemas": [
    "manifest.schema.json",
    "alphanet.schema.json"
  ],
  "checks": [
    {
      "name": "manifest_schema",
      "status": "pass"
    },
    {
      "name": "signal_references_resolved",
      "status": "pass"
    },
    {
      "name": "candidate_baskets_valid",
      "status": "pass"
    }
  ],
  "warnings": [],
  "errors": []
}
```

---

## First Implementation Milestone

Build the compiler with `manual` mode first.

Milestone 1 should:

1. Load `manifest.json`.
2. Load `strategy.md`.
3. Load `rules.json`.
4. Validate required fields.
5. Normalize portfolio configuration from manifest.
6. Create basic AIR.
7. Write `compiled/strategy.ir.json`.
8. Write `compiled/provenance.json`.
9. Write `compiled/reasoning.md`.
10. Write `compiled/validation-report.json`.

No real LLM required.

No real agent integration required.

---

## Second Implementation Milestone

Add schema validation.

Tasks:

1. Load JSON schemas from `specs/v0.1`.
2. Validate manifest.
3. Validate AIR.
4. Validate rules.
5. Validate portfolio schema.
6. Emit validation report.

---

## Third Implementation Milestone

Add stub engine support.

Tasks:

1. Define `Engine` interface.
2. Add `stub` engine.
3. Add `single` mode.
4. Add `ensemble` mode with union merge.
5. Record engine output in reasoning file.
6. Support stub candidate basket and selection policy suggestions.

---

## Fourth Implementation Milestone

Add real agent adapters.

Potential adapters:

- TradingAgents adapter
- AI Hedge Fund adapter
- generic CLI adapter
- generic HTTP adapter
- local LLM adapter

A generic CLI adapter may be the best first step:

```json
{
  "name": "custom-cli-engine",
  "version": "0.1.0",
  "config": {
    "command": "python run_agent.py --input {{input}} --output {{output}}"
  }
}
```

---

## Testing Plan

Test cases:

### Valid

- minimal manifest
- manifest with initial allocation
- manifest with candidate baskets
- manifest with selection policy
- manual strategy
- ensemble strategy
- precompiled AIR mode
- valid seed rules
- basket-targeting rules

### Invalid

- missing spec version
- unsupported compiler mode
- duplicate rule id
- duplicate basket id
- unresolved signal reference
- unresolved basket reference
- invalid initial allocation
- invalid training window
- invalid engine config
- invalid rule priority
- invalid portfolio constraint

---

## Non-Goals for v0.1

Do not implement yet:

- full hosted compute billing
- live trading
- production agent orchestration
- distributed compilation
- prompt optimization
- automatic strategy repair
- leaderboard submission protocol

---

## Success Criteria

The compiler is successful when:

1. A user can provide a strategy folder.
2. The compiler emits valid AIR.
3. The AIR validates against v0.1 schemas.
4. The AIR contains normalized portfolio initialization and candidate basket data.
5. The backtester can consume the AIR without reading source files.
6. Provenance and reasoning outputs are produced.
