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
	SignalID    string       `json:"signal_id"`
	Family      string       `json:"family"`
	Type        string       `json:"type,omitempty"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	Source      SignalSource `json:"source,omitempty"`
	Instrument  string       `json:"instrument,omitempty"`
	Symbol      string       `json:"symbol,omitempty"`
	Field       string       `json:"field,omitempty"`
	Transform   string       `json:"transform,omitempty"`
	Window      string       `json:"window,omitempty"`
	Frequency   string       `json:"frequency,omitempty"`
	Unit        string       `json:"unit,omitempty"`
	Date        string       `json:"date,omitempty"`
	Value       any          `json:"value,omitempty"`
	Confidence  float64      `json:"confidence,omitempty"`
	Rationale   string       `json:"rationale,omitempty"`
	ValueRange  []float64    `json:"value_range,omitempty"`
}

// SignalInterest defines a signal the backtester should compute or watch over time.
type SignalInterest struct {
	SignalID      string                    `json:"signal_id"`
	Family        string                    `json:"family"`
	Type          string                    `json:"type,omitempty"`
	Name          string                    `json:"name,omitempty"`
	Description   string                    `json:"description,omitempty"`
	Source        SignalSource              `json:"source,omitempty"`
	Instrument    string                    `json:"instrument,omitempty"`
	Symbol        string                    `json:"symbol,omitempty"`
	Field         string                    `json:"field,omitempty"`
	Transform     string                    `json:"transform,omitempty"`
	Window        string                    `json:"window,omitempty"`
	Frequency     string                    `json:"frequency,omitempty"`
	Unit          string                    `json:"unit,omitempty"`
	Reason        string                    `json:"reason,omitempty"`
	ExtractedFrom string                    `json:"extracted_from,omitempty"`
	Confidence    float64                   `json:"confidence,omitempty"`
	Tags          []string                  `json:"tags,omitempty"`
	Thresholds    []SignalInterestThreshold `json:"thresholds,omitempty"`
}

// SignalInterestThreshold captures a watched threshold extracted from strategy text or an agent report.
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
	MaxSinglePosition   *float64           `json:"max_single_position,omitempty"`
	MaxAssetClassWeight map[string]float64 `json:"max_asset_class_weight,omitempty"`
	MaxSectorWeight     map[string]float64 `json:"max_sector_weight,omitempty"`
	MaxLeverage         *float64           `json:"max_leverage,omitempty"`
	AllowShorting       bool               `json:"allow_shorting,omitempty"`
	AllowMargin         bool               `json:"allow_margin,omitempty"`
}

// RiskBudgets defines portfolio-level risk limits.
type RiskBudgets struct {
	MaxPortfolioVolatility *float64 `json:"max_portfolio_volatility,omitempty"`
	MaxDrawdown            *float64 `json:"max_drawdown,omitempty"`
	MaxDailyTurnover       *float64 `json:"max_daily_turnover,omitempty"`
	MaxTradeWeight         *float64 `json:"max_trade_weight,omitempty"`
}

// RebalancePolicy defines how and when the portfolio rebalances.
type RebalancePolicy struct {
	Frequency             string  `json:"frequency"`
	Threshold             float64 `json:"threshold,omitempty"`
	AllowPartialRebalance bool    `json:"allow_partial_rebalance,omitempty"`
}

// SoftTarget is an allocation target with a tolerance band.
type SoftTarget struct {
	Target    string  `json:"target"`
	Weight    float64 `json:"weight"`
	Tolerance float64 `json:"tolerance"`
	Priority  int     `json:"priority,omitempty"`
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
