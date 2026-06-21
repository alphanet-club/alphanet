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
	Notes            string                      `json:"notes,omitempty"`
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
	Notes   string             `json:"notes,omitempty"`
}

type TradingAgentsEngine struct {
	name          string
	version       string
	config        air.EngineConfig
	python        string
	adapterScript string
	taPath        string
	analysts      []string
	maxSymbols    int
}

func (e *TradingAgentsEngine) Name() string    { return e.name }
func (e *TradingAgentsEngine) Version() string { return e.version }

func (e *TradingAgentsEngine) Init(config air.EngineConfig) error {
	e.name = config.Name
	e.version = config.Version
	e.config = config

	e.adapterScript = stringConfig(config.Config, "adapter_script")
	if e.adapterScript == "" {
		e.adapterScript = stringConfig(config.Config, "wrapper_script")
	}
	if e.adapterScript == "" {
		e.adapterScript = "scripts/tradingagents_adapter.py"
	}
	e.adapterScript = absPath(e.adapterScript)

	e.taPath = stringConfig(config.Config, "ta_path")
	if e.taPath == "" {
		e.taPath = os.Getenv("TRADINGAGENTS_HOME")
	}
	e.taPath = expandPath(e.taPath)

	e.python = stringConfig(config.Config, "python")
	if e.python == "" {
		e.python = stringConfig(config.Config, "python_path")
	}
	if e.python == "" {
		e.python = os.Getenv("ALPHANET_TA_PYTHON")
	}
	if e.python == "" && e.taPath != "" {
		e.python = findTradingAgentsPython(e.taPath)
	}
	if e.python == "" {
		e.python = "python3"
	}
	e.python = expandPath(e.python)

	e.analysts = stringSliceConfig(config.Config, "analysts")
	if len(e.analysts) == 0 {
		e.analysts = []string{"market"}
	}

	e.maxSymbols = intConfig(config.Config, "max_symbols")
	if e.maxSymbols <= 0 {
		e.maxSymbols = intEnv("ALPHANET_TA_MAX_SYMBOLS")
	}

	return nil
}

func (e *TradingAgentsEngine) Analyze(ctx context.Context, input EngineInput) (EngineOutput, error) {
	symbols := e.symbolsForEngine(input)
	if e.maxSymbols > 0 && len(symbols) > e.maxSymbols {
		symbols = symbols[:e.maxSymbols]
	}
	if len(symbols) == 0 {
		return EngineOutput{Notes: fmt.Sprintf("%s: skipped; no symbols configured", e.name)}, nil
	}

	req := adapterInput{
		Symbols:  symbols,
		Date:     resolveDate(input),
		Analysts: e.analysts,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return EngineOutput{}, fmt.Errorf("marshal TradingAgents request: %w", err)
	}

	args := []string{e.adapterScript}
	if e.taPath != "" {
		args = append(args, "--ta-home", e.taPath)
	}

	if os.Getenv("ALPHANET_TA_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[TradingAgents] python=%s adapter=%s symbols=%v date=%s\n", e.python, e.adapterScript, symbols, req.Date)
	}

	cmd := exec.CommandContext(ctx, e.python, args...)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	if e.taPath != "" {
		cmd.Env = append(cmd.Env, "TRADINGAGENTS_HOME="+e.taPath)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return EngineOutput{}, fmt.Errorf("TradingAgents adapter failed: %w; stderr=%s; stdout=%s", err, tail(stderr.String(), 4000), tail(stdout.String(), 2000))
	}

	if stderr.Len() > 0 && os.Getenv("ALPHANET_TA_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "%s", stderr.String())
	}

	raw := strings.TrimSpace(stdout.String())
	var out adapterOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return EngineOutput{}, fmt.Errorf("parse TradingAgents adapter JSON: %w; stdout=%s; stderr=%s", err, tail(raw, 4000), tail(stderr.String(), 2000))
	}

	return EngineOutput{
		Signals: out.Signals,
		Notes:   fmt.Sprintf("%s (v%s): %s", e.name, e.version, out.Notes),
	}, nil
}

func (e *TradingAgentsEngine) symbolsForEngine(input EngineInput) []string {
	set := map[string]bool{}
	for _, s := range e.config.Symbols {
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

type AIHedgeFundEngine struct {
	name    string
	version string
	config  air.EngineConfig
}

func (e *AIHedgeFundEngine) Name() string    { return e.name }
func (e *AIHedgeFundEngine) Version() string { return e.version }

func (e *AIHedgeFundEngine) Init(config air.EngineConfig) error {
	e.name = config.Name
	e.version = config.Version
	e.config = config
	return nil
}

func (e *AIHedgeFundEngine) Analyze(ctx context.Context, input EngineInput) (EngineOutput, error) {
	return EngineOutput{
		Notes: fmt.Sprintf("%s (v%s): adapter not implemented; no changes emitted", e.name, e.version),
	}, nil
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

func findTradingAgentsPython(taPath string) string {
	for _, rel := range []string{
		"venv/bin/python3",
		"venv/bin/python",
		".venv/bin/python3",
		".venv/bin/python",
	} {
		p := filepath.Join(taPath, rel)
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
