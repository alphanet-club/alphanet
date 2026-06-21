package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/terminal"
)

type Engine interface {
	Name() string
	Version() string
	Analyze(ctx context.Context, input EngineInput) (EngineOutput, error)
	Init(config air.EngineConfig) error
}

type EngineInput struct {
	Manifest       air.Manifest
	StrategyMD     string
	Rules          []air.Rule
	Signals        []air.Signal
	Relations      []air.Relation
	Regimes        []air.Regime
	TrainingWindow *air.TrainingWindow
	Universe       []string
	Portfolio      air.PortfolioConfig
}

type EngineOutput struct {
	Signals          []SignalSuggestion          `json:"signals,omitempty"`
	Relations        []RelationSuggestion        `json:"relations,omitempty"`
	Regimes          []RegimeSuggestion          `json:"regimes,omitempty"`
	Rules            []RuleSuggestion            `json:"rules,omitempty"`
	CandidateBaskets []CandidateBasketSuggestion `json:"candidate_baskets,omitempty"`
	SelectionPolicy  *SelectionPolicySuggestion  `json:"selection_policy,omitempty"`
	Portfolio        *PortfolioSuggestion        `json:"portfolio,omitempty"`
	Reports          []EngineReport              `json:"reports,omitempty"`
	Notes            string                      `json:"notes,omitempty"`
}

type EngineReport struct {
	Engine  string `json:"engine"`
	Symbol  string `json:"symbol,omitempty"`
	Date    string `json:"date,omitempty"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

type SignalSuggestion struct {
	Action     string     `json:"action"`
	Signal     air.Signal `json:"signal"`
	Confidence float64    `json:"confidence,omitempty"`
	Rationale  string     `json:"rationale,omitempty"`
}

type RelationSuggestion struct {
	Action     string       `json:"action"`
	Relation   air.Relation `json:"relation"`
	Confidence float64      `json:"confidence,omitempty"`
	Rationale  string       `json:"rationale,omitempty"`
}

type RegimeSuggestion struct {
	Action     string     `json:"action"`
	Regime     air.Regime `json:"regime"`
	Confidence float64    `json:"confidence,omitempty"`
	Rationale  string     `json:"rationale,omitempty"`
}

type RuleSuggestion struct {
	Action     string   `json:"action"`
	Rule       air.Rule `json:"rule"`
	Confidence float64  `json:"confidence,omitempty"`
	Rationale  string   `json:"rationale,omitempty"`
}

type CandidateBasketSuggestion struct {
	Action     string              `json:"action"`
	Basket     air.CandidateBasket `json:"basket"`
	Confidence float64             `json:"confidence,omitempty"`
	Rationale  string              `json:"rationale,omitempty"`
}

type SelectionPolicySuggestion struct {
	SelectionPolicy air.SelectionPolicy `json:"selection_policy"`
	Confidence      float64             `json:"confidence,omitempty"`
	Rationale       string              `json:"rationale,omitempty"`
}

type PortfolioSuggestion struct {
	Targets     map[string]float64        `json:"targets,omitempty"`
	Constraints *air.PortfolioConstraints `json:"constraints,omitempty"`
	RiskBudgets *air.RiskBudgets          `json:"risk_budgets,omitempty"`
	Confidence  float64                   `json:"confidence,omitempty"`
	Rationale   string                    `json:"rationale,omitempty"`
}

type adapterInput struct {
	Symbols  []string `json:"symbols"`
	Date     string   `json:"date"`
	Analysts []string `json:"analysts,omitempty"`
}

type adapterOutput struct {
	Signals []SignalSuggestion `json:"signals,omitempty"`
	Reports []EngineReport     `json:"reports,omitempty"`
	Notes   string             `json:"notes,omitempty"`
}

type pythonAdapterConfig struct {
	engineName      string
	engineVersion   string
	config          air.EngineConfig
	python          string
	adapterScript   string
	home            string
	homeEnv         string
	pythonEnv       string
	debugEnv        string
	maxSymbolsEnv   string
	defaultScript   string
	defaultAnalysts []string
	analysts        []string
	maxSymbols      int
}

func (c *pythonAdapterConfig) init(config air.EngineConfig) {
	c.engineName = config.Name
	c.engineVersion = config.Version
	c.config = config

	c.adapterScript = stringConfig(config.Config, "adapter_script")
	if c.adapterScript == "" {
		c.adapterScript = stringConfig(config.Config, "wrapper_script")
	}
	if c.adapterScript == "" {
		c.adapterScript = c.defaultScript
	}
	c.adapterScript = absPath(c.adapterScript)

	c.home = stringConfig(config.Config, "home")
	if c.home == "" {
		switch c.engineName {
		case "TauricResearch/TradingAgents", "tradingagents":
			c.home = stringConfig(config.Config, "ta_path")
		case "virattt/ai-hedge-fund", "priley86/ai-hedge-fund", "ai-hedge-fund":
			c.home = stringConfig(config.Config, "ahf_path")
		}
	}
	if c.home == "" && c.homeEnv != "" {
		c.home = os.Getenv(c.homeEnv)
	}
	c.home = expandPath(c.home)

	c.python = stringConfig(config.Config, "python")
	if c.python == "" {
		c.python = stringConfig(config.Config, "python_path")
	}
	if c.python == "" && c.pythonEnv != "" {
		c.python = os.Getenv(c.pythonEnv)
	}
	if c.python == "" && c.home != "" {
		c.python = findVenvPython(c.home)
	}
	if c.python == "" {
		c.python = "python3"
	}
	c.python = expandPath(c.python)

	c.analysts = stringSliceConfig(config.Config, "analysts")
	if len(c.analysts) == 0 {
		c.analysts = append([]string{}, c.defaultAnalysts...)
	}

	c.maxSymbols = intConfig(config.Config, "max_symbols")
	if c.maxSymbols <= 0 && c.maxSymbolsEnv != "" {
		c.maxSymbols = intEnv(c.maxSymbolsEnv)
	}
}

func (c *pythonAdapterConfig) analyze(ctx context.Context, input EngineInput) (EngineOutput, error) {
	symbols := c.symbolsForEngine(input)
	if c.maxSymbols > 0 && len(symbols) > c.maxSymbols {
		symbols = symbols[:c.maxSymbols]
	}
	if len(symbols) == 0 {
		return EngineOutput{Notes: fmt.Sprintf("%s: skipped; no symbols configured", c.engineName)}, nil
	}

	req := adapterInput{Symbols: symbols, Date: resolveDate(input), Analysts: c.analysts}
	payload, err := json.Marshal(req)
	if err != nil {
		return EngineOutput{}, fmt.Errorf("marshal %s request: %w", c.engineName, err)
	}

	args := []string{c.adapterScript}
	if c.home != "" {
		switch c.engineName {
		case "TauricResearch/TradingAgents", "tradingagents":
			args = append(args, "--ta-home", c.home)
		case "virattt/ai-hedge-fund", "priley86/ai-hedge-fund", "ai-hedge-fund":
			args = append(args, "--ahf-home", c.home)
		default:
			args = append(args, "--home", c.home)
		}
	}

	if os.Getenv(c.debugEnv) != "" || os.Getenv("ALPHANET_ENGINE_DEBUG") != "" {
		terminal.Info("[%s] python=%s adapter=%s symbols=%v date=%s", c.engineName, c.python, c.adapterScript, symbols, req.Date)
	}

	cmd := exec.CommandContext(ctx, c.python, args...)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	if c.home != "" && c.homeEnv != "" {
		cmd.Env = append(cmd.Env, c.homeEnv+"="+c.home)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return EngineOutput{}, fmt.Errorf("%s adapter failed: %w; stderr=%s; stdout=%s", c.engineName, err, tail(stderr.String(), 4000), tail(stdout.String(), 2000))
	}

	if stderr.Len() > 0 && (os.Getenv(c.debugEnv) != "" || os.Getenv("ALPHANET_ENGINE_DEBUG") != "") {
		fmt.Fprint(os.Stderr, terminal.ColorizeLogLine(stderr.String()))
	}

	raw := strings.TrimSpace(stdout.String())
	var out adapterOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return EngineOutput{}, fmt.Errorf("parse %s adapter JSON: %w; stdout=%s; stderr=%s", c.engineName, err, tail(raw, 4000), tail(stderr.String(), 2000))
	}

	return EngineOutput{
		Signals: out.Signals,
		Reports: out.Reports,
		Notes:   fmt.Sprintf("%s (v%s): %s", c.engineName, c.engineVersion, out.Notes),
	}, nil
}

func (c *pythonAdapterConfig) symbolsForEngine(input EngineInput) []string {
	set := map[string]bool{}
	for _, s := range c.config.Symbols {
		addSymbol(set, s)
	}
	if len(set) == 0 {
		for _, s := range input.Universe {
			addSymbol(set, s)
		}
	}
	if len(set) == 0 {
		for _, s := range input.Manifest.Universe.Symbols {
			addSymbol(set, s)
		}
	}
	return sortedSymbols(set)
}

type TradingAgentsEngine struct{ adapter pythonAdapterConfig }

func (e *TradingAgentsEngine) Name() string    { return e.adapter.engineName }
func (e *TradingAgentsEngine) Version() string { return e.adapter.engineVersion }

func (e *TradingAgentsEngine) Init(config air.EngineConfig) error {
	e.adapter = pythonAdapterConfig{
		homeEnv:         "TRADINGAGENTS_HOME",
		pythonEnv:       "ALPHANET_TA_PYTHON",
		debugEnv:        "ALPHANET_TA_DEBUG",
		maxSymbolsEnv:   "ALPHANET_TA_MAX_SYMBOLS",
		defaultScript:   "scripts/tradingagents_adapter.py",
		defaultAnalysts: []string{"market"},
	}
	e.adapter.init(config)
	return nil
}

func (e *TradingAgentsEngine) Analyze(ctx context.Context, input EngineInput) (EngineOutput, error) {
	return e.adapter.analyze(ctx, input)
}

type AIHedgeFundEngine struct{ adapter pythonAdapterConfig }

func (e *AIHedgeFundEngine) Name() string    { return e.adapter.engineName }
func (e *AIHedgeFundEngine) Version() string { return e.adapter.engineVersion }

func (e *AIHedgeFundEngine) Init(config air.EngineConfig) error {
	e.adapter = pythonAdapterConfig{
		homeEnv:         "AI_HEDGE_FUND_HOME",
		pythonEnv:       "ALPHANET_AHF_PYTHON",
		debugEnv:        "ALPHANET_AHF_DEBUG",
		maxSymbolsEnv:   "ALPHANET_AHF_MAX_SYMBOLS",
		defaultScript:   "scripts/ai_hedge_fund_adapter.py",
		defaultAnalysts: []string{"ben_graham"},
	}
	e.adapter.init(config)
	return nil
}

func (e *AIHedgeFundEngine) Analyze(ctx context.Context, input EngineInput) (EngineOutput, error) {
	return e.adapter.analyze(ctx, input)
}

func resolveDate(input EngineInput) string {
	if input.Manifest.Backtest.End != "" {
		return input.Manifest.Backtest.End
	}
	if input.TrainingWindow != nil && input.TrainingWindow.End != "" {
		return input.TrainingWindow.End
	}
	return "2024-05-10"
}

func addSymbol(set map[string]bool, s string) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" || s == "CASH" {
		return
	}
	set[s] = true
}

func sortedSymbols(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func stringConfig(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func stringSliceConfig(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	case string:
		if strings.TrimSpace(x) == "" {
			return nil
		}
		parts := strings.Split(x, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func intConfig(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	default:
		return 0
	}
}

func intEnv(key string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	return n
}

func findVenvPython(home string) string {
	for _, rel := range []string{
		"venv/bin/python3",
		"venv/bin/python",
		".venv/bin/python3",
		".venv/bin/python",
	} {
		p := filepath.Join(home, rel)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func absPath(path string) string {
	path = expandPath(path)
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func expandPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return os.ExpandEnv(path)
}

func tail(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}
