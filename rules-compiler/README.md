# AlphaNet Rules Compiler

The `rules-compiler` converts AlphaNet strategy source files into a compiled **AlphaNet Intermediate Representation (AIR)** artifact. It reads human-authored strategy inputs (`manifest.json`, `strategy.md`, `rules.json`, and optional source files such as `signals.json`) and emits deterministic artifacts consumed by the backtester.

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
└── signals.json        # optional but recommended when rules reference seed signals
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
  - `virattt/ai-hedge-fund`

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
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode single --verbose
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
  --dry-run                    Validate and print output without writing files
  --emit-reasoning             Emit reasoning.md
  --validate-only              Run validation only, skip compilation
  --verbose                    Enable verbose logging

Future/prepared flags:
  --engine string              Override engine name
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
└── compiled/
```

Optional source files can define reusable inputs. For this branch, `signals.json` is recommended when `rules.json` references seed signals.

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
git clone https://github.com/virattt/ai-hedge-fund.git ~/Github/ai-hedge-fund
cd ~/Github/ai-hedge-fund
poetry install
cp .env.example .env
```

At minimum, configure one LLM provider key and the financial data key expected by the project:

```bash
OPENAI_API_KEY=
ANTHROPIC_API_KEY=
GROQ_API_KEY=
DEEPSEEK_API_KEY=
FINANCIAL_DATASETS_API_KEY=
```

Test the adapter:

```bash
cd rules-compiler

echo '{"symbols":["QQQ"],"date":"2025-12-31"}' \
| AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
  ALPHANET_AHF_DEBUG=1 \
  python3 scripts/ai_hedge_fund_adapter.py --dry-run
```

Real run with Poetry:

```bash
echo '{"symbols":["QQQ"],"date":"2025-12-31"}' \
| AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
  ALPHANET_AHF_DEBUG=1 \
  ALPHANET_AHF_USE_POETRY=1 \
  python3 scripts/ai_hedge_fund_adapter.py
```

Environment variables used by the relay compiler:

| Variable | Purpose |
| --- | --- |
| `AI_HEDGE_FUND_HOME` | Path to local `virattt/ai-hedge-fund` clone. |
| `ALPHANET_AHF_PYTHON` | Optional explicit Python path. Defaults to local venv Python when present, otherwise `python3`. |
| `ALPHANET_AHF_DEBUG` | Print adapter and engine debug logs. |
| `ALPHANET_AHF_MAX_SYMBOLS` | Limit symbols during local testing. |
| `ALPHANET_AHF_USE_POETRY` | Run `poetry run python src/main.py` from the AI Hedge Fund repo. |
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
        "name": "virattt/ai-hedge-fund",
        "version": "2026.6.17",
        "weight": 0.5,
        "role": "fundamental_sentiment_portfolio_reasoning",
        "symbols": ["QQQ", "NVDA", "AMD", "TLT", "USO"],
        "config": {
          "ahf_path": "~/Github/ai-hedge-fund",
          "adapter_script": "scripts/ai_hedge_fund_adapter.py",
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
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode single --verbose
```

After AI Hedge Fund is configured:

```bash
TRADINGAGENTS_HOME=~/Github/TradingAgents \
AI_HEDGE_FUND_HOME=~/Github/ai-hedge-fund \
ALPHANET_TA_MAX_SYMBOLS=1 \
ALPHANET_AHF_MAX_SYMBOLS=1 \
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode ensemble --verbose
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

If validation reports unresolved signal references, add the signal definitions to the strategy source, usually `signals.json`, then regenerate `compiled/`.
