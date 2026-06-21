package air

// AIR is the top-level AlphaNet Intermediate Representation artifact.
// It mirrors alphanet.schema.json and is the deterministic input to the backtester.
type AIR struct {
	Metadata            AIRMetadata           `json:"metadata"`
	Universe            AIRUniverse           `json:"universe"`
	Signals             []Signal              `json:"signals"`
	SignalInterests     []SignalInterest      `json:"signal_interests,omitempty"`
	Relations           []Relation            `json:"relations,omitempty"`
	Regimes             []Regime              `json:"regimes,omitempty"`
	Portfolio           AIRPortfolio          `json:"portfolio"`
	DecisionHierarchy   DecisionHierarchy     `json:"decision_hierarchy"`
	Rules               []Rule                `json:"rules"`
	Execution           ExecutionConfig       `json:"execution"`
	Provenance          *Provenance           `json:"provenance,omitempty"`
	Benchmarks          []Benchmark           `json:"benchmarks,omitempty"`
	Extensions          map[string]any        `json:"extensions,omitempty"`
	AgentReportContents []AgentReportArtifact `json:"-"`
}

// AIRMetadata contains strategy identity and compilation metadata.
type AIRMetadata struct {
	StrategyName    string   `json:"strategy_name"`
	StrategyID      string   `json:"strategy_id,omitempty"`
	Description     string   `json:"description,omitempty"`
	Author          string   `json:"author,omitempty"`
	Version         string   `json:"version,omitempty"`
	SpecVersion     string   `json:"spec_version"`
	CompilerVersion string   `json:"compiler_version"`
	GeneratedAt     string   `json:"generated_at"`
	Tags            []string `json:"tags,omitempty"`
}

// AIRUniverse defines all tradable assets and groupings in the strategy universe.
type AIRUniverse struct {
	Assets       []Asset  `json:"assets"`
	AssetClasses []string `json:"asset_classes,omitempty"`
	Sectors      []string `json:"sectors,omitempty"`
	Themes       []Theme  `json:"themes,omitempty"`
	Benchmark    string   `json:"benchmark,omitempty"`
}

// Asset represents a single tradable instrument in the universe.
type Asset struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name,omitempty"`
	AssetClass string `json:"asset_class"`
	Sector     string `json:"sector,omitempty"`
	Currency   string `json:"currency"`
	Tradable   bool   `json:"tradable"`
}

// Theme groups related assets under a named strategy theme.
type Theme struct {
	ThemeID     string   `json:"theme_id"`
	Name        string   `json:"name"`
	Members     []string `json:"members"`
	Description string   `json:"description,omitempty"`
}

// Signal defines a named observation, derived metric, or point-in-time value.
type Signal struct {
	SignalID         string            `json:"signal_id"`
	SignalKind       string            `json:"signal_kind,omitempty"`
	Family           string            `json:"family"`
	Type             string            `json:"type,omitempty"`
	Name             string            `json:"name,omitempty"`
	Description      string            `json:"description,omitempty"`
	Source           SignalSource      `json:"source,omitempty"`
	Instrument       string            `json:"instrument,omitempty"`
	Symbol           string            `json:"symbol,omitempty"`
	Field            string            `json:"field,omitempty"`
	Transform        string            `json:"transform,omitempty"`
	TransformSpec    *TransformSpec    `json:"transform_spec,omitempty"`
	Window           string            `json:"window,omitempty"`
	Frequency        string            `json:"frequency,omitempty"`
	Unit             string            `json:"unit,omitempty"`
	Date             string            `json:"date,omitempty"`
	Value            any               `json:"value,omitempty"`
	Confidence       float64           `json:"confidence,omitempty"`
	Rationale        string            `json:"rationale,omitempty"`
	Recommendation   *Recommendation   `json:"recommendation,omitempty"`
	Lifecycle        *Lifecycle        `json:"lifecycle,omitempty"`
	DataRequirements []DataRequirement `json:"data_requirements,omitempty"`
	RequiredFields   []string          `json:"required_fields,omitempty"`
	InputPrice       string            `json:"input_price,omitempty"`
	Parameters       map[string]any    `json:"parameters,omitempty"`
	ValueRange       []float64         `json:"value_range,omitempty"`
}

// SignalInterest defines a signal the backtester should compute or watch over time.
type SignalInterest struct {
	SignalID         string                         `json:"signal_id"`
	Family           string                         `json:"family"`
	Type             string                         `json:"type,omitempty"`
	Name             string                         `json:"name,omitempty"`
	Description      string                         `json:"description,omitempty"`
	Source           SignalSource                   `json:"source,omitempty"`
	Instrument       string                         `json:"instrument,omitempty"`
	Symbol           string                         `json:"symbol,omitempty"`
	Field            string                         `json:"field,omitempty"`
	Transform        string                         `json:"transform,omitempty"`
	TransformSpec    *TransformSpec                 `json:"transform_spec,omitempty"`
	Window           string                         `json:"window,omitempty"`
	Frequency        string                         `json:"frequency,omitempty"`
	Unit             string                         `json:"unit,omitempty"`
	Reason           string                         `json:"reason,omitempty"`
	ExtractedFrom    string                         `json:"extracted_from,omitempty"`
	Confidence       float64                        `json:"confidence,omitempty"`
	Tags             []string                       `json:"tags,omitempty"`
	Status           string                         `json:"status,omitempty"`
	Lifecycle        *Lifecycle                     `json:"lifecycle,omitempty"`
	Resolution       *SignalInterestResolution      `json:"resolution,omitempty"`
	DataRequirements []DataRequirement              `json:"data_requirements,omitempty"`
	RequiredFields   []string                       `json:"required_fields,omitempty"`
	InputPrice       string                         `json:"input_price,omitempty"`
	Parameters       map[string]any                 `json:"parameters,omitempty"`
	Thresholds       []SignalInterestThreshold      `json:"thresholds,omitempty"`
	Interpretations  []SignalInterestInterpretation `json:"interpretations,omitempty"`
}

// SignalInterestThreshold captures a watched threshold extracted from strategy text or an agent report.
// Lifecycle captures effective dating, expiration, and supersession behavior for time-versioned artifacts.
type Lifecycle struct {
	SourceType       string `json:"source_type,omitempty"`
	EffectiveDate    string `json:"effective_date,omitempty"`
	ExpiresDate      string `json:"expires_date,omitempty"`
	CategoryKey      string `json:"category_key,omitempty"`
	SupersedesPolicy string `json:"supersedes_policy,omitempty"`
	SourceReportRef  string `json:"source_report_ref,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

// DataRequirement tells the backtester what raw data it must fetch before computing a signal or signal interest.
type DataRequirement struct {
	Dataset         string   `json:"dataset,omitempty"`
	Provider        string   `json:"provider,omitempty"`
	Symbol          string   `json:"symbol,omitempty"`
	RequiredFields  []string `json:"required_fields,omitempty"`
	Frequency       string   `json:"frequency,omitempty"`
	Lookback        string   `json:"lookback,omitempty"`
	PriceAdjustment string   `json:"price_adjustment,omitempty"`
	RequiresOHLCV   bool     `json:"requires_ohlcv,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

// TransformSpec describes how a derived feature should be computed from raw data.
type TransformSpec struct {
	Name       string         `json:"name"`
	Inputs     []string       `json:"inputs,omitempty"`
	Window     string         `json:"window,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

// Recommendation captures buy/sell/hold-style intent inside the existing signals[] model.
type Recommendation struct {
	Action         string  `json:"action,omitempty"`
	Rating         string  `json:"rating,omitempty"`
	Direction      string  `json:"direction,omitempty"`
	EntryPrice     float64 `json:"entry_price,omitempty"`
	TargetPrice    float64 `json:"target_price,omitempty"`
	StopLoss       float64 `json:"stop_loss,omitempty"`
	TimeHorizon    string  `json:"time_horizon,omitempty"`
	HorizonDays    int     `json:"horizon_days,omitempty"`
	PositionSizing string  `json:"position_sizing,omitempty"`
	AllocationPct  float64 `json:"allocation_pct,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	Rationale      string  `json:"rationale,omitempty"`
}

// SignalInterestResolution captures whether the compiler/backtester can currently compute a signal interest.
type SignalInterestResolution struct {
	Status              string   `json:"status,omitempty"`
	DataProvider        string   `json:"data_provider,omitempty"`
	BacktesterSupported bool     `json:"backtester_supported,omitempty"`
	RequiresOHLCV       bool     `json:"requires_ohlcv,omitempty"`
	MissingFields       []string `json:"missing_fields,omitempty"`
	Notes               string   `json:"notes,omitempty"`
}

// SignalInterestInterpretation describes intent: when a computed feature is in a state, what bias should be inferred.
type SignalInterestInterpretation struct {
	InterpretationID   string                    `json:"interpretation_id,omitempty"`
	Condition          *SignalInterestCondition  `json:"condition,omitempty"`
	Meaning            string                    `json:"meaning,omitempty"`
	RecommendationBias string                    `json:"recommendation_bias,omitempty"`
	ActionBias         string                    `json:"action_bias,omitempty"`
	AppliesTo          string                    `json:"applies_to,omitempty"`
	Confidence         float64                   `json:"confidence,omitempty"`
	SourceText         string                    `json:"source_text,omitempty"`
	Thresholds         []SignalInterestThreshold `json:"thresholds,omitempty"`
	Metadata           map[string]any            `json:"metadata,omitempty"`
}

// SignalInterestCondition is a simple expression used to describe signal-interest intent.
type SignalInterestCondition struct {
	Left     string `json:"left,omitempty"`
	Operator string `json:"operator,omitempty"`
	Right    string `json:"right,omitempty"`
	Value    any    `json:"value,omitempty"`
	Unit     string `json:"unit,omitempty"`
}

type SignalInterestThreshold struct {
	Label      string `json:"label,omitempty"`
	Operator   string `json:"operator,omitempty"`
	Value      any    `json:"value,omitempty"`
	Unit       string `json:"unit,omitempty"`
	SourceText string `json:"source_text,omitempty"`
}

// AgentReportRef is the lightweight IR reference to a raw report artifact.
type AgentReportRef struct {
	Engine string `json:"engine"`
	Symbol string `json:"symbol,omitempty"`
	Date   string `json:"date,omitempty"`
	Format string `json:"format"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

// AgentReportArtifact carries report body content during output writing.
// It is intentionally excluded from strategy.ir.json.
type AgentReportArtifact struct {
	Ref     AgentReportRef `json:"ref"`
	Content string         `json:"content"`
}

// SignalSource describes where the signal data comes from.
type SignalSource struct {
	Name     string `json:"name"`
	SeriesID string `json:"series_id,omitempty"`
	Adapter  string `json:"adapter,omitempty"`
	Dataset  string `json:"dataset,omitempty"`
}

// Relation defines a cross-asset or cross-domain relationship.
type Relation struct {
	RelationID  string     `json:"relation_id"`
	Description string     `json:"description,omitempty"`
	Drivers     []string   `json:"drivers"`
	Targets     []string   `json:"targets"`
	Effect      string     `json:"effect"`
	Conditions  *Condition `json:"conditions,omitempty"`
	Confidence  float64    `json:"confidence,omitempty"`
	Strength    float64    `json:"strength,omitempty"`
	Window      string     `json:"window,omitempty"`
}

// Regime defines a named market environment state.
type Regime struct {
	RegimeID     string        `json:"regime_id"`
	Name         string        `json:"name,omitempty"`
	Description  string        `json:"description,omitempty"`
	Conditions   *Condition    `json:"conditions"`
	Confidence   float64       `json:"confidence,omitempty"`
	Lifecycle    *Lifecycle    `json:"lifecycle,omitempty"`
	StateMode    string        `json:"state_mode"`
	Implications []Implication `json:"implications,omitempty"`
}

// Implication captures the effect of a regime on a target.
type Implication struct {
	Target     string  `json:"target"`
	Effect     string  `json:"effect"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Condition is a recursive condition tree with all/any/not operators and leaf conditions.
type Condition struct {
	Signal          string      `json:"signal,omitempty"`
	Relation        string      `json:"relation,omitempty"`
	Regime          string      `json:"regime,omitempty"`
	PortfolioMetric string      `json:"portfolio_metric,omitempty"`
	Operator        string      `json:"operator,omitempty"`
	Value           any         `json:"value,omitempty"`
	Unit            string      `json:"unit,omitempty"`
	ConfidenceMin   float64     `json:"confidence_min,omitempty"`
	All             []Condition `json:"all,omitempty"`
	Any             []Condition `json:"any,omitempty"`
	Not             *Condition  `json:"not,omitempty"`
}

// Rule is a conditional decision that proposes actions.
type Rule struct {
	RuleID      string      `json:"rule_id"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Enabled     bool        `json:"enabled"`
	Layer       string      `json:"layer"`
	Priority    int         `json:"priority"`
	Confidence  float64     `json:"confidence,omitempty"`
	Lifecycle   *Lifecycle  `json:"lifecycle,omitempty"`
	When        Condition   `json:"when"`
	Then        []Action    `json:"then"`
	Else        []Action    `json:"else,omitempty"`
	Cooldown    *Cooldown   `json:"cooldown,omitempty"`
	Limits      *RuleLimits `json:"limits,omitempty"`
}

// Action is a single proposed portfolio action.
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

// Cooldown prevents a rule from re-firing too quickly.
type Cooldown struct {
	Enabled bool   `json:"enabled"`
	Days    int    `json:"days"`
	Scope   string `json:"scope"`
}

// RuleLimits constrains a rule's actions.
type RuleLimits struct {
	MaxSingleActionWeight     float64 `json:"max_single_action_weight,omitempty"`
	RequiresPortfolioApproval bool    `json:"requires_portfolio_approval,omitempty"`
}

// AIRPortfolio is the normalized portfolio configuration consumed by the backtester.
type AIRPortfolio struct {
	BaseCurrency      string                `json:"base_currency"`
	StartingCash      float64               `json:"starting_cash"`
	Targets           map[string]float64    `json:"targets,omitempty"`
	SoftTargets       []SoftTarget          `json:"soft_targets,omitempty"`
	Constraints       *PortfolioConstraints `json:"constraints,omitempty"`
	RiskBudgets       *RiskBudgets          `json:"risk_budgets,omitempty"`
	Rebalance         *RebalancePolicy      `json:"rebalance,omitempty"`
	InitialAllocation *InitialAllocation    `json:"initial_allocation,omitempty"`
	CandidateBaskets  []CandidateBasket     `json:"candidate_baskets,omitempty"`
	SelectionPolicy   *SelectionPolicy      `json:"selection_policy,omitempty"`
}

// PortfolioConstraints defines hard limits that must never be violated.
type PortfolioConstraints struct {
	CashMin             *float64           `json:"cash_min,omitempty"`
	CashMax             *float64           `json:"cash_max,omitempty"`
	MaxSinglePosition   *float64           `json:"max_single_position,omitempty"`
	MinSinglePosition   *float64           `json:"min_single_position,omitempty"`
	MaxPositions        *int               `json:"max_positions,omitempty"`
	MinPositions        *int               `json:"min_positions,omitempty"`
	MaxAssetClassWeight map[string]float64 `json:"max_asset_class_weight,omitempty"`
	MinAssetClassWeight map[string]float64 `json:"min_asset_class_weight,omitempty"`
	MaxSectorWeight     map[string]float64 `json:"max_sector_weight,omitempty"`
	MinSectorWeight     map[string]float64 `json:"min_sector_weight,omitempty"`
	MaxThemeWeight      map[string]float64 `json:"max_theme_weight,omitempty"`
	MaxFactorExposure   map[string]float64 `json:"max_factor_exposure,omitempty"`
	MinFactorExposure   map[string]float64 `json:"min_factor_exposure,omitempty"`
	MaxLeverage         *float64           `json:"max_leverage,omitempty"`
	AllowShorting       *bool              `json:"allow_shorting,omitempty"`
	AllowMargin         *bool              `json:"allow_margin,omitempty"`
	AllowedAssets       []string           `json:"allowed_assets,omitempty"`
	BlockedAssets       []string           `json:"blocked_assets,omitempty"`
	AllowedAssetClasses []string           `json:"allowed_asset_classes,omitempty"`
	BlockedAssetClasses []string           `json:"blocked_asset_classes,omitempty"`
}

// RiskBudgets defines portfolio-level risk limits.
type RiskBudgets struct {
	MaxPortfolioVolatility    *float64         `json:"max_portfolio_volatility,omitempty"`
	TargetPortfolioVolatility *float64         `json:"target_portfolio_volatility,omitempty"`
	MaxDrawdown               *float64         `json:"max_drawdown,omitempty"`
	MaxBeta                   *float64         `json:"max_beta,omitempty"`
	MinBeta                   *float64         `json:"min_beta,omitempty"`
	MaxDailyTurnover          *float64         `json:"max_daily_turnover,omitempty"`
	MaxMonthlyTurnover        *float64         `json:"max_monthly_turnover,omitempty"`
	MaxTradeWeight            *float64         `json:"max_trade_weight,omitempty"`
	MaxTradeNotional          *float64         `json:"max_trade_notional,omitempty"`
	MaxDailyLoss              *float64         `json:"max_daily_loss,omitempty"`
	MaxVaR                    *float64         `json:"max_var,omitempty"`
	MaxCVaR                   *float64         `json:"max_cvar,omitempty"`
	DeRiskingPolicy           *DeRiskingPolicy `json:"de_risking_policy,omitempty"`
}

// DeRiskingPolicy defines what to do after a risk-budget breach.
type DeRiskingPolicy struct {
	Enabled      *bool   `json:"enabled,omitempty"`
	Trigger      string  `json:"trigger,omitempty"`
	Action       string  `json:"action,omitempty"`
	Amount       float64 `json:"amount,omitempty"`
	CooldownDays int     `json:"cooldown_days,omitempty"`
}

// RebalancePolicy defines how and when the portfolio rebalances.
type RebalancePolicy struct {
	Frequency             string          `json:"frequency"`
	Threshold             float64         `json:"threshold,omitempty"`
	AllowPartialRebalance bool            `json:"allow_partial_rebalance,omitempty"`
	MinTradeWeight        float64         `json:"min_trade_weight,omitempty"`
	MaxRebalanceTurnover  float64         `json:"max_rebalance_turnover,omitempty"`
	Calendar              *CalendarConfig `json:"calendar,omitempty"`
}

// CalendarConfig defines calendar parameters for scheduled rebalances.
type CalendarConfig struct {
	DayOfWeek  string `json:"day_of_week,omitempty"`
	DayOfMonth int    `json:"day_of_month,omitempty"`
}

// SoftTarget is an allocation target with a tolerance band.
type SoftTarget struct {
	Target                string   `json:"target"`
	TargetKind            string   `json:"target_kind,omitempty"`
	Weight                float64  `json:"weight"`
	Tolerance             float64  `json:"tolerance"`
	Min                   *float64 `json:"min,omitempty"`
	Max                   *float64 `json:"max,omitempty"`
	Priority              int      `json:"priority,omitempty"`
	RebalanceWhenBreached *bool    `json:"rebalance_when_breached,omitempty"`
	Description           string   `json:"description,omitempty"`
}

// InitialAllocation defines the starting portfolio state.
type InitialAllocation struct {
	Mode      string            `json:"mode"`
	Positions []InitialPosition `json:"positions,omitempty"`
}

// InitialPosition assigns an initial holding.
type InitialPosition struct {
	Symbol string  `json:"symbol"`
	Weight float64 `json:"weight,omitempty"`
	Shares float64 `json:"shares,omitempty"`
	Amount float64 `json:"amount,omitempty"`
}

// CandidateBasket is a group of securities the portfolio may trade as a unit.
type CandidateBasket struct {
	BasketID          string   `json:"basket_id"`
	Description       string   `json:"description,omitempty"`
	AssetClass        string   `json:"asset_class,omitempty"`
	Sector            string   `json:"sector,omitempty"`
	Sectors           []string `json:"sectors,omitempty"`
	Theme             string   `json:"theme,omitempty"`
	Symbols           []string `json:"symbols"`
	Role              string   `json:"role,omitempty"`
	MinWeight         *float64 `json:"min_weight,omitempty"`
	MaxWeight         *float64 `json:"max_weight,omitempty"`
	MinPositionWeight *float64 `json:"min_position_weight,omitempty"`
	MaxPositionWeight *float64 `json:"max_position_weight,omitempty"`
}

// SelectionPolicy defines how the portfolio selects and weights positions within baskets.
type SelectionPolicy struct {
	DefaultMethod      string                           `json:"default_method"`
	RebalanceThreshold float64                          `json:"rebalance_threshold,omitempty"`
	MinHoldingDays     int                              `json:"min_holding_days,omitempty"`
	Baskets            map[string]BasketSelectionPolicy `json:"baskets,omitempty"`
}

// BasketSelectionPolicy defines per-basket selection rules.
type BasketSelectionPolicy struct {
	Method            string             `json:"method"`
	MaxPositions      int                `json:"max_positions,omitempty"`
	MinPositionWeight float64            `json:"min_position_weight,omitempty"`
	MaxPositionWeight float64            `json:"max_position_weight,omitempty"`
	Ranking           []RankingCriterion `json:"ranking,omitempty"`
	ReplacementPolicy *ReplacementPolicy `json:"replacement_policy,omitempty"`
}

// RankingCriterion defines how to score a candidate position.
type RankingCriterion struct {
	Signal    string  `json:"signal"`
	Weight    float64 `json:"weight"`
	Direction string  `json:"direction"`
}

// ReplacementPolicy controls when and how basket positions are replaced.
type ReplacementPolicy struct {
	Enabled            bool    `json:"enabled"`
	RebalanceThreshold float64 `json:"rebalance_threshold,omitempty"`
	MinHoldingDays     int     `json:"min_holding_days,omitempty"`
}

// DecisionHierarchy defines how rule conflicts are resolved.
type DecisionHierarchy struct {
	Layers             []Layer  `json:"layers"`
	ConflictResolution []string `json:"conflict_resolution"`
	TieBreaker         string   `json:"tie_breaker"`
}

// Layer groups rules by concern with a priority level.
type Layer struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// ExecutionConfig defines how orders are placed and costs are modeled.
type ExecutionConfig struct {
	RebalanceFrequency    string          `json:"rebalance_frequency"`
	OrderTiming           string          `json:"order_timing,omitempty"`
	TransactionCostBps    int             `json:"transaction_cost_bps,omitempty"`
	SlippageBps           int             `json:"slippage_bps,omitempty"`
	AllowFractionalShares bool            `json:"allow_fractional_shares,omitempty"`
	DividendHandling      string          `json:"dividend_handling,omitempty"`
	CorporateActions      string          `json:"corporate_actions,omitempty"`
	DecisionSampling      *SamplingConfig `json:"decision_sampling,omitempty"`
	ValuationFrequency    string          `json:"valuation_frequency,omitempty"`
	Benchmarks            []Benchmark     `json:"benchmarks,omitempty"`
}

// SamplingConfig defines how dates are sampled for decision points.
type SamplingConfig struct {
	Frequency         string `json:"frequency"`
	Anchor            string `json:"anchor"`
	Calendar          string `json:"calendar"`
	MissingDatePolicy string `json:"missing_date_policy,omitempty"`
	IncludeStart      bool   `json:"include_start,omitempty"`
	IncludeEnd        bool   `json:"include_end,omitempty"`
	Timezone          string `json:"timezone,omitempty"`
}

// Benchmark defines a performance benchmark for comparison.
type Benchmark struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Currency  string `json:"currency,omitempty"`
	DataField string `json:"data_field,omitempty"`
}

// Provenance records how AIR was generated.
// This struct is serialized both in provenance.json (standalone) and embedded in strategy.ir.json.
type Provenance struct {
	StrategyName    string            `json:"strategy_name,omitempty"`
	StrategyID      string            `json:"strategy_id,omitempty"`
	SpecVersion     string            `json:"spec_version,omitempty"`
	CompilerVersion string            `json:"compiler_version"`
	GeneratedAt     string            `json:"generated_at"`
	CompilerMode    string            `json:"compiler_mode,omitempty"`
	EnsembleMethod  string            `json:"ensemble_method,omitempty"`
	Engines         []EngineConfig    `json:"engines,omitempty"`
	TrainingWindow  *TrainingWindow   `json:"training_window,omitempty"`
	SourceFiles     []string          `json:"source_files,omitempty"`
	SourceHashes    map[string]string `json:"source_hashes"`
	Outputs         map[string]string `json:"outputs,omitempty"`
	IRSHA256        string            `json:"ir_sha256"`
	Notes           string            `json:"notes,omitempty"`
}

// EngineConfig configures a compiler engine.
type EngineConfig struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Weight  float64        `json:"weight,omitempty"`
	Role    string         `json:"role,omitempty"`
	Symbols []string       `json:"symbols,omitempty"`
	Enabled *bool          `json:"enabled,omitempty"`
	Config  map[string]any `json:"config,omitempty"`
	Notes   string         `json:"notes,omitempty"`
}

// TrainingWindow bounds the data the compiler may analyze.
type TrainingWindow struct {
	LookbackDays  int             `json:"lookback_days,omitempty"`
	Start         string          `json:"start,omitempty"`
	End           string          `json:"end,omitempty"`
	Sampling      *SamplingConfig `json:"sampling,omitempty"`
	IncludeRanges []DateRange     `json:"include_ranges,omitempty"`
	ExcludeRanges []DateRange     `json:"exclude_ranges,omitempty"`
}

// DateRange defines an inclusive date interval.
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Label string `json:"label,omitempty"`
}
