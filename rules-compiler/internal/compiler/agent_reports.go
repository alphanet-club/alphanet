package compiler

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alphanet/rules-compiler/internal/air"
	"github.com/alphanet/rules-compiler/internal/engines"
)

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

		base := fmt.Sprintf("%s-%s-%s%s",
			slug(report.Engine),
			strings.ToUpper(strings.TrimSpace(report.Symbol)),
			slug(report.Date),
			ext,
		)
		if strings.Trim(base, "-.") == "" {
			base = "agent-report" + ext
		}

		if seenNames[base] > 0 {
			suffix := seenNames[base] + 1
			base = strings.TrimSuffix(base, ext) + fmt.Sprintf("-%d%s", suffix, ext)
		}
		seenNames[base]++

		relPath := filepath.ToSlash(filepath.Join("agent-reports", base))
		sum := sha256.Sum256([]byte(report.Content))
		hash := fmt.Sprintf("sha256:%x", sum)

		refs = append(refs, air.AgentReportRef{
			Engine: report.Engine,
			Symbol: strings.ToUpper(strings.TrimSpace(report.Symbol)),
			Date:   report.Date,
			Format: format,
			Path:   relPath,
			SHA256: hash,
		})
		artifacts = append(artifacts, air.AgentReportArtifact{
			Ref: air.AgentReportRef{
				Engine: report.Engine,
				Symbol: strings.ToUpper(strings.TrimSpace(report.Symbol)),
				Date:   report.Date,
				Format: format,
				Path:   relPath,
				SHA256: hash,
			},
			Content: report.Content,
		})
	}

	return refs, artifacts
}

func extractSignalInterestsFromReports(reports []engines.EngineReport) []air.SignalInterest {
	type catalogItem struct {
		id          string
		family      string
		typ         string
		name        string
		transform   string
		unit        string
		needles     []string
		description string
	}

	catalog := []catalogItem{
		{"close_10_ema", "technical", "moving_average", "10 EMA", "ema_10d", "price", []string{"close_10_ema", "10 ema", "10-ema"}, "Short-term exponential moving average watched by the agent report."},
		{"close_50_sma", "technical", "moving_average", "50 SMA", "sma_50d", "price", []string{"close_50_sma", "50 sma", "50-sma"}, "Medium-term simple moving average watched by the agent report."},
		{"close_200_sma", "technical", "moving_average", "200 SMA", "sma_200d", "price", []string{"close_200_sma", "200 sma", "200-sma"}, "Long-term simple moving average watched by the agent report."},
		{"rsi", "technical", "momentum", "RSI", "rsi", "score", []string{"rsi"}, "Relative strength index watched by the agent report."},
		{"macd", "technical", "momentum", "MACD", "macd", "price", []string{"macd"}, "MACD trend/momentum indicator watched by the agent report."},
		{"macd_signal", "technical", "momentum", "MACD signal", "macd_signal", "price", []string{"macds", "macd signal"}, "MACD signal line watched by the agent report."},
		{"macd_histogram", "technical", "momentum", "MACD histogram", "macd_histogram", "price", []string{"histogram", "macd histogram"}, "MACD histogram / momentum divergence watched by the agent report."},
		{"bollinger_mid", "technical", "volatility", "Bollinger middle band", "bollinger_mid", "price", []string{"boll ", "middle band", "bollinger"}, "Bollinger-band centerline watched by the agent report."},
		{"bollinger_upper", "technical", "volatility", "Bollinger upper band", "bollinger_upper", "price", []string{"boll_ub", "upper band", "upper bollinger"}, "Bollinger upper-band resistance watched by the agent report."},
		{"bollinger_lower", "technical", "volatility", "Bollinger lower band", "bollinger_lower", "price", []string{"boll_lb", "lower band", "lower bollinger"}, "Bollinger lower-band support watched by the agent report."},
		{"atr", "risk", "volatility", "ATR", "atr", "price", []string{"atr"}, "Average true range used for stop/risk sizing in the agent report."},
		{"volume", "technical", "volume", "Volume", "volume", "shares", []string{"volume", "shares"}, "Volume / participation watched by the agent report."},
		{"support_level", "technical", "support_resistance", "Support level", "support_level", "price", []string{"support"}, "Support levels referenced by the agent report."},
		{"resistance_level", "technical", "support_resistance", "Resistance level", "resistance_level", "price", []string{"resistance"}, "Resistance levels referenced by the agent report."},
		{"price_target", "portfolio", "target", "Price target", "price_target", "price", []string{"price target", "target"}, "Price target referenced by the agent report."},
		{"stop_loss", "risk", "risk_control", "Stop loss", "stop_loss", "price", []string{"stop loss", "stop-loss", "stop:"}, "Stop-loss level referenced by the agent report."},
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
			matched := false
			for _, needle := range item.needles {
				if strings.Contains(text, strings.ToLower(needle)) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			interestID := strings.ToLower(symbol) + "_" + item.id
			if seen[interestID] {
				continue
			}
			seen[interestID] = true

			out = append(out, air.SignalInterest{
				SignalID:      interestID,
				Family:        item.family,
				Type:          item.typ,
				Name:          symbol + " " + item.name,
				Description:   item.description,
				Source:        air.SignalSource{Name: "agent_report", Adapter: report.Engine},
				Symbol:        symbol,
				Transform:     item.transform,
				Frequency:     "daily",
				Unit:          item.unit,
				Reason:        "Extracted from raw agent report artifact.",
				ExtractedFrom: report.Engine,
				Confidence:    0.55,
				Tags:          []string{"agent_extracted", "deterministic_v0"},
			})
		}
	}

	return out
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
