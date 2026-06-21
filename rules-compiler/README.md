# AlphaNet Rules Compiler

The `rules-compiler` converts AlphaNet strategy source files into a compiled **AlphaNet Intermediate Representation (AIR)** artifact. It reads human-authored strategy inputs (`manifest.json`, `strategy.md`, `rules.json`, and optional source files such as `signals.json` and `signal_interests.json`) and emits deterministic artifacts consumed by the backtester.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Build](#build)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
- [Compiler Modes](#compiler-modes)
- [Strategy Folder Structure](#strategy-folder-structure)
- [Agent Engine Setup](#agent-engine-setup)
- [TradingAgents](#tradingagents)
- [AI Hedge Fund](#ai-hedge-fund)
- [Configuring Engines in manifest.json](#configuring-engines-in-manifestjson)
- [Example Workflow](#example-workflow)
- [Outputs](#outputs)
- [Engine Adapter Contract](#engine-adapter-contract)
- [Troubleshooting](#troubleshooting)

## Overview

The compiler reads source files from a strategy folder:

```text
strategy/
├── manifest.json
├── strategy.md
├── rules.json
├── signals.json             # optional concrete/user-starting signals
└── signal_interests.json    # optional tracked features to compute/watch
```

and writes outputs to `strategy/compiled/`:

```text
compiled/
├── strategy.ir.json
├── provenance.json
├── reasoning.md
└── validation-report.json
```

The compiler may call agent engines during compilation, but the backtester never calls agents. The backtester consumes only `compiled/strategy.ir.json`.

## Prerequisites

- Go 1.22+
- Python 3.11+ for agent adapters
- Agent repositories only when using `single` or `ensemble` mode:
  - `TauricResearch/TradingAgents`
  - `priley86/ai-hedge-fund`

## Build

```bash
cd rules-compiler
go build -a -o alphanet-compile ./cmd/alphanet-compile
```

Verify:

```bash
./alphanet-compile --help
```

## Quick Start

Manual mode requires no agent dependencies:

```bash
cd rules-compiler
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode manual --verbose
```

Single-engine test with TradingAgents:

```bash
TRADINGAGENTS_HOME=~/Github/TradingAgents \
ALPHANET_TA_DEBUG=1 \
ALPHANET_TA_MAX_SYMBOLS=1 \
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech \
  --mode single \
  --engine TauricResearch/TradingAgents \
  --verbose
```

Single-engine test with AI Hedge Fund:

```bash
AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
ALPHANET_AHF_DEBUG=1 \
ALPHANET_AHF_USE_POETRY=1 \
ALPHANET_AHF_MAX_SYMBOLS=1 \
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech \
  --mode single \
  --engine priley86/ai-hedge-fund \
  --verbose
```

## CLI Usage

```text
alphanet-compile <strategy-dir> [flags]

Arguments:
  strategy-dir                 Path to strategy folder containing manifest.json,
                               strategy.md, and rules.json

Flags:
  --spec string                Path to specs directory
  --out string                 Output directory (default: <strategy-dir>/compiled)
  --mode string                Override compiler mode (none, manual, single, ensemble)
  --engine string              Override engine name
  --dry-run                    Validate and print output without writing files
  --emit-reasoning             Emit reasoning.md
  --validate-only              Run validation only, skip compilation
  --verbose                    Enable verbose logging

Future/prepared flags:
  --training-start string      Training window start date
  --training-end string        Training window end date
  --lookback-days int          Training window lookback in days
  --allow-network              Allow network access during compilation
  --no-network                 Disable network access
  --hosted                     Use hosted compute
  --local                      Use local compute only
```

## Compiler Modes

| Mode | Description |
| --- | --- |
| `none` | No new compilation; useful for validation-only flows. |
| `manual` | Deterministic compile with no agent calls. |
| `single` | Runs the first enabled engine and merges its suggestions into AIR. |
| `ensemble` | Runs all enabled engines and merges suggestions according to the manifest’s ensemble settings. |

## Strategy Folder Structure

A valid strategy folder must contain:

```text
strategy/
├── manifest.json
├── strategy.md
├── rules.json
├── signals.json             # optional concrete/user-starting signals
├── signal_interests.json    # optional user-authored tracked features
└── compiled/
```

Optional source files can define reusable inputs:

- `signals.json` defines user-authored concrete signals: executable definitions, measured observations, point-in-time values, or recommendations that should appear in compiled `signals[]`.
- `signal_interests.json` defines user-authored signal interests: features the backtester should track or compute over time. These appear in compiled `signal_interests[]` alongside interests extracted from agent reports.

Rules may reference a signal id that is supplied by either `signals.json` or `signal_interests.json`. Use `signal_interests.json` when the rule depends on a feature to compute, not on an already materialized signal value.

## Agent Engine Setup

The compiler invokes local agent repositories through small Python adapters:

```text
Go compiler -> engine venv Python -> scripts/<engine>_adapter.py -> JSON
```

This keeps engine execution testable outside Go and avoids recursive wrapper execution.

### TradingAgents

Clone and install:

```bash
git clone https://github.com/TauricResearch/TradingAgents.git ~/Github/TradingAgents
cd ~/Github/TradingAgents
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

Create `~/Github/TradingAgents/.env`:

```bash
# LLM provider keys
OPENAI_API_KEY=
GOOGLE_API_KEY=
ANTHROPIC_API_KEY=
OPENROUTER_API_KEY=sk-or-v1-...

# Example OpenRouter setup
TRADINGAGENTS_LLM_PROVIDER=openrouter
TRADINGAGENTS_DEEP_THINK_LLM=z-ai/glm-5.2
TRADINGAGENTS_QUICK_THINK_LLM=deepseek/deepseek-v4-flash
```

Test the adapter directly before testing Go:

```bash
cd rules-compiler

echo '{"symbols":["QQQ"],"date":"2025-12-31","analysts":["market"]}' \
| TRADINGAGENTS_HOME=~/Github/TradingAgents \
  ALPHANET_TA_DEBUG=1 \
  ~/Github/TradingAgents/venv/bin/python \
  scripts/tradingagents_adapter.py --dry-run
```

Environment variables used by the relay compiler:

| Variable | Purpose |
| --- | --- |
| `TRADINGAGENTS_HOME` | Path to local TradingAgents clone. |
| `ALPHANET_TA_PYTHON` | Optional explicit Python path. Defaults to `TRADINGAGENTS_HOME/venv/bin/python*`. |
| `ALPHANET_TA_DEBUG` | Print adapter and engine debug logs. |
| `ALPHANET_TA_MAX_SYMBOLS` | Limit symbols during local testing. |

### AI Hedge Fund

Clone and install:

```bash
git clone https://github.com/priley86/ai-hedge-fund.git ~/Github/ai-hedge-fund
cd ~/Github/ai-hedge-fund
poetry env use /opt/homebrew/bin/python3.12 # recommended on macOS/Homebrew
poetry install
cp .env.example .env
```

AlphaNet currently targets Patrick Riley's fork because it adds non-interactive `.env` model selection used by the relay compiler. At minimum, configure one LLM provider key, the model defaults, and the financial data key expected by AI Hedge Fund:

```bash
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GROQ_API_KEY=
DEEPSEEK_API_KEY=
OPENROUTER_API_KEY=

AI_HEDGE_FUND_LLM_PROVIDER=OpenRouter
AI_HEDGE_FUND_LLM_MODEL=z-ai/glm-5.2

FINANCIAL_DATASETS_API_KEY=
```

The AI Hedge Fund adapter maps AlphaNet's engine request to the AI Hedge Fund CLI as:

```text
symbols[] -> --ticker
date      -> --end-date
analysts  -> --analysts
```

The compiler already derives symbols and dates from the strategy manifest/training window. Choose AI Hedge Fund analysts through the engine `config.analysts` array. If omitted, AlphaNet defaults AI Hedge Fund to `ben_graham` so adapter runs stay non-interactive.

Test the adapter:

```bash
cd rules-compiler

echo '{"symbols":["QQQ"],"date":"2025-12-31","analysts":["ben_graham"]}' \
| AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
  ALPHANET_AHF_DEBUG=1 \
  python3 scripts/ai_hedge_fund_adapter.py --dry-run
```

Real run with Poetry:

```bash
echo '{"symbols":["QQQ"],"date":"2025-12-31","analysts":["ben_graham"]}' \
| AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
  ALPHANET_AHF_DEBUG=1 \
  ALPHANET_AHF_USE_POETRY=1 \
  python3 scripts/ai_hedge_fund_adapter.py
```

Environment variables used by the relay compiler:

| Variable | Purpose |
| --- | --- |
| `AI_HEDGE_FUND_HOME` | Path to local `priley86/ai-hedge-fund` clone. |
| `ALPHANET_AHF_PYTHON` | Optional explicit Python path. Defaults to local venv Python when present, otherwise `python3`. |
| `ALPHANET_AHF_DEBUG` | Print adapter and engine debug logs. |
| `ALPHANET_AHF_MAX_SYMBOLS` | Limit symbols during local testing. |
| `ALPHANET_AHF_USE_POETRY` | Run `poetry run python -m src.main` from the AI Hedge Fund repo. |
| `ALPHANET_AHF_OLLAMA` | Adds `--ollama` to the AI Hedge Fund CLI. |

## Configuring Engines in manifest.json

Example engine config:

```json
{
  "compiler": {
    "mode": "ensemble",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "weight": 0.5,
        "role": "market_and_technical_reasoning",
        "symbols": ["QQQ", "NVDA", "AMD", "TLT", "USO"],
        "config": {
          "ta_path": "~/Github/TradingAgents",
          "adapter_script": "scripts/tradingagents_adapter.py",
          "analysts": ["market"],
          "max_symbols": 1
        }
      },
      {
        "name": "priley86/ai-hedge-fund",
        "version": "2026.6.17",
        "weight": 0.5,
        "role": "fundamental_sentiment_portfolio_reasoning",
        "symbols": ["QQQ", "NVDA", "AMD", "TLT", "USO"],
        "config": {
          "ahf_path": "~/Github/ai-hedge-fund",
          "adapter_script": "scripts/ai_hedge_fund_adapter.py",
          "analysts": ["ben_graham"],
          "max_symbols": 1
        }
      }
    ],
    "ensemble_method": "weighted_vote"
  }
}
```

## Example Workflow

```bash
cd rules-compiler
go build -a -o alphanet-compile ./cmd/alphanet-compile

TRADINGAGENTS_HOME=~/Github/TradingAgents \
ALPHANET_TA_DEBUG=1 \
ALPHANET_TA_MAX_SYMBOLS=1 \
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech \
  --mode single \
  --engine TauricResearch/TradingAgents \
  --verbose
```

Run the example as defined by its manifest, using both configured engines:

```bash
TRADINGAGENTS_HOME=~/Github/TradingAgents \
AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
ALPHANET_TA_DEBUG=1 \
ALPHANET_AHF_DEBUG=1 \
ALPHANET_TA_MAX_SYMBOLS=1 \
ALPHANET_AHF_MAX_SYMBOLS=1 \
ALPHANET_AHF_USE_POETRY=1 \
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech \
  --mode ensemble \
  --verbose
```

## Outputs

### `compiled/strategy.ir.json`

The canonical AIR artifact consumed by the backtester.

### `compiled/provenance.json`

Tracks source hashes, engine versions, compiler mode, and output hashes.

### `compiled/reasoning.md`

Human-readable explanation of the compilation, accepted rules, warnings, and engine notes.

### `compiled/validation-report.json`

Machine-readable validation checks. Missing source signals should be fixed in source files, not by editing generated reports.

## Engine Adapter Contract

Adapters read this JSON from stdin:

```json
{
  "symbols": ["QQQ"],
  "date": "2025-12-31",
  "analysts": ["market"]
}
```

Adapters write this JSON to stdout:

```json
{
  "signals": [],
  "notes": "adapter completed"
}
```

## Troubleshooting

### `ModuleNotFoundError` from an adapter

You are probably running the adapter with system Python. Use the engine repository’s venv Python or set the relevant env var:

```bash
export ALPHANET_TA_PYTHON=~/Github/TradingAgents/venv/bin/python
```

### TradingAgents structured-output warnings

Messages like `structured-output invocation failed ... retrying once as free text` come from TradingAgents. The adapter can still succeed if TradingAgents eventually returns a decision.

### Unresolved signal references

If validation reports unresolved signal references, add the referenced id to the strategy source, then regenerate `compiled/`.

- Use `signal_interests.json` for tracked features such as `wti_change_20d`, `ust10y_change_20d`, RSI, MACD, volatility, sentiment, or ranked-basket inputs that the backtester should compute.
- Use `signals.json` for concrete starting signals, observed values, executable signal definitions, or point-in-time recommendations that should be present in compiled `signals[]`.

## Agent report preservation and later extraction

Agent engines produce two kinds of useful output:

1. **Raw research reports**: rich markdown/text artifacts from the engine run.
2. **Compiled schema objects**: `signals`, `signal_interests`, rules, and portfolio/trading constructs.

The compiler preserves raw reports in:

```json
{
  "extensions": {
    "agent_reports": [
      {
        "engine": "TradingAgents",
        "symbol": "SPY",
        "date": "2026-06-21",
        "format": "markdown",
        "content": "# TradingAgents Research Report: SPY\\n..."
      }
    ]
  }
}
```

This is intentionally separate from schema extraction. A later relay extraction pass should convert those reports into structured AlphaNet objects:

```text
agent report markdown
  -> deterministic extractor runs on agent reports & observations
  -> signals[]           # agent decisions/recommendations and other concrete values
  -> signal_interests[]  # tracked features extracted from report text
  -> rules[]
  -> portfolio/trading constructs
```

Agent decisions still remain in the existing `signals[]` structure as dated point-in-time signals. The backtester should consume compiled AIR and should not call the agent engines daily unless a future explicit expensive runtime mode enables that behavior.

## Agent report artifacts

Raw agent reports are compiled artifacts, not canonical strategy IR payloads.

The compiler writes report bodies to sibling files under:

```text
compiled/agent-reports/
```

Example:

```text
compiled/agent-reports/tradingagents-AMD-2025-12-31.md
```

`strategy.ir.json` keeps only lightweight report references:

```json
{
  "extensions": {
    "agent_reports": [
      {
        "engine": "TradingAgents",
        "symbol": "AMD",
        "date": "2025-12-31",
        "format": "markdown",
        "path": "agent-reports/tradingagents-AMD-2025-12-31.md",
        "sha256": "sha256:..."
      }
    ]
  }
}
```

The report body is then available for a later extraction stage. The first deterministic extraction pass creates `signal_interests[]` from obvious report terms such as RSI, MACD, moving averages, Bollinger bands, ATR, support/resistance, stop-loss, and price targets. An extractor can enrich those interests into more precise rules and portfolio/trading constructs.

## Lifecycle, recommendations, and signal interests

AlphaNet distinguishes three related concepts:

- `signals[]`: executable signal definitions, measured observations, or point-in-time recommendations. Agent decisions such as `Buy`, `Sell`, `Hold`, `Overweight`, or `Trim` remain in this existing structure using `signal_kind: "recommendation"` and a `recommendation` object.
- `signal_interests[]`: user-authored or research-derived feature requests that tell the backtester what data/features to capture and compute. They include `interpretations[]` so the rule generator/backtester knows whether a computed state implies buy, sell, reduce, hold, hedge, or watch bias.
- `rules[]`: executable portfolio actions. Signal interests do not trade directly; rules consume computed signals/interests and produce portfolio actions.

Agent-extracted artifacts should use lifecycle metadata:

```json
{
  "lifecycle": {
    "source_type": "agent_extracted",
    "effective_date": "2025-12-31",
    "expires_date": "2026-03-31",
    "category_key": "AMD:close_50_sma:medium_term_trend",
    "supersedes_policy": "latest_effective_date_wins",
    "source_report_ref": "agent-reports/tradingagents-AMD-2025-12-31.md"
  }
}
```

Backtester as-of logic:

```text
For backtest date D:
  keep artifacts where effective_date <= D
  drop artifacts where expires_date exists and D > expires_date
  group by category_key
  keep latest effective_date per category_key
  evaluate rules against resolved signals
```

User-authored strategic rules normally omit `expires_date`. Agent-extracted rules, signals, and signal interests default to a 90-day `expires_date` unless the user edits or removes it.

Default agent sampling should be conservative because engines can be slow:

```json
{
  "training_window": {
    "lookback_days": 60,
    "sampling": {
      "frequency": "monthly"
    }
  }
}
```

This gives roughly two research snapshots per symbol by default, while still allowing users to opt into denser or longer agent runs.

Example flow:

```text
signal_interest:
  AMD close below 50 SMA means medium-term bearish pressure / underweight bias

rule:
  if amd_close_vs_50_sma < 0 then reduce AMD weight

backtester:
  compute close and 50 SMA from OHLCV, evaluate the rule as-of each backtest date
```
