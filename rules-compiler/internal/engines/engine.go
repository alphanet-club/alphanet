package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alphanet/rules-compiler/internal/air"
)

// Engine defines the interface for agent engines that can analyze strategy sources
// and provide suggestions during compilation.
type Engine interface {
	Name() string
	Version() string
	Analyze(ctx context.Context, input EngineInput) (EngineOutput, error)
	Init(config air.EngineConfig) error
}

// EngineInput contains all data an engine may need to analyze.
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

// EngineOutput contains suggestions produced by an engine.
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

// wrapperOutput mirrors the JSON output from tradingagents_wrapper.py
type wrapperOutput struct {
	Signals []struct {
		Action     string     `json:"action"`
		Signal     air.Signal `json:"signal"`
		Confidence float64    `json:"confidence"`
		Rationale  string     `json:"rationale"`
	} `json:"signals"`
	Notes string `json:"notes"`
}

// SignalSuggestion wraps a signal with metadata.
type SignalSuggestion struct {
	Action     string     `json:"action"`
	Signal     air.Signal `json:"signal"`
	Confidence float64    `json:"confidence,omitempty"`
	Rationale  string     `json:"rationale,omitempty"`
}

// RelationSuggestion wraps a relation with metadata.
type RelationSuggestion struct {
	Action     string       `json:"action"`
	Relation   air.Relation `json:"relation"`
	Confidence float64      `json:"confidence,omitempty"`
	Rationale  string       `json:"rationale,omitempty"`
}

// RegimeSuggestion wraps a regime with metadata.
type RegimeSuggestion struct {
	Action     string     `json:"action"`
	Regime     air.Regime `json:"regime"`
	Confidence float64    `json:"confidence,omitempty"`
	Rationale  string     `json:"rationale,omitempty"`
}

// RuleSuggestion wraps a rule with metadata.
type RuleSuggestion struct {
	Action     string   `json:"action"`
	Rule       air.Rule `json:"rule"`
	Confidence float64  `json:"confidence,omitempty"`
	Rationale  string   `json:"rationale,omitempty"`
}

// CandidateBasketSuggestion wraps a basket suggestion.
type CandidateBasketSuggestion struct {
	Action     string              `json:"action"`
	Basket     air.CandidateBasket `json:"basket"`
	Confidence float64             `json:"confidence,omitempty"`
	Rationale  string              `json:"rationale,omitempty"`
}

// SelectionPolicySuggestion wraps a selection policy suggestion.
type SelectionPolicySuggestion struct {
	SelectionPolicy air.SelectionPolicy `json:"selection_policy"`
	Confidence      float64             `json:"confidence,omitempty"`
	Rationale       string              `json:"rationale,omitempty"`
}

// PortfolioSuggestion wraps portfolio-level adjustments.
type PortfolioSuggestion struct {
	Targets     map[string]float64        `json:"targets,omitempty"`
	Constraints *air.PortfolioConstraints `json:"constraints,omitempty"`
	RiskBudgets *air.RiskBudgets          `json:"risk_budgets,omitempty"`
	Confidence  float64                   `json:"confidence,omitempty"`
	Rationale   string                    `json:"rationale,omitempty"`
}

// wrapperInput is the input JSON we send to the TradingAgents wrapper script.
type wrapperInput struct {
	Symbols  []string `json:"symbols"`
	Date     string   `json:"date"`
	Analysts []string `json:"analysts"`
}

// TradingAgentsEngine implements the Engine interface for the TradingAgents framework.
type TradingAgentsEngine struct {
	name    string
	version string
	config  air.EngineConfig

	// wrapperScript is the path to tradingagents_wrapper.py.
	wrapperScript string
}

func (e *TradingAgentsEngine) Name() string    { return e.name }
func (e *TradingAgentsEngine) Version() string { return e.version }

func (e *TradingAgentsEngine) Init(config air.EngineConfig) error {
	e.name = config.Name
	e.version = config.Version
	e.config = config
	if cfg, ok := config.Config["wrapper_script"]; ok {
		if path, ok := cfg.(string); ok && path != "" {
			e.wrapperScript = path
		}
	}
	if e.wrapperScript == "" {
		e.wrapperScript = "scripts/tradingagents_wrapper.py"
	}
	return nil
}

func (e *TradingAgentsEngine) Analyze(ctx context.Context, input EngineInput) (EngineOutput, error) {
	_ = os.Stderr
	fmt.Fprintf(os.Stderr, "[TradingAgents] starting analysis...\n")

	symbols := buildSymbolList(input)
	date := resolveDate(input)
	fmt.Fprintf(os.Stderr, "[TradingAgents] symbols=%v date=%s\n", symbols, date)

	wi := wrapperInput{Symbols: symbols, Date: date, Analysts: []string{"market"}}
	wiJSON, _ := json.Marshal(wi)
	fmt.Fprintf(os.Stderr, "[TradingAgents] input JSON (%d bytes)\n", len(wiJSON))

	wrapperArgs := []string{e.wrapperScript}
	if cfg, ok := e.config.Config["ta_path"]; ok {
		if path, ok := cfg.(string); ok && path != "" {
			wrapperArgs = append(wrapperArgs, "--ta-path", path)
		}
	}

	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		pythonCmd = "python"
	}
	fmt.Fprintf(os.Stderr, "[TradingAgents] running: %s %v\n", pythonCmd, wrapperArgs)

	ctxT, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxT, pythonCmd, wrapperArgs...)
	cmd.Stdin = bytes.NewReader(wiJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Fprintf(os.Stderr, "[TradingAgents] starting subprocess...\n")
	if err := cmd.Start(); err != nil {
		return EngineOutput{}, fmt.Errorf("start: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[TradingAgents] waiting for subprocess (PID=%d)...\n", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		if ctxT.Err() == context.DeadlineExceeded {
			return EngineOutput{}, fmt.Errorf("timed out after 600s")
		}
		return EngineOutput{}, fmt.Errorf("run: %w (stderr: %s)", err, stderr.String())
	}

	fmt.Fprintf(os.Stderr, "[TradingAgents] subprocess done\n")

	rawStdout := stdout.String()
	if len(rawStdout) > 500 {
		fmt.Fprintf(os.Stderr, "[TradingAgents] stdout (first 500): %s\n", rawStdout[:500])
	} else {
		fmt.Fprintf(os.Stderr, "[TradingAgents] stdout: %s\n", rawStdout)
	}
	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "[TradingAgents] stderr: %s\n", stderr.String())
	}

	var wo wrapperOutput
	if err := json.Unmarshal([]byte(rawStdout), &wo); err != nil {
		return EngineOutput{}, fmt.Errorf("parse: %w (raw: %s)", err, rawStdout)
	}

	out := EngineOutput{Notes: fmt.Sprintf("TradingAgents (v%s): %s", e.version, wo.Notes)}
	for _, ws := range wo.Signals {
		out.Signals = append(out.Signals, SignalSuggestion{
			Action: ws.Action, Signal: ws.Signal,
			Confidence: ws.Confidence, Rationale: ws.Rationale,
		})
	}
	return out, nil
}

func buildSymbolList(input EngineInput) []string {
	symbolSet := make(map[string]bool)
	for _, s := range input.Manifest.Universe.Symbols {
		if s != "cash" && s != "CASH" {
			symbolSet[s] = true
		}
	}
	for _, e := range input.Manifest.Compiler.Engines {
		for _, sym := range e.Symbols {
			if sym != "cash" && sym != "CASH" {
				symbolSet[sym] = true
			}
		}
	}
	syms := make([]string, 0, len(symbolSet))
	for s := range symbolSet {
		syms = append(syms, s)
	}
	for i := 0; i < len(syms); i++ {
		for j := i + 1; j < len(syms); j++ {
			if syms[i] > syms[j] {
				syms[i], syms[j] = syms[j], syms[i]
			}
		}
	}
	return syms
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

// AIHedgeFundEngine implements the Engine interface for the AI Hedge Fund framework.
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
		Notes: fmt.Sprintf("AI Hedge Fund (v%s): engine adapter not yet implemented.", e.version),
	}, nil
}
