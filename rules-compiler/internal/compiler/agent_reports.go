package compiler

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/engines"
)

const defaultAgentRuleValidityDays = 90

type signalInterestCatalogItem struct {
	id             string
	intent         string
	family         string
	typ            string
	name           string
	transform      string
	transformName  string
	window         string
	unit           string
	requiredFields []string
	inputs         []string
	parameters     map[string]any
	requiresOHLCV  bool
	needles        []string
	description    string
	bullMeaning    string
	bearMeaning    string
	bullBias       string
	bearBias       string
	bullCondition  *air.SignalInterestCondition
	bearCondition  *air.SignalInterestCondition
}

func buildAgentReportArtifacts(reports []engines.EngineReport) ([]air.AgentReportRef, []air.AgentReportArtifact) {
	refs := make([]air.AgentReportRef, 0, len(reports))
	artifacts := make([]air.AgentReportArtifact, 0, len(reports))
	seenNames := map[string]int{}

	for _, report := range reports {
		if strings.TrimSpace(report.Content) == "" {
			continue
		}
		format := strings.TrimSpace(report.Format)
		if format == "" {
			format = "markdown"
		}
		ext := ".md"
		if format != "markdown" {
			ext = ".txt"
		}

		base := fmt.Sprintf("%s-%s-%s%s", slug(report.Engine), strings.ToUpper(strings.TrimSpace(report.Symbol)), slug(report.Date), ext)
		if strings.Trim(base, "-.") == "" {
			base = "agent-report" + ext
		}
		if seenNames[base] > 0 {
			base = strings.TrimSuffix(base, ext) + fmt.Sprintf("-%d%s", seenNames[base]+1, ext)
		}
		seenNames[base]++

		relPath := filepath.ToSlash(filepath.Join("agent-reports", base))
		sum := sha256.Sum256([]byte(report.Content))
		ref := air.AgentReportRef{
			Engine: report.Engine,
			Symbol: strings.ToUpper(strings.TrimSpace(report.Symbol)),
			Date:   report.Date,
			Format: format,
			Path:   relPath,
			SHA256: fmt.Sprintf("sha256:%x", sum),
		}
		refs = append(refs, ref)
		artifacts = append(artifacts, air.AgentReportArtifact{Ref: ref, Content: report.Content})
	}
	return refs, artifacts
}

func extractSignalInterestsFromReports(reports []engines.EngineReport) []air.SignalInterest {
	catalog := []signalInterestCatalogItem{
		{
			id: "close_10_ema", intent: "short_term_momentum", family: "technical", typ: "moving_average", name: "10 EMA",
			transform: "ema_10d", transformName: "ema", window: "10d", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			needles:     []string{"close_10_ema", "10 ema", "10-ema"},
			description: "Track price versus the 10-day EMA because the agent report used it as an immediate momentum/support filter.",
			bullMeaning: "price above the 10-day EMA indicates short-term momentum remains constructive",
			bearMeaning: "price below the 10-day EMA indicates immediate momentum weakening",
			bullBias:    "buy_or_hold", bearBias: "trim_or_reduce",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "close_10_ema"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "close_10_ema"},
		},
		{
			id: "close_50_sma", intent: "medium_term_trend", family: "technical", typ: "moving_average", name: "50 SMA",
			transform: "sma_50d", transformName: "sma", window: "50d", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			needles:     []string{"close_50_sma", "50 sma", "50-sma"},
			description: "Track price versus the 50-day SMA because the agent report used it as medium-term trend support/resistance.",
			bullMeaning: "price above the 50-day SMA indicates medium-term trend support or resistance breakout",
			bearMeaning: "price below the 50-day SMA indicates medium-term bearish pressure or failed resistance",
			bullBias:    "buy_or_increase", bearBias: "sell_or_underweight",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "close_50_sma"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "close_50_sma"},
		},
		{
			id: "close_200_sma", intent: "long_term_trend", family: "technical", typ: "moving_average", name: "200 SMA",
			transform: "sma_200d", transformName: "sma", window: "200d", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			needles:     []string{"close_200_sma", "200 sma", "200-sma"},
			description: "Track price versus the 200-day SMA because the agent report used it as the long-term trend anchor.",
			bullMeaning: "price above the 200-day SMA indicates long-term trend remains positive",
			bearMeaning: "price below the 200-day SMA indicates long-term trend breakdown",
			bullBias:    "allow_long_or_hold", bearBias: "reduce_or_avoid",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "close_200_sma"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "close_200_sma"},
		},
		{
			id: "rsi_14d", intent: "momentum_state", family: "technical", typ: "momentum", name: "RSI 14D",
			transform: "rsi_14d", transformName: "rsi", window: "14d", unit: "score",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			needles:     []string{"rsi"},
			description: "Track 14-day RSI because the agent report used RSI as a momentum and overbought/oversold gauge.",
			bullMeaning: "RSI above neutral with room below overbought indicates constructive momentum",
			bearMeaning: "RSI deterioration or overbought reversal indicates weakening momentum",
			bullBias:    "buy_or_hold", bearBias: "trim_or_watch",
			bullCondition: &air.SignalInterestCondition{Left: "rsi_14d", Operator: ">", Value: 50, Unit: "score"},
			bearCondition: &air.SignalInterestCondition{Left: "rsi_14d", Operator: "<", Value: 50, Unit: "score"},
		},
		{
			id: "macd", intent: "trend_momentum", family: "technical", typ: "momentum", name: "MACD",
			transform: "macd_12_26_9", transformName: "macd", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			parameters:  map[string]any{"fast_window": "12d", "slow_window": "26d", "signal_window": "9d"},
			needles:     []string{"macd"},
			description: "Track MACD because the agent report used MACD trend/momentum and divergence as decision evidence.",
			bullMeaning: "MACD confirmation indicates positive trend momentum",
			bearMeaning: "MACD deterioration or negative crossover indicates weakening trend momentum",
			bullBias:    "buy_or_increase", bearBias: "trim_or_reduce",
			bullCondition: &air.SignalInterestCondition{Left: "macd", Operator: ">", Right: "macd_signal"},
			bearCondition: &air.SignalInterestCondition{Left: "macd", Operator: "<", Right: "macd_signal"},
		},
		{
			id: "macd_histogram", intent: "momentum_divergence", family: "technical", typ: "momentum", name: "MACD histogram",
			transform: "macd_histogram_12_26_9", transformName: "macd_histogram", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			parameters:  map[string]any{"fast_window": "12d", "slow_window": "26d", "signal_window": "9d"},
			needles:     []string{"histogram", "macd histogram"},
			description: "Track MACD histogram because the agent report used histogram divergence/acceleration as a warning signal.",
			bullMeaning: "improving MACD histogram indicates momentum acceleration",
			bearMeaning: "negative or deteriorating MACD histogram indicates momentum deceleration",
			bullBias:    "buy_or_hold", bearBias: "trim_or_underweight",
			bullCondition: &air.SignalInterestCondition{Left: "macd_histogram", Operator: ">", Value: 0},
			bearCondition: &air.SignalInterestCondition{Left: "macd_histogram", Operator: "<", Value: 0},
		},
		{
			id: "bollinger_bands_20d", intent: "volatility_support_resistance", family: "technical", typ: "volatility", name: "Bollinger Bands 20D",
			transform: "bollinger_bands_20d", transformName: "bollinger_bands", window: "20d", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			parameters:  map[string]any{"stddev": 2},
			needles:     []string{"boll", "bollinger", "upper band", "lower band"},
			description: "Track 20-day Bollinger bands because the agent report used upper/lower bands as support, resistance, and volatility context.",
			bullMeaning: "price holding above the middle band or bouncing from lower band indicates constructive support",
			bearMeaning: "price rejecting at the upper band or breaking below the lower band indicates downside risk",
			bullBias:    "buy_or_hold", bearBias: "trim_or_reduce",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "bollinger_mid"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "bollinger_lower"},
		},
		{
			id: "atr_14d", intent: "risk_sizing", family: "risk", typ: "volatility", name: "ATR 14D",
			transform: "atr_14d", transformName: "atr", window: "14d", unit: "price",
			requiredFields: []string{"high", "low", "close"}, inputs: []string{"high", "low", "close"}, requiresOHLCV: true,
			needles:     []string{"atr"},
			description: "Track 14-day ATR because the agent report used ATR for stop placement, volatility context, and position sizing.",
			bullMeaning: "ATR provides risk sizing and stop distance context, not a standalone directional signal",
			bearMeaning: "rising ATR may require smaller position sizes or wider stops",
			bullBias:    "size_position", bearBias: "reduce_position_size",
			bullCondition: &air.SignalInterestCondition{Left: "atr_14d", Operator: "is_available"},
			bearCondition: &air.SignalInterestCondition{Left: "atr_14d", Operator: "rising"},
		},
		{
			id: "support_level", intent: "support_failure", family: "technical", typ: "support_resistance", name: "Support level",
			transform: "support_level", transformName: "support_level", unit: "price",
			requiredFields: []string{"high", "low", "close"}, inputs: []string{"high", "low", "close"}, requiresOHLCV: true,
			needles:     []string{"support"},
			description: "Track support levels because the agent report referenced support zones for entries, stops, and failure conditions.",
			bullMeaning: "price holding above support supports long exposure",
			bearMeaning: "price breaking below support implies downside risk or stop trigger",
			bullBias:    "buy_or_hold", bearBias: "sell_or_reduce",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "support_level"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "support_level"},
		},
		{
			id: "resistance_level", intent: "resistance_breakout", family: "technical", typ: "support_resistance", name: "Resistance level",
			transform: "resistance_level", transformName: "resistance_level", unit: "price",
			requiredFields: []string{"high", "low", "close"}, inputs: []string{"high", "low", "close"}, requiresOHLCV: true,
			needles:     []string{"resistance"},
			description: "Track resistance levels because the agent report referenced resistance zones for breakouts and profit targets.",
			bullMeaning: "price breaking above resistance supports buy/increase bias",
			bearMeaning: "price failing below resistance supports caution or underweight bias",
			bullBias:    "buy_or_increase", bearBias: "hold_or_underweight",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "resistance_level"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "resistance_level"},
		},
		{
			id: "stop_loss", intent: "risk_control", family: "risk", typ: "risk_control", name: "Stop loss",
			transform: "agent_report_level", transformName: "agent_report_level", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			needles:     []string{"stop loss", "stop-loss", "stop:"},
			description: "Track agent-described stop-loss levels as forward-looking risk-control interests.",
			bullMeaning: "price above stop keeps thesis active",
			bearMeaning: "price below stop invalidates thesis and should reduce exposure",
			bullBias:    "hold", bearBias: "sell_or_reduce",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: ">", Right: "stop_loss"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "stop_loss"},
		},
		{
			id: "price_target", intent: "target_tracking", family: "portfolio", typ: "target", name: "Price target",
			transform: "agent_report_level", transformName: "agent_report_level", unit: "price",
			requiredFields: []string{"close"}, inputs: []string{"close"}, requiresOHLCV: true,
			needles:     []string{"price target", "target"},
			description: "Track agent-described price targets as forward-looking portfolio/trading interests.",
			bullMeaning: "price below target leaves upside potential",
			bearMeaning: "price reaching or exceeding target suggests taking profit or reassessing",
			bullBias:    "hold_or_increase", bearBias: "trim_or_take_profit",
			bullCondition: &air.SignalInterestCondition{Left: "close", Operator: "<", Right: "price_target"},
			bearCondition: &air.SignalInterestCondition{Left: "close", Operator: ">=", Right: "price_target"},
		},
	}

	out := []air.SignalInterest{}
	seen := map[string]bool{}

	for _, report := range reports {
		symbol := strings.ToUpper(strings.TrimSpace(report.Symbol))
		if symbol == "" {
			continue
		}
		text := strings.ToLower(report.Content)
		for _, item := range catalog {
			if !matchesAny(text, item.needles) {
				continue
			}
			interestID := strings.ToLower(symbol) + "_" + item.id
			categoryKey := symbol + ":" + item.id + ":" + item.intent
			if seen[interestID] {
				continue
			}
			seen[interestID] = true

			lifecycle := reportLifecycle(report, categoryKey)
			out = append(out, air.SignalInterest{
				SignalID:      interestID,
				Family:        item.family,
				Type:          item.typ,
				Name:          symbol + " " + item.name,
				Description:   item.description,
				Source:        air.SignalSource{Name: "agent_report", Adapter: report.Engine},
				Symbol:        symbol,
				Field:         primaryField(item.requiredFields),
				Transform:     item.transform,
				TransformSpec: &air.TransformSpec{Name: item.transformName, Inputs: item.inputs, Window: item.window, Parameters: item.parameters},
				Window:        item.window,
				Frequency:     "daily",
				Unit:          item.unit,
				Reason:        "Extracted from raw agent report artifact to tell the backtester what market data/features matter and how to interpret them.",
				ExtractedFrom: report.Engine,
				Confidence:    0.6,
				Tags:          []string{"agent_extracted", "deterministic_v1"},
				Status:        "candidate",
				Lifecycle:     &lifecycle,
				Resolution: &air.SignalInterestResolution{
					Status:              "candidate",
					DataProvider:        "market_prices",
					BacktesterSupported: false,
					RequiresOHLCV:       item.requiresOHLCV,
					Notes:               "Backtester should resolve this interest against its transform registry before promoting it to an executable signal.",
				},
				DataRequirements: []air.DataRequirement{{
					Dataset:         "market_prices",
					Symbol:          symbol,
					RequiredFields:  item.requiredFields,
					Frequency:       "daily",
					Lookback:        lookbackFor(item),
					PriceAdjustment: "adjusted",
					RequiresOHLCV:   item.requiresOHLCV,
					Notes:           "Derived from agent report signal interest.",
				}},
				RequiredFields:  item.requiredFields,
				InputPrice:      inputPrice(item.requiredFields),
				Parameters:      item.parameters,
				Interpretations: interpretationsFor(item, report),
			})
		}
	}
	return out
}

func reportLifecycle(report engines.EngineReport, categoryKey string) air.Lifecycle {
	return air.Lifecycle{
		SourceType:       "agent_extracted",
		EffectiveDate:    report.Date,
		ExpiresDate:      defaultExpiresDate(report.Date),
		CategoryKey:      categoryKey,
		SupersedesPolicy: "latest_effective_date_wins",
		SourceReportRef:  filepath.ToSlash(filepath.Join("agent-reports", fmt.Sprintf("%s-%s-%s.md", slug(report.Engine), strings.ToUpper(strings.TrimSpace(report.Symbol)), slug(report.Date)))),
	}
}

func defaultExpiresDate(effectiveDate string) string {
	if strings.TrimSpace(effectiveDate) == "" {
		return ""
	}
	d, err := time.Parse("2006-01-02", effectiveDate)
	if err != nil {
		return ""
	}
	return d.AddDate(0, 0, defaultAgentRuleValidityDays).Format("2006-01-02")
}

func interpretationsFor(item signalInterestCatalogItem, report engines.EngineReport) []air.SignalInterestInterpretation {
	out := []air.SignalInterestInterpretation{}
	if item.bullMeaning != "" {
		out = append(out, air.SignalInterestInterpretation{InterpretationID: item.id + "_bullish", Condition: item.bullCondition, Meaning: item.bullMeaning, RecommendationBias: item.bullBias, ActionBias: item.bullBias, AppliesTo: "rules", Confidence: 0.55, SourceText: sourceSnippet(report.Content, item.needles)})
	}
	if item.bearMeaning != "" {
		out = append(out, air.SignalInterestInterpretation{InterpretationID: item.id + "_bearish", Condition: item.bearCondition, Meaning: item.bearMeaning, RecommendationBias: item.bearBias, ActionBias: item.bearBias, AppliesTo: "rules", Confidence: 0.55, SourceText: sourceSnippet(report.Content, item.needles)})
	}
	return out
}

func sourceSnippet(content string, needles []string) string {
	lines := strings.Split(content, "\n")
	for _, needle := range needles {
		needle = strings.ToLower(needle)
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), needle) {
				return strings.TrimSpace(line)
			}
		}
	}
	return ""
}

func matchesAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func primaryField(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func inputPrice(fields []string) string {
	for _, field := range fields {
		if field == "adjusted_close" {
			return "adjusted_close"
		}
	}
	for _, field := range fields {
		if field == "close" {
			return "close"
		}
	}
	return ""
}

func lookbackFor(item signalInterestCatalogItem) string {
	if item.window != "" {
		return item.window
	}
	if item.transformName == "macd" || item.transformName == "macd_histogram" {
		return "35d"
	}
	return "252d"
}

func slug(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "tauricresearch/tradingagents", "tradingagents")
	s = strings.ReplaceAll(s, "virattt/ai-hedge-fund", "ai-hedge-fund")
	re := regexp.MustCompile(`[^a-z0-9A-Z._-]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-._")
	if s == "" {
		return "unknown"
	}
	return s
}
