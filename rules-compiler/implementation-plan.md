# Rules Compiler — Implementation Plan

## 1. Overview

The `rules-compiler` is a Go program that converts AlphaNet strategy source files (`manifest.json`, `strategy.md`, `rules.json`) into a compiled AlphaNet Intermediate Representation (AIR) artifact (`strategy.ir.json`) plus supporting outputs (`provenance.json`, `reasoning.md`, `validation-report.json`).

The compiler is permitted to use LLMs, agent engines (TradingAgents, ai-hedge-fund), and historical data during compilation, but its output must be deterministic enough to be validated, stored, hashed, and consumed by the backtester.

---

## 2. Go Package Structure

```
rules-compiler/
├── go.mod                          # module github.com/alphanet/rules-compiler
├── cmd/
│   └── alphanet-compile/
│       └── main.go                 # CLI entry point; flag parsing, orchestration
├── internal/
│   ├── air/
│   │   ├── model.go                # All AIR Go struct definitions (mirrors alphanet.schema.json)
│   │   ├── build.go                # BuildAIR() — assembles AIR from normalized source + engine output
│   │   ├── hash.go                 # Canonical JSON serialization + SHA-256 hashing
│   │   └── write.go                # WriteAIR() — writes all 4 output files to disk
│   ├── app/
│   │   └── compile.go              # Orchestrator: loads sources, resolves mode, runs pipeline, writes outputs
│   ├── compiler/
│   │   ├── compiler.go             # Compile() — top-level pipeline function
│   │   ├── manual.go               # ManualCompile() — no engines, just validate + normalize
│   │   ├── single.go               # SingleEngineCompile() — one engine
│   │   └── ensemble.go             # EnsembleCompile() — multiple engines, merge strategies
│   ├── engines/
│   │   ├── engine.go               # Engine interface definition
│   │   ├── stub.go                 # StubEngine — returns deterministic suggestions for testing
│   │   ├── tradingagents.go        # TradingAgents adapter (stub in v0.1, real later)
│   │   └── aihedgefund.go          # AI Hedge Fund adapter (stub in v0.1, real later)
│   ├── manifest/
│   │   ├── load.go                 # LoadManifest() — reads + unmarshals manifest.json
│   │   └── validate.go             # ValidateManifest() — required fields, types, constraints
│   ├── portfolio/
│   │   ├── normalize.go            # NormalizePortfolio() — validate and normalize portfolio config
│   │   ├── baskets.go              # NormalizeBaskets() — validate candidate baskets
│   │   └── selection_policy.go     # NormalizeSelectionPolicy() — validate selection policy
│   ├── provenance/
│   │   └── provenance.go           # BuildProvenance() — source hashes, timestamps, compiler version
│   ├── reasoning/
│   │   └── reasoning.go            # BuildReasoning() — assemble human-readable compilation narrative
│   ├── rules/
│   │   ├── load.go                 # LoadRules() — reads + unmarshals rules.json
│   │   └── normalize.go            # NormalizeRules() — validate, deduplicate, resolve references
│   ├── strategy/
│   │   ├── load.go                 # LoadStrategyMD() — reads strategy.md
│   │   └── source.go               # SourceContext — holds all raw source data
│   └── validation/
│       ├── schema.go               # SchemaValidator — loads JSON schemas, validates against them
│       ├── semantic.go             # SemanticValidator — cross-reference checks, uniqueness, resolution
│       └── report.go               # ValidationReport — structured validation result
└── testdata/
    ├── valid/
    │   ├── minimal/                # Minimal valid strategy (just name + spec_version)
    │   ├── manual/                 # manual mode strategy with candidate baskets
    │   ├── ensemble/               # ensemble mode strategy with multiple engines
    │   └── precompiled/            # none mode strategy with existing strategy.ir.json
    └── invalid/
        ├── missing-spec-version/
        ├── unsupported-mode/
        ├── duplicate-rule-id/
        ├── duplicate-basket-id/
        ├── unresolved-signal-ref/
        ├── unresolved-basket-ref/
        ├── invalid-initial-allocation/
        ├── invalid-training-window/
        ├── invalid-engine-config/
        ├── invalid-rule-priority/
        └── invalid-portfolio-constraint/
```

---

## 3. Core Go Types

### 3.1 Manifest (`internal/manifest/`)

```go
type Manifest struct {
    Name          string          `json:"name" validate:"required"`
    StrategyID    string          `json:"strategy_id,omitempty"`
    Description   string          `json:"description,omitempty"`
    Author        string          `json:"author,omitempty"`
    SpecVersion   string          `json:"spec_version" validate:"required"`
    Version       string          `json:"version,omitempty"`
    Tags          []string        `json:"tags,omitempty"`
    Compiler      CompilerConfig  `json:"compiler,omitempty"`
    Universe      UniverseConfig  `json:"universe,omitempty"`
    Portfolio     PortfolioConfig `json:"portfolio,omitempty"`
    Backtest      BacktestConfig  `json:"backtest,omitempty"`
}

type CompilerConfig struct {
    Mode               string          `json:"mode" validate:"oneof=none manual single ensemble"`
    Engines            []EngineConfig  `json:"engines,omitempty"`
    EnsembleMethod     string          `json:"ensemble_method,omitempty"`
    TrainingWindow     TrainingWindow  `json:"training_window,omitempty"`
    AllowNetwork       bool            `json:"allow_network,omitempty"`
    AllowHostedCompute bool            `json:"allow_hosted_compute,omitempty"`
    Schedule           *ScheduleConfig `json:"schedule,omitempty"`
    Notes              string          `json:"notes,omitempty"`
}

type EngineConfig struct {
    Name    string `json:"name" validate:"required"`
    Version string `json:"version" validate:"required"`
    Weight  float64 `json:"weight,omitempty" validate:"min=0,max=1"`
    Role    string `json:"role,omitempty"`
    Symbols []string `json:"symbols,omitempty"`
    Enabled *bool   `json:"enabled,omitempty"`
    Config  map[string]any `json:"config,omitempty"`
    Notes   string `json:"notes,omitempty"`
}

type TrainingWindow struct {
    LookbackDays  int                 `json:"lookback_days,omitempty"`
    Start         string              `json:"start,omitempty"` // "2020-01-01"
    End           string              `json:"end,omitempty"`
    Sampling      *SamplingConfig     `json:"sampling,omitempty"`
    IncludeRanges []DateRange         `json:"include_ranges,omitempty"`
    ExcludeRanges []DateRange         `json:"exclude_ranges,omitempty"`
}

type UniverseConfig struct {
    Symbols      []string `json:"symbols,omitempty"`
    AssetClasses []string `json:"asset_classes,omitempty"`
}
```

### 3.2 Portfolio (`internal/portfolio/`)

```go
type PortfolioConfig struct {
    BaseCurrency      string              `json:"base_currency"`
    StartingCash      float64             `json:"starting_cash"`
    InitialAllocation InitialAllocation   `json:"initial_allocation,omitempty"`
    CandidateBaskets  []CandidateBasket   `json:"candidate_baskets,omitempty"`
    Targets           map[string]float64  `json:"targets,omitempty"`
    SoftTargets       []SoftTarget        `json:"soft_targets,omitempty"`
    Constraints       *PortfolioConstraints `json:"constraints,omitempty"`
    SelectionPolicy   *SelectionPolicy    `json:"selection_policy,omitempty"`
    RiskBudgets       *RiskBudgets        `json:"risk_budgets,omitempty"`
    Rebalance         *RebalancePolicy    `json:"rebalance,omitempty"`
}

type InitialAllocation struct {
    Mode      string             `json:"mode" validate:"oneof=cash weights dollars shares"`
    Positions []InitialPosition  `json:"positions,omitempty"`
}

type InitialPosition struct {
    Symbol string  `json:"symbol"`
    Weight float64 `json:"weight,omitempty"`
    Shares float64 `json:"shares,omitempty"`
    Amount float64 `json:"amount,omitempty"`
}

type CandidateBasket struct {
    BasketID          string    `json:"basket_id" validate:"required"`
    Description       string    `json:"description,omitempty"`
    AssetClass        string    `json:"asset_class,omitempty"`
    Sector            string    `json:"sector,omitempty"`
    Sectors           []string  `json:"sectors,omitempty"`
    Theme             string    `json:"theme,omitempty"`
    Symbols           []string  `json:"symbols" validate:"required,min=1"`
    Role              string    `json:"role,omitempty"`
    MinWeight         *float64  `json:"min_weight,omitempty"`
    MaxWeight         *float64  `json:"max_weight,omitempty"`
    MinPositionWeight  *float64  `json:"min_position_weight,omitempty"`
    MaxPositionWeight  *float64  `json:"max_position_weight,omitempty"`
}

type SelectionPolicy struct {
    DefaultMethod      string                           `json:"default_method"`
    RebalanceThreshold float64                          `json:"rebalance_threshold,omitempty"`
    MinHoldingDays     int                              `json:"min_holding_days,omitempty"`
    Baskets            map[string]BasketSelectionPolicy `json:"baskets,omitempty"`
}

type BasketSelectionPolicy struct {
    Method            string            `json:"method"`
    MaxPositions      int               `json:"max_positions,omitempty"`
    MinPositionWeight float64           `json:"min_position_weight,omitempty"`
    MaxPositionWeight float64           `json:"max_position_weight,omitempty"`
    Ranking           []RankingCriterion `json:"ranking,omitempty"`
    ReplacementPolicy *ReplacementPolicy `json:"replacement_policy,omitempty"`
}

type RankingCriterion struct {
    Signal    string  `json:"signal" validate:"required"`
    Weight    float64 `json:"weight"`
    Direction string  `json:"direction" validate:"oneof=higher_is_better lower_is_better"`
}

type PortfolioConstraints struct {
    CashMin           *float64           `json:"cash_min,omitempty"`
    MaxSinglePosition *float64           `json:"max_single_position,omitempty"`
    MaxAssetClassWeight map[string]float64 `json:"max_asset_class_weight,omitempty"`
    MaxSectorWeight     map[string]float64 `json:"max_sector_weight,omitempty"`
    MaxLeverage       *float64           `json:"max_leverage,omitempty"`
    AllowShorting     bool               `json:"allow_shorting,omitempty"`
    AllowMargin       bool               `json:"allow_margin,omitempty"`
}

type RiskBudgets struct {
    MaxPortfolioVolatility *float64 `json:"max_portfolio_volatility,omitempty"`
    MaxDrawdown            *float64 `json:"max_drawdown,omitempty"`
    MaxDailyTurnover       *float64 `json:"max_daily_turnover,omitempty"`
    MaxTradeWeight         *float64 `json:"max_trade_weight,omitempty"`
}

type RebalancePolicy struct {
    Frequency            string `json:"frequency"`
    Threshold            float64 `json:"threshold,omitempty"`
    AllowPartialRebalance bool   `json:"allow_partial_rebalance,omitempty"`
}

type SoftTarget struct {
    Target    string  `json:"target"`
    Weight    float64 `json:"weight"`
    Tolerance float64 `json:"tolerance"`
    Priority  int     `json:"priority,omitempty"`
}
```

### 3.3 AIR Model (`internal/air/model.go`)

The AIR model mirrors `specs/v0.1/alphanet.schema.json`:

```go
type AIR struct {
    Metadata          AIRMetadata       `json:"metadata"`
    Universe          AIRUniverse       `json:"universe"`
    Signals           []Signal          `json:"signals"`
    Relations         []Relation        `json:"relations,omitempty"`
    Regimes           []Regime          `json:"regimes,omitempty"`
    Portfolio         AIRPortfolio      `json:"portfolio"`
    DecisionHierarchy DecisionHierarchy `json:"decision_hierarchy"`
    Rules             []Rule            `json:"rules"`
    Execution         ExecutionConfig   `json:"execution"`
    Provenance        Provenance        `json:"provenance,omitempty"`
    Benchmarks        []Benchmark       `json:"benchmarks,omitempty"`
    Extensions        map[string]any    `json:"extensions,omitempty"`
}

type AIRMetadata struct {
    StrategyName    string   `json:"strategy_name"`
    StrategyID      string   `json:"strategy_id,omitempty"`
    Description     string   `json:"description,omitempty"`
    Author          string   `json:"author,omitempty"`
    Version         string   `json:"version,omitempty"`
    SpecVersion     string   `json:"spec_version"`
    CompilerVersion string   `json:"compiler_version"`
    GeneratedAt     string   `json:"generated_at"` // ISO 8601
    Tags            []string `json:"tags,omitempty"`
}

type AIRUniverse struct {
    Assets       []Asset       `json:"assets"`
    AssetClasses []string      `json:"asset_classes,omitempty"`
    Sectors      []string      `json:"sectors,omitempty"`
    Themes       []Theme       `json:"themes,omitempty"`
    Benchmark    string        `json:"benchmark,omitempty"`
}

type Asset struct {
    Symbol     string `json:"symbol"`
    Name       string `json:"name,omitempty"`
    AssetClass string `json:"asset_class"`
    Sector     string `json:"sector,omitempty"`
    Currency   string `json:"currency"`
    Tradable   bool   `json:"tradable"`
}

type Theme struct {
    ThemeID     string   `json:"theme_id"`
    Name        string   `json:"name"`
    Members     []string `json:"members"`
    Description string   `json:"description,omitempty"`
}

type Signal struct {
    SignalID    string      `json:"signal_id"`
    Family      string      `json:"family"`
    Type        string      `json:"type,omitempty"`
    Name        string      `json:"name,omitempty"`
    Description string      `json:"description,omitempty"`
    Source      SignalSource `json:"source,omitempty"`
    Instrument  string      `json:"instrument,omitempty"`
    Symbol      string      `json:"symbol,omitempty"`
    Field       string      `json:"field,omitempty"`
    Transform   string      `json:"transform,omitempty"`
    Window      string      `json:"window,omitempty"`
    Frequency   string      `json:"frequency,omitempty"`
    Unit        string      `json:"unit,omitempty"`
    ValueRange  []float64   `json:"value_range,omitempty"`
}

type SignalSource struct {
    Name     string `json:"name"`
    SeriesID string `json:"series_id,omitempty"`
    Adapter  string `json:"adapter,omitempty"`
}

type Relation struct {
    RelationID  string          `json:"relation_id"`
    Description string          `json:"description,omitempty"`
    Drivers     []string        `json:"drivers"`
    Targets     []string        `json:"targets"`
    Effect      string          `json:"effect"`
    Conditions  *Condition      `json:"conditions,omitempty"`
    Confidence  float64         `json:"confidence,omitempty"`
    Strength    float64         `json:"strength,omitempty"`
    Window      string          `json:"window,omitempty"`
}

type Regime struct {
    RegimeID    string          `json:"regime_id"`
    Name        string          `json:"name,omitempty"`
    Description string          `json:"description,omitempty"`
    Conditions  *Condition      `json:"conditions"`
    Confidence  float64         `json:"confidence,omitempty"`
    StateMode   string          `json:"state_mode"`
    Implications []Implication  `json:"implications,omitempty"`
}

type Implication struct {
    Target     string  `json:"target"`
    Effect     string  `json:"effect"`
    Confidence float64 `json:"confidence,omitempty"`
}

type AIRPortfolio struct {
    BaseCurrency      string              `json:"base_currency"`
    StartingCash      float64             `json:"starting_cash"`
    Targets           map[string]float64  `json:"targets,omitempty"`
    SoftTargets       []SoftTarget        `json:"soft_targets,omitempty"`
    Constraints       *PortfolioConstraints `json:"constraints,omitempty"`
    RiskBudgets       *RiskBudgets        `json:"risk_budgets,omitempty"`
    Rebalance         *RebalancePolicy    `json:"rebalance,omitempty"`
    InitialAllocation InitialAllocation   `json:"initial_allocation,omitempty"`
    CandidateBaskets  []CandidateBasket   `json:"candidate_baskets,omitempty"`
    SelectionPolicy   *SelectionPolicy    `json:"selection_policy,omitempty"`
}

type DecisionHierarchy struct {
    Layers            []Layer   `json:"layers"`
    ConflictResolution []string `json:"conflict_resolution"`
    TieBreaker        string    `json:"tie_breaker"`
}

type Layer struct {
    Name     string `json:"name"`
    Priority int    `json:"priority"`
}

type Rule struct {
    RuleID      string       `json:"rule_id"`
    Name        string       `json:"name,omitempty"`
    Description string       `json:"description,omitempty"`
    Enabled     bool         `json:"enabled"`
    Layer       string       `json:"layer"`
    Priority    int          `json:"priority"`
    Confidence  float64      `json:"confidence,omitempty"`
    When        Condition    `json:"when"`
    Then        []Action     `json:"then"`
    Else        []Action     `json:"else,omitempty"`
    Cooldown    *Cooldown    `json:"cooldown,omitempty"`
    Limits      *RuleLimits  `json:"limits,omitempty"`
}

type Condition struct {
    Signal         string              `json:"signal,omitempty"`
    Relation       string              `json:"relation,omitempty"`
    Regime         string              `json:"regime,omitempty"`
    PortfolioMetric string             `json:"portfolio_metric,omitempty"`
    Operator       string              `json:"operator,omitempty"`
    Value          any                 `json:"value,omitempty"`
    Unit           string              `json:"unit,omitempty"`
    ConfidenceMin  float64             `json:"confidence_min,omitempty"`
    All            []Condition         `json:"all,omitempty"`
    Any            []Condition         `json:"any,omitempty"`
    Not            *Condition          `json:"not,omitempty"`
}

type Action struct {
    Action     string  `json:"action"`
    Target     string  `json:"target"`
    TargetType string  `json:"target_type,omitempty"`
    From       string  `json:"from,omitempty"`
    To         string  `json:"to,omitempty"`
    Amount     float64 `json:"amount,omitempty"`
    Unit       string  `json:"unit,omitempty"`
    Reason     string  `json:"reason,omitempty"`
}

type Cooldown struct {
    Enabled bool   `json:"enabled"`
    Days    int    `json:"days"`
    Scope   string `json:"scope"`
}

type RuleLimits struct {
    MaxSingleActionWeight      float64 `json:"max_single_action_weight,omitempty"`
    RequiresPortfolioApproval  bool    `json:"requires_portfolio_approval,omitempty"`
}

type ExecutionConfig struct {
    RebalanceFrequency   string            `json:"rebalance_frequency"`
    OrderTiming          string            `json:"order_timing,omitempty"`
    TransactionCostBps   int               `json:"transaction_cost_bps,omitempty"`
    SlippageBps          int               `json:"slippage_bps,omitempty"`
    AllowFractionalShares bool             `json:"allow_fractional_shares,omitempty"`
    DividendHandling      string           `json:"dividend_handling,omitempty"`
    CorporateActions      string           `json:"corporate_actions,omitempty"`
    DecisionSampling      *SamplingConfig  `json:"decision_sampling,omitempty"`
    ValuationFrequency    string           `json:"valuation_frequency,omitempty"`
    Benchmarks            []Benchmark      `json:"benchmarks,omitempty"`
}

type Benchmark struct {
    Symbol   string `json:"symbol"`
    Name     string `json:"name,omitempty"`
    Type     string `json:"type,omitempty"`
    Currency string `json:"currency,omitempty"`
    DataField string `json:"data_field,omitempty"`
}

type Provenance struct {
    CompilerVersion string            `json:"compiler_version"`
    GeneratedAt     string            `json:"generated_at"`
    Engines         []EngineConfig    `json:"engines,omitempty"`
    TrainingWindow  *TrainingWindow   `json:"training_window,omitempty"`
    SourceHashes    map[string]string `json:"source_hashes"`
    IRSHA256        string            `json:"ir_sha256"`
}
```

### 3.4 Engine Interface (`internal/engines/engine.go`)

```go
type Engine interface {
    Name() string
    Version() string
    Analyze(ctx context.Context, input EngineInput) (EngineOutput, error)
}

type EngineInput struct {
    Manifest       Manifest
    StrategyMD     string
    RulesJSON      []byte
    TrainingWindow TrainingWindow
    Universe       []string
    Portfolio      PortfolioConfig
}

type EngineOutput struct {
    Signals         []SignalSuggestion           `json:"signals,omitempty"`
    Relations       []RelationSuggestion         `json:"relations,omitempty"`
    Regimes         []RegimeSuggestion           `json:"regimes,omitempty"`
    Rules           []RuleSuggestion             `json:"rules,omitempty"`
    CandidateBaskets []CandidateBasketSuggestion `json:"candidate_baskets,omitempty"`
    SelectionPolicy *SelectionPolicySuggestion   `json:"selection_policy,omitempty"`
    Portfolio       *PortfolioSuggestion         `json:"portfolio,omitempty"`
    Notes           string                       `json:"notes,omitempty"`
}

type SignalSuggestion struct {
    Action      string  `json:"action"` // "add" | "remove" | "modify"
    Signal      Signal  `json:"signal"`
    Confidence  float64 `json:"confidence,omitempty"`
    Rationale   string  `json:"rationale,omitempty"`
}

type RelationSuggestion struct {
    Action     string   `json:"action"`
    Relation   Relation `json:"relation"`
    Confidence float64  `json:"confidence,omitempty"`
    Rationale  string   `json:"rationale,omitempty"`
}

type RegimeSuggestion struct {
    Action     string  `json:"action"`
    Regime     Regime  `json:"regime"`
    Confidence float64 `json:"confidence,omitempty"`
    Rationale  string  `json:"rationale,omitempty"`
}

type RuleSuggestion struct {
    Action     string `json:"action"`
    Rule       Rule   `json:"rule"`
    Confidence float64 `json:"confidence,omitempty"`
    Rationale  string  `json:"rationale,omitempty"`
}

type CandidateBasketSuggestion struct {
    Action         string          `json:"action"`
    Basket         CandidateBasket `json:"basket"`
    Confidence     float64         `json:"confidence,omitempty"`
    Rationale      string          `json:"rationale,omitempty"`
}

type SelectionPolicySuggestion struct {
    SelectionPolicy SelectionPolicy `json:"selection_policy"`
    Confidence     float64         `json:"confidence,omitempty"`
    Rationale      string          `json:"rationale,omitempty"`
}

type PortfolioSuggestion struct {
    Targets     map[string]float64  `json:"targets,omitempty"`
    Constraints *PortfolioConstraints `json:"constraints,omitempty"`
    RiskBudgets *RiskBudgets        `json:"risk_budgets,omitempty"`
    Confidence  float64             `json:"confidence,omitempty"`
    Rationale   string              `json:"rationale,omitempty"`
}
```

### 3.5 Validation Report (`internal/validation/report.go`)

```go
type ValidationReport struct {
    Status    string          `json:"status"` // "valid" | "invalid" | "warning"
    Schemas   []string        `json:"schemas"`
    Checks    []ValidationCheck `json:"checks"`
    Warnings  []string        `json:"warnings"`
    Errors    []string        `json:"errors"`
}

type ValidationCheck struct {
    Name   string `json:"name"`
    Status string `json:"status"` // "pass" | "fail" | "warn"
    Detail string `json:"detail,omitempty"`
}
```

---

## 4. Compilation Pipeline

The full pipeline implemented in [`internal/compiler/compiler.go`](rules-compiler/internal/compiler/compiler.go):

```
LoadSourceFiles()                   ← manifest.json, strategy.md, rules.json
    ↓
ValidateManifest()                  ← check required fields, spec_version, mode
    ↓
ValidateSeedRules()                 ← validate rules.json against rule.schema.json
    ↓
NormalizePortfolio()                ← validate + default portfolio config
    ↓
ResolveCompilerMode()               ← none | manual | single | ensemble
    ↓
[if single or ensemble]
  LoadEngines()                     ← instantiate engine adapters from manifest
    ↓
  CollectEngineFeedback()           ← call each engine's Analyze()
    ↓
  NormalizeSuggestions()            ← validate engine outputs
    ↓
  MergeUserRulesAndAgentSuggestions() ← apply merge strategy
    ↓
  MergeBasketAndSelectionPolicy()   ← incorporate basket/selection suggestions
    ↓
[endif]
    ↓
BuildAIR()                          ← assemble final AIR struct
    ↓
SemanticValidation()                ← cross-reference checks
    ↓
SchemaValidation()                  ← validate AIR against alphanet.schema.json
    ↓
WriteOutputs()                      ← write all 4 files + compute hashes
```

---

## 5. Compiler Modes

### 5.1 `none`

- Expects `compiled/strategy.ir.json` to already exist.
- Optionally validates it against schema.
- Writes provenance and reasoning marking it as precompiled.
- No engine calls.

### 5.2 `manual`

- Validates and normalizes user-provided source files.
- No engine calls.
- All signals referenced by rules must be defined by the compiler (added from a built-in signal catalog or inferred).
- Portfolio configuration is normalized from manifest.
- This is the **first milestone** to ship.

### 5.3 `single`

- One engine is configured in `manifest.json`.
- The compiler calls that engine's `Analyze()` method.
- Engine suggestions are merged with user rules.
- This is the **third milestone**.

### 5.4 `ensemble`

- Multiple engines are configured.
- The compiler calls each engine.
- Suggestions are merged using `ensemble_method`:
  - `union`: include all non-conflicting suggestions
  - `intersection`: include only suggestions supported by multiple engines
  - `weighted_vote`: score by engine weights
  - `priority_order`: earlier engines win conflicts
  - `human_review`: emit unresolved for manual handling
- This is the **third milestone** (full), **fourth milestone** (real engines).

---

## 6. Portfolio Normalization Rules

Implemented in [`internal/portfolio/normalize.go`](rules-compiler/internal/portfolio/normalize.go):

1. **`starting_cash`**: default to 100000 if missing or zero.
2. **`base_currency`**: default to `"USD"`.
3. **`initial_allocation`**:
   - If missing, default to 100% cash (mode: `"cash"`, no positions).
   - If mode is `"weights"`, validate weights sum to ~1.0 (allow 0.01 tolerance).
   - If mode is `"dollars"`, validate sum does not exceed starting_cash.
   - Validate all referenced symbols exist in universe.
4. **`candidate_baskets`**:
   - Validate basket_id uniqueness.
   - Validate each basket has at least 1 symbol.
   - Add basket symbols to universe if not already present.
   - Emit warnings for empty baskets.
5. **`selection_policy`**:
   - Default method: `"equal_weight"` if not specified.
   - Validate all basket references in policy resolve to defined baskets.
   - Validate ranking signal references will resolve.
6. **`targets`**: validate weights are 0-1; allow residual cash.
7. **`constraints`**:
   - Default `max_leverage` to 1.0.
   - Default `allow_shorting` to false.
   - Default `allow_margin` to false.
   - Validate `max_single_position` ≤ 1.0.
   - Validate sector constraints are <= 1.0.
8. **`risk_budgets`**: all optional, no defaults needed.
9. **`rebalance`**: default frequency to `"daily"`, threshold to 0.05.

---

## 7. Semantic Validation Rules

Implemented in [`internal/validation/semantic.go`](rules-compiler/internal/validation/semantic.go):

| Check | What It Validates |
|-------|-------------------|
| unique_signal_ids | No duplicate signal_id across all signals |
| unique_rule_ids | No duplicate rule_id across all rules |
| unique_regime_ids | No duplicate regime_id across all regimes |
| unique_relation_ids | No duplicate relation_id across all relations |
| unique_basket_ids | No duplicate basket_id across candidate baskets |
| signal_references_resolve | Every signal reference in rules, regimes, relations points to a defined signal |
| regime_references_resolve | Every regime reference in rules points to a defined regime |
| relation_references_resolve | Every relation reference in rules points to a defined relation |
| basket_references_resolve | Every basket reference in selection_policy and rules points to a defined basket |
| layer_exists | Every rule.layer exists in decision_hierarchy.layers |
| action_target_known | Every rule action target is a known symbol, basket, or valid target (cash, portfolio) |
| portfolio_constraints_valid | Constraints are internally consistent (e.g., cash_min < 1.0) |
| initial_allocation_valid | Initial allocation is valid per its mode |
| engines_version_pinned | Every engine has a version specified |
| training_window_valid | Training window dates are valid (start < end, lookback_days > 0) |
| operators_supported | All operators used in conditions are known and supported |

---

## 8. Rule Merge Strategy (for engine suggestions)

Implemented in [`internal/compiler/manual.go`](rules-compiler/internal/compiler/manual.go) and [`internal/compiler/ensemble.go`](rules-compiler/internal/compiler/ensemble.go):

1. **Preserve user rules** when valid — user rules are the baseline.
2. **Add compiler-generated signals** required by rules that reference signals not yet defined (e.g., look up from a built-in catalog).
3. **Add regimes** only when referenced by rules or when helpful.
4. **Add relations** when cross-asset logic is present.
5. **Add portfolio safety rules** if missing (e.g., `block_cash_below_minimum`).
6. **Preserve explicit manifest portfolio constraints** unless agent suggestion is accepted with rationale.
7. **Emit warnings** when agent suggestions conflict with user rules or portfolio configuration.
8. **Engine suggestions with confidence < 0.5** are discarded.
9. **Duplicate rule_id from engines** is rejected (user rules take precedence by default).
10. **Engine-added signals** must not conflict with existing signal_ids.

---

## 9. Default Decision Hierarchy

When the manifest does not define a decision hierarchy, the compiler uses:

```go
var DefaultLayers = []Layer{
    {Name: "portfolio_safety",       Priority: 100},
    {Name: "risk_management",        Priority: 90},
    {Name: "regime",                 Priority: 80},
    {Name: "cross_asset_relations",  Priority: 70},
    {Name: "strategy",               Priority: 60},
    {Name: "tactical",               Priority: 50},
    {Name: "experimental",           Priority: 25},
}

var DefaultConflictResolution = []string{
    "layer_priority",
    "rule_priority",
    "confidence",
    "tie_breaker",
}

var DefaultTieBreaker = "rule_order"
```

---

## 10. Provenance Generation

Implemented in [`internal/provenance/provenance.go`](rules-compiler/internal/provenance/provenance.go):

```go
type Provenance struct {
    CompilerVersion string            `json:"compiler_version"`
    GeneratedAt     string            `json:"generated_at"`
    Engines         []EngineConfig    `json:"engines,omitempty"`
    TrainingWindow  *TrainingWindow   `json:"training_window,omitempty"`
    SourceHashes    map[string]string `json:"source_hashes"`
    IRSHA256        string            `json:"ir_sha256"`
}
```

- **Compiler version**: hardcoded `"v0.1.0"` or read from build tag.
- **Generated at**: current UTC timestamp in ISO 8601 format.
- **Source hashes**: SHA-256 of each source file's canonical form.
- **IR SHA-256**: computed after canonical JSON serialization of the AIR.

---

## 11. Reasoning Generation

Implemented in [`internal/reasoning/reasoning.go`](rules-compiler/internal/reasoning/reasoning.go):

The reasoning markdown should include sections for:

1. **Strategy Intent** — summary of what the strategy does (from strategy.md)
2. **Compiler Mode** — which mode was used
3. **Accepted Rules** — list of rules that made it into AIR (user + engine)
4. **Rejected Rules** — rules that were rejected and why (e.g., schema violations, conflicts)
5. **Signals Added** — signals added by compiler or engines
6. **Regimes Added** — regimes added by compiler or engines
7. **Relations Added** — relations added by compiler or engines
8. **Candidate Baskets** — normalized basket list
9. **Selection Policy** — normalized selection policy
10. **Agent Feedback Summary** — if engines were used, summarize their output
11. **Major Design Decisions** — any notable decisions made during compilation

---

## 12. Hashing Strategy

Implemented in [`internal/air/hash.go`](rules-compiler/internal/air/hash.go):

1. Serialize to JSON with `json.Marshal()` (Go's default produces deterministic output for structs with consistent field ordering).
2. For canonical JSON, use a deterministic serializer that:
   - Sorts object keys alphabetically.
   - Uses no extra whitespace.
   - Produces consistent output for the same data.
3. Compute SHA-256 hash of canonical JSON bytes.
4. Store hash in provenance.

---

## 13. CLI Flags

Implemented in [`cmd/alphanet-compile/main.go`](rules-compiler/cmd/alphanet-compile/main.go):

```
Usage: alphanet-compile <strategy-dir> [flags]

Arguments:
  strategy-dir            Path to strategy folder (required)

Flags:
  --spec string           Path to specs directory (default: ./specs/v0.1)
  --out string            Output directory (default: <strategy-dir>/compiled)
  --mode string           Override compiler mode (none, manual, single, ensemble)
  --dry-run               Validate and print without writing files
  --emit-reasoning        Always emit reasoning.md (default: true)
  --validate-only         Only run validation, skip compilation
  --verbose               Enable verbose logging

Future flags (prepared for Milestone 4):
  --engine string         Override engine name (repeatable)
  --training-start string Training window start date (YYYY-MM-DD)
  --training-end string   Training window end date (YYYY-MM-DD)
  --lookback-days int     Training window lookback in days
  --allow-network         Allow network access during compilation
  --no-network            Disable network access
  --hosted                Use hosted compute
  --local                 Use local compute only
```

---

## 14. Implementation Milestones

### Milestone 1: Manual Mode (v0.1 MVP)

**Goal**: A working compiler that reads source files and produces valid AIR without engine calls.

Files to implement:
- [`go.mod`](rules-compiler/go.mod)
- [`cmd/alphanet-compile/main.go`](rules-compiler/cmd/alphanet-compile/main.go)
- [`internal/app/compile.go`](rules-compiler/internal/app/compile.go) — orchestrator
- [`internal/manifest/load.go`](rules-compiler/internal/manifest/load.go)
- [`internal/manifest/validate.go`](rules-compiler/internal/manifest/validate.go)
- [`internal/strategy/load.go`](rules-compiler/internal/strategy/load.go)
- [`internal/strategy/source.go`](rules-compiler/internal/strategy/source.go)
- [`internal/rules/load.go`](rules-compiler/internal/rules/load.go)
- [`internal/rules/normalize.go`](rules-compiler/internal/rules/normalize.go)
- [`internal/portfolio/normalize.go`](rules-compiler/internal/portfolio/normalize.go)
- [`internal/portfolio/baskets.go`](rules-compiler/internal/portfolio/baskets.go)
- [`internal/portfolio/selection_policy.go`](rules-compiler/internal/portfolio/selection_policy.go)
- [`internal/air/model.go`](rules-compiler/internal/air/model.go)
- [`internal/air/build.go`](rules-compiler/internal/air/build.go)
- [`internal/air/hash.go`](rules-compiler/internal/air/hash.go)
- [`internal/air/write.go`](rules-compiler/internal/air/write.go)
- [`internal/compiler/compiler.go`](rules-compiler/internal/compiler/compiler.go)
- [`internal/compiler/manual.go`](rules-compiler/internal/compiler/manual.go)
- [`internal/provenance/provenance.go`](rules-compiler/internal/provenance/provenance.go)
- [`internal/reasoning/reasoning.go`](rules-compiler/internal/reasoning/reasoning.go)
- [`internal/validation/semantic.go`](rules-compiler/internal/validation/semantic.go)
- [`internal/validation/report.go`](rules-compiler/internal/validation/report.go)

**Success Criteria**:
- `alphanet-compile ./strategies/oil-rates-growth-tech` produces `compiled/` with all 4 files
- Output validates against `alphanet.schema.json`
- Portfolio is normalized correctly

### Milestone 2: Schema Validation

**Goal**: Load JSON schemas from `specs/v0.1/` and validate all inputs/outputs against them.

Files to implement:
- [`internal/validation/schema.go`](rules-compiler/internal/validation/schema.go) — JSON schema loader + validator

**Requirements**:
- Embed or reference JSON schemas
- Validate `manifest.json` against `manifest.schema.json`
- Validate `rules.json` against `rule.schema.json`
- Validate `strategy.ir.json` against `alphanet.schema.json`
- Emit structured `validation-report.json`

### Milestone 3: Stub Engine + Single & Ensemble Modes

**Goal**: Define the Engine interface, implement stub engines, add single and ensemble compilation modes.

Files to implement:
- [`internal/engines/engine.go`](rules-compiler/internal/engines/engine.go)
- [`internal/engines/stub.go`](rules-compiler/internal/engines/stub.go)
- [`internal/compiler/single.go`](rules-compiler/internal/compiler/single.go)
- [`internal/compiler/ensemble.go`](rules-compiler/internal/compiler/ensemble.go)

### Milestone 4: Real Engine Adapters

**Goal**: Implement real adapters for TradingAgents and ai-hedge-fund.

Files to implement:
- [`internal/engines/tradingagents.go`](rules-compiler/internal/engines/tradingagents.go)
- [`internal/engines/aihedgefund.go`](rules-compiler/internal/engines/aihedgefund.go)

**Adapter patterns**:
- **CLI adapter**: Execute a command-line tool, pass input as JSON, read output from stdout
- **HTTP adapter**: Send input to a local or remote HTTP endpoint
- **gRPC adapter** (future): Use gRPC for structured communication

---

## 15. Testing Strategy

### Unit Tests (per package)

| Package | Test File | What It Tests |
|---------|-----------|---------------|
| `manifest` | `load_test.go`, `validate_test.go` | Loading valid/invalid manifests |
| `rules` | `load_test.go`, `normalize_test.go` | Rule loading, normalization, dedup |
| `portfolio` | `normalize_test.go`, `baskets_test.go`, `selection_policy_test.go` | Portfolio normalization edge cases |
| `air` | `build_test.go`, `hash_test.go`, `write_test.go` | AIR assembly, canonical hashing |
| `validation` | `schema_test.go`, `semantic_test.go` | Schema validation, semantic checks |
| `compiler` | `manual_test.go`, `single_test.go`, `ensemble_test.go` | End-to-end compilation modes |
| `engines` | `stub_test.go` | Stub engine returns valid output |

### Integration Tests

- Compile the `oil-rates-growth-tech` example strategy
- Verify output files exist and are valid JSON
- Verify AIR validates against schema
- Verify provenance contains correct hashes

### Testdata

Located in [`testdata/`](rules-compiler/testdata/):

```
testdata/
├── valid/
│   ├── minimal/
│   │   ├── manifest.json       # just name + spec_version
│   │   ├── strategy.md         # empty or minimal
│   │   └── rules.json          # empty rules
│   ├── manual/
│   │   ├── manifest.json       # manual mode with portfolio
│   │   ├── strategy.md
│   │   └── rules.json
│   ├── ensemble/
│   │   ├── manifest.json       # ensemble mode with 2 engines
│   │   ├── strategy.md
│   │   └── rules.json
│   └── precompiled/
│       ├── compiled/
│       │   └── strategy.ir.json
│       ├── manifest.json       # none mode
│       ├── strategy.md
│       └── rules.json
└── invalid/
    ├── missing-spec-version/manifest.json
    ├── unsupported-mode/manifest.json
    ├── duplicate-rule-id/rules.json + manifest.json
    ├── duplicate-basket-id/manifest.json
    ├── unresolved-signal-ref/manifest.json + rules.json
    ├── unresolved-basket-ref/manifest.json
    ├── invalid-initial-allocation/manifest.json
    ├── invalid-training-window/manifest.json
    ├── invalid-engine-config/manifest.json
    ├── invalid-rule-priority/rules.json + manifest.json
    └── invalid-portfolio-constraint/manifest.json
```

---

## 16. Dependencies

### Go Modules

```
module github.com/alphanet/rules-compiler

go 1.22

require (
    github.com/santhosh-tekuri/jsonschema/v6  // JSON schema validation
    github.com/google/uuid                      // UUID generation (optional, for IDs)
)
```

### Optional (Milestone 4 — Real Engines)

```
// For HTTP adapter:
net/http (stdlib)

// For CLI adapter:
os/exec (stdlib)
```

---

## 17. Error Handling Strategy

- **Validation errors** are collected and returned as part of `ValidationReport` — they do not halt compilation immediately.
- **Fatal errors** (e.g., cannot read file, JSON parse failure) halt compilation and return an error to the CLI.
- **Warnings** are collected but do not prevent output generation.
- The compiler always attempts to produce a `validation-report.json` even on partial failure.

---

## 18. File Writing Strategy

The [`internal/air/write.go`](rules-compiler/internal/air/write.go) writer:

1. Ensure output directory exists (`mkdir -p`).
2. Write `strategy.ir.json` — pretty-printed JSON (2-space indent).
3. Write `provenance.json` — pretty-printed JSON.
4. Write `reasoning.md` — markdown text.
5. Write `validation-report.json` — pretty-printed JSON.
6. If `--dry-run`, skip writing and print to stdout instead.

---

## 19. Schema Loading Strategy

The [`internal/validation/schema.go`](rules-compiler/internal/validation/schema.go) schema validator:

- Schema files are loaded from the `--spec` directory (default: `./specs/v0.1`).
- Schemas may reference each other via `$ref` — the validator must resolve these.
- Use `santhosh-tekuri/jsonschema` which supports `$ref` resolution from file paths.
- Schemas are cached in memory after first load.

---

## 20. Key Design Decisions

1. **Go over Python** — the plan specifies Go; the backtester is also Go; keeps the stack consistent.
2. **No LLM dependency in v0.1** — the manual mode must work without any external calls.
3. **Stub engines first** — allows testing single/ensemble modes without real agent dependencies.
4. **Schema validation from file** — schemas live in `specs/v0.1/` and are loaded at runtime, not embedded, to allow spec evolution without recompilation.
5. **Canonical JSON for hashing** — sorting keys ensures deterministic hashes across different platforms.
6. **Validation doesn't halt** — the compiler always tries to produce outputs and a validation report, even when warnings or errors exist.
7. **Portfolio is normalized, not copied** — the manifest portfolio is validated and normalized into AIR, not blindly copied.
8. **Cash as asset class** — cash is treated as a first-class asset class and allocation target.