package air

import "encoding/json"

// Manifest is the strategy manifest loaded from manifest.json.
type Manifest struct {
	Name        string          `json:"name"`
	StrategyID  string          `json:"strategy_id,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	SpecVersion string          `json:"spec_version"`
	Version     string          `json:"version,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Compiler    CompilerConfig  `json:"compiler,omitempty"`
	Universe    UniverseConfig  `json:"universe,omitempty"`
	Portfolio   PortfolioConfig `json:"portfolio,omitempty"`
	Backtest    BacktestConfig  `json:"backtest,omitempty"`
}

// CompilerConfig defines how the compiler should behave.
type CompilerConfig struct {
	Mode               string          `json:"mode,omitempty"`
	Engines            []EngineConfig  `json:"engines,omitempty"`
	EnsembleMethod     string          `json:"ensemble_method,omitempty"`
	TrainingWindow     *TrainingWindow `json:"training_window,omitempty"`
	AllowNetwork       bool            `json:"allow_network,omitempty"`
	AllowHostedCompute bool            `json:"allow_hosted_compute,omitempty"`
	Schedule           *ScheduleConfig `json:"schedule,omitempty"`
	Notes              string          `json:"notes,omitempty"`
}

// ScheduleConfig defines when compilation should run.
type ScheduleConfig struct {
	Type string `json:"type,omitempty"`
}

// UniverseConfig defines the strategy's tradable universe.
type UniverseConfig struct {
	Symbols      []string `json:"symbols,omitempty"`
	AssetClasses []string `json:"asset_classes,omitempty"`
}

// PortfolioConfig is the authored portfolio configuration from the manifest.
type PortfolioConfig struct {
	BaseCurrency      string                `json:"base_currency,omitempty"`
	StartingCash      float64               `json:"starting_cash,omitempty"`
	InitialAllocation *InitialAllocation    `json:"initial_allocation,omitempty"`
	CandidateBaskets  []CandidateBasket     `json:"candidate_baskets,omitempty"`
	Targets           map[string]float64    `json:"targets,omitempty"`
	SoftTargets       []SoftTarget          `json:"soft_targets,omitempty"`
	Constraints       *PortfolioConstraints `json:"constraints,omitempty"`
	SelectionPolicy   *SelectionPolicy      `json:"selection_policy,omitempty"`
	RiskBudgets       *RiskBudgets          `json:"risk_budgets,omitempty"`
	Rebalance         *RebalancePolicy      `json:"rebalance,omitempty"`
}

// BacktestConfig defines backtesting defaults.
type BacktestConfig struct {
	Start              string          `json:"start,omitempty"`
	End                string          `json:"end,omitempty"`
	StartingCash       float64         `json:"starting_cash,omitempty"`
	BaseCurrency       string          `json:"base_currency,omitempty"`
	DecisionSampling   *SamplingConfig `json:"decision_sampling,omitempty"`
	ValuationFrequency string          `json:"valuation_frequency,omitempty"`
	IncludeRanges      []DateRange     `json:"include_ranges,omitempty"`
	ExcludeRanges      []DateRange     `json:"exclude_ranges,omitempty"`
	Benchmarks         []Benchmark     `json:"benchmarks,omitempty"`
}

// RulesFile is the top-level structure for rules.json.
type RulesFile struct {
	SpecVersion string `json:"spec_version,omitempty"`
	Description string `json:"description,omitempty"`
	Rules       []Rule `json:"rules"`
}

// UnmarshalJSON implements json.Unmarshaler for Manifest.
func (m *Manifest) UnmarshalJSON(data []byte) error {
	type alias Manifest
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*m = Manifest(a)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler for RulesFile.
func (rf *RulesFile) UnmarshalJSON(data []byte) error {
	type alias RulesFile
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*rf = RulesFile(a)
	return nil
}
