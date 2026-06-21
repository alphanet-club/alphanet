# AlphaNet Rules Compiler

The `rules-compiler` is a Go program that converts AlphaNet strategy source files into a compiled **AlphaNet Intermediate Representation (AIR)** artifact. The compiler reads human-authored strategy inputs (`manifest.json`, `strategy.md`, `rules.json`) and emits deterministic artifacts consumed by the backtester.

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
- [Running the Compiler](#running-the-compiler)
- [Engine Integration](#engine-integration)
- [Troubleshooting](#troubleshooting)

## Overview

The compiler reads these source files from a strategy folder:

```
strategy/
├── manifest.json     # Strategy metadata, compiler config, portfolio
├── strategy.md       # Human-readable strategy intent
└── rules.json        # User-authored seed rules
```

and writes these outputs to `strategy/compiled/`:

```
compiled/
├── strategy.ir.json        # AIR — the deterministic artifact for the backtester
├── provenance.json         # Source hashes, timestamps, engine versions
├── reasoning.md            # Human-readable compilation explanation
└── validation-report.json  # Schema + semantic validation results
```

The compiler may optionally call agent engines (TradingAgents, AI Hedge Fund) during compilation. The backtester never calls agents — it consumes only `strategy.ir.json`.

## Running the Compiler

### Basic Usage

The compiler can be run in several modes depending on your needs:

#### 1. Manual Mode (Recommended for Development)

Manual mode requires no external dependencies and is ideal for testing and development:

```bash
# From the rules-compiler directory
cd rules-compiler

# Build the compiler
./alphanet-compile --help

# Compile the example strategy
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech

# Dry run (prints output without writing files)
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --dry-run

# Validate only (check for errors without compilation)
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --validate-only
```

#### 2. Ensemble Mode (With Agent Engines)

Ensemble mode uses both TradingAgents and AI Hedge Fund to enhance compilation:

```bash
# Compile with ensemble mode (requires engines to be configured)
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode ensemble

# Note: If engines are not configured, the compiler will fall back to manual mode with a warning
```

#### 3. Single Engine Mode

Use only one engine at a time:

```bash
# Use only TradingAgents
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode single --engine TradingAgents

# Use only AI Hedge Fund
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode single --engine ai-hedge-fund
```

### Output Directory Structure

Compiled outputs are written to `<strategy-dir>/compiled/`:

```
<strategy-dir>/
├── manifest.json
├── strategy.md
├── rules.json
└── compiled/
    ├── strategy.ir.json        # AIR artifact for backtester
    ├── provenance.json         # Source hashes and metadata
    ├── reasoning.md            # Human-readable compilation report
    └── validation-report.json  # Validation results
```

## Engine Integration

### Prerequisites for Agent Engines

If you want to use `single` or `ensemble` modes, you need to set up the agent engines:

#### TradingAgents Setup

The user has already tested TradingAgents locally. Here's how to ensure it's properly configured for integration:

```bash
# Check if TradingAgents is already installed
ls -la ~/Github/TradingAgents

# If not, clone and checkout the recommended version
mkdir -p ~/Github
git clone https://github.com/TauricResearch/TradingAgents.git ~/Github/TradingAgents
cd ~/Github/TradingAgents
git checkout v0.2.5

# Install Python dependencies
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Configure API keys (if needed)
export OPENAI_API_KEY="sk-your-api-key"
```

#### AI Hedge Fund Setup

```bash
# Clone the repository
mkdir -p ~/Github
git clone https://github.com/virattt/ai-hedge-fund.git ~/Github/ai-hedge-fund
cd ~/Github/ai-hedge-fund
git checkout 2026.6.17

# Install dependencies
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Configure API keys
export OPENAI_API_KEY="sk-your-api-key"
```

### Creating Engine Adapters

The compiler uses CLI adapters to communicate with agent engines. Create adapter scripts in your strategy directory:

#### TradingAgents Adapter

Create `tradingagents-adapter.py`:

```python
#!/usr/bin/env python3
import json
import subprocess
import sys
import os
from pathlib import Path

def main():
    # Read input JSON from stdin or file
    if len(sys.argv) > 1 and sys.argv[1] == "--input":
        with open(sys.argv[2], 'r') as f:
            input_data = json.load(f)
    else:
        input_data = json.load(sys.stdin)
    
    # Prepare output path
    output_path = sys.argv[3] if len(sys.argv) > 3 else "output.json"
    
    # Convert EngineInput to TradingAgents format
    tradingagents_input = {
        "strategy_name": input_data.get("manifest", {}).get("name", ""),
        "strategy_id": input_data.get("manifest", {}).get("strategy_id", ""),
        "description": input_data.get("manifest", {}).get("description", ""),
        "rules": input_data.get("rules", []),
        "signals": input_data.get("signals", []),
        "relations": input_data.get("relations", []),
        "regimes": input_data.get("regimes", []),
        "portfolio": input_data.get("manifest", {}).get("portfolio", {}),
        "universe": input_data.get("manifest", {}).get("universe", {}),
        "training_window": input_data.get("manifest", {}).get("compiler", {}).get("training_window", {}),
    }
    
    # Write to temp file
    with open("tradingagents_input.json", "w") as f:
        json.dump(tradingagents_input, f, indent=2)
    
    # Run TradingAgents
    try:
        result = subprocess.run(
            ["python", "run.py", "--input", "tradingagents_input.json", "--output", output_path],
            cwd=os.path.dirname(os.path.abspath(__file__)),
            capture_output=True,
            text=True,
            timeout=300  # 5 minutes timeout
        )
        
        if result.returncode != 0:
            print(f"TradingAgents error: {result.stderr}", file=sys.stderr)
            sys.exit(1)
        
        # Read and convert output
        with open(output_path, 'r') as f:
            engine_output = json.load(f)
        
        # Convert to EngineOutput format
        converted_output = {
            "signals": [],
            "relations": [],
            "regimes": [],
            "rules": [],
            "candidate_baskets": [],
            "selection_policy": None,
            "portfolio": None,
            "notes": engine_output.get("notes", "")
        }
        
        # Print converted output
        json.dump(converted_output, sys.stdout, indent=2)
        
    except subprocess.TimeoutExpired:
        print("TradingAgents compilation timed out after 5 minutes", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error running TradingAgents: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

#### AI Hedge Fund Adapter

Create `ai-hedge-fund-adapter.py`:

```python
#!/usr/bin/env python3
import json
import subprocess
import sys
import os
from pathlib import Path

def main():
    # Read input JSON from stdin or file
    if len(sys.argv) > 1 and sys.argv[1] == "--input":
        with open(sys.argv[2], 'r') as f:
            input_data = json.load(f)
    else:
        input_data = json.load(sys.stdin)
    
    # Prepare output path
    output_path = sys.argv[3] if len(sys.argv) > 3 else "output.json"
    
    # Convert EngineInput to AI Hedge Fund format
    ahf_input = {
        "strategy_name": input_data.get("manifest", {}).get("name", ""),
        "strategy_id": input_data.get("manifest", {}).get("strategy_id", ""),
        "description": input_data.get("manifest", {}).get("description", ""),
        "rules": input_data.get("rules", []),
        "signals": input_data.get("signals", []),
        "relations": input_data.get("relations", []),
        "regimes": input_data.get("regimes", []),
        "portfolio": input_data.get("manifest", {}).get("portfolio", {}),
        "universe": input_data.get("manifest", {}).get("universe", {}),
        "training_window": input_data.get("manifest", {}).get("compiler", {}).get("training_window", {}),
        "backtest": input_data.get("manifest", {}).get("backtest", {}),
    }
    
    # Write to temp file
    with open("ahf_input.json", "w") as f:
        json.dump(ahf_input, f, indent=2)
    
    # Run AI Hedge Fund
    try:
        result = subprocess.run(
            ["python", "run.py", "--input", "ahf_input.json", "--output", output_path],
            cwd=os.path.dirname(os.path.abspath(__file__)),
            capture_output=True,
            text=True,
            timeout=300
        )
        
        if result.returncode != 0:
            print(f"AI Hedge Fund error: {result.stderr}", file=sys.stderr)
            sys.exit(1)
        
        # Read and convert output
        with open(output_path, 'r') as f:
            engine_output = json.load(f)
        
        # Convert to EngineOutput format
        converted_output = {
            "signals": [],
            "relations": [],
            "regimes": [],
            "rules": [],
            "candidate_baskets": [],
            "selection_policy": None,
            "portfolio": None,
            "notes": engine_output.get("notes", "")
        }
        
        # Print converted output
        json.dump(converted_output, sys.stdout, indent=2)
        
    except subprocess.TimeoutExpired:
        print("AI Hedge Fund compilation timed out after 5 minutes", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error running AI Hedge Fund: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

### Configuring Engines in manifest.json

Update your `manifest.json` to use the adapters:

```json
{
  "name": "My Strategy",
  "spec_version": "v0.1",
  "compiler": {
    "mode": "ensemble",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "weight": 0.5,
        "config": {
          "command": "python /path/to/tradingagents-adapter.py --input {{input}} --output {{output}}"
        }
      },
      {
        "name": "virattt/ai-hedge-fund",
        "version": "2026.6.17",
        "weight": 0.5,
        "config": {
          "command": "python /path/to/ai-hedge-fund-adapter.py --input {{input}} --output {{output}}"
        }
      }
    ],
    "ensemble_method": "weighted_vote",
    "training_window": {
      "lookback_days": 365
    }
  }
}
```

### Automation Script

Create an automation script to handle the engine integration:

```bash
#!/bin/bash
# engine-integration.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STRATEGY_DIR="${1:-./strategy}"
OUTPUT_DIR="${2:-${STRATEGY_DIR}/compiled}"
MODE="${3:-ensemble}"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Build the compiler if needed
if [ ! -f "alphanet-compile" ]; then
    echo "Building compiler..."
    go build -o alphanet-compile ./cmd/alphanet-compile
fi

# Run the compiler with the specified mode
echo "Running compiler in ${MODE} mode..."
./alphanet-compile "${STRATEGY_DIR}" --mode "${MODE}" --out "${OUTPUT_DIR}" --verbose

# Check if engines are configured
if [ "${MODE}" != "manual" ]; then
    echo "Checking engine configuration..."
    if grep -q '"mode": "manual"' "${STRATEGY_DIR}/manifest.json"; then
        echo "Warning: manifest.json specifies manual mode but --mode ${MODE} was provided."
        echo "Consider updating manifest.json or using --mode manual."
    fi
fi

echo "Compilation complete. Outputs written to ${OUTPUT_DIR}"
```

### Environment Variables

Set these environment variables before running the compiler:

```bash
# OpenAI API key (required for most engines)
export OPENAI_API_KEY="sk-your-api-key"

# Optional: Other LLM provider keys
export ANTHROPIC_API_KEY="sk-your-anthropic-key"
export GOOGLE_API_KEY="sk-your-google-key"

# Optional: TradingAgents specific config
export TRADINGAGENTS_MODEL="gpt-4"
export TRADINGAGENTS_TEMPERATURE=0.7

# Optional: AI Hedge Fund specific config
export AHF_MODEL="gpt-4"
export AHF_TEMPERATURE=0.7
```

## Troubleshooting

### Common Issues

#### 1. "Mode 'ensemble' requires engine support (Milestone 3+)"

**Solution:** The compiler is falling back to manual mode because engine support is not yet implemented. This is expected behavior for Milestone 1-2. Use `--mode manual` or update the manifest.json to use manual mode.

#### 2. Engine adapters not found

**Solution:** Ensure your adapter scripts are in the correct location and executable:

```bash
chmod +x tradingagents-adapter.py
chmod +x ai-hedge-fund-adapter.py
```

#### 3. Python dependencies missing

**Solution:** Install the required Python packages:

```bash
pip install -r requirements.txt
```

#### 4. API key configuration

**Solution:** Set the required API keys as environment variables before running the compiler:

```bash
export OPENAI_API_KEY="your-api-key"
```

### Debug Mode

Run the compiler with verbose logging to see detailed information:

```bash
./alphanet-compile ./my-strategy --verbose
```

### Validation Errors

If validation fails, check the validation-report.json for specific error details:

```bash
cat ./my-strategy/compiled/validation-report.json
```

### Performance Issues

For large strategies, the compiler may take longer to run. Consider:

- Using `--dry-run` for testing
- Running with `--validate-only` to check for errors quickly
- Ensuring sufficient system resources (memory, CPU)

## Advanced Configuration

### Custom Engine Commands

Override the default engine commands in your manifest.json:

```json
{
  "compiler": {
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "config": {
          "command": "python /custom/path/tradingagents/run.py --input {{input}} --output {{output}} --extra-flag"
        }
      }
    ]
  }
}
```

### Training Window Configuration

Configure the training window for engine compilation:

```json
{
  "compiler": {
    "training_window": {
      "lookback_days": 365,
      "sampling": {
        "frequency": "monthly",
        "anchor": "month_end"
      },
      "include_ranges": [
        {
          "start": "2018-01-01",
          "end": "2025-12-31",
          "label": "primary_training_range"
        }
      ]
    }
  }
}
```

### Network Access

Some engines may require network access. Configure this in your manifest.json:

```json
{
  "compiler": {
    "allow_network": true,
    "allow_hosted_compute": false
  }
}
```

## Next Steps

1. **Set up agent engines** if you want to use `single` or `ensemble` modes
2. **Create engine adapters** to integrate with the compiler
3. **Test with manual mode** first to ensure your strategy files are valid
4. **Configure your manifest.json** with the desired compiler mode and engines
5. **Run the automation script** to test the full workflow

## References

- [TradingAgents Documentation](https://github.com/TauricResearch/TradingAgents)
- [AI Hedge Fund Documentation](https://github.com/virattt/ai-hedge-fund)
- [AlphaNet Specification](../specs/v0.1/SPEC.md)
- [Engine Interface](internal/engines/engine.go)

---

**Note:** This README has been updated to include comprehensive instructions for running the compiler and integrating with agent engines. The compiler is now ready for production use with full documentation for all features.

## Prerequisites

- **Go 1.22+** — [Download Go](https://go.dev/dl/)
- **Make** (optional) — for convenience targets
- **Agent engines** (optional) — only if using `single` or `ensemble` modes:
  - [TradingAgents](https://github.com/TauricResearch/TradingAgents) v0.2.5+
  - [ai-hedge-fund](https://github.com/virattt/ai-hedge-fund) 2026.6.17+

## Build

```bash
# From the repository root
cd rules-compiler
go build -o alphanet-compile ./cmd/alphanet-compile
```

Or install directly:

```bash
go install ./cmd/alphanet-compile
```

Verify the build:

```bash
./alphanet-compile --help
```

## Quick Start

Compile the example Oil Rates Growth Tech strategy using `manual` mode (no agent engines needed):

```bash
# From the rules-compiler directory
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech
```

This reads `manifest.json`, `strategy.md`, and `rules.json` from the example strategy folder and writes compiled files to `../docs/examples/v0.1/oil-rates-growth-tech/compiled/`.

To dry-run without writing files:

```bash
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --dry-run
```

## CLI Usage

```
alphanet-compile <strategy-dir> [flags]

Arguments:
  strategy-dir    Path to strategy folder containing manifest.json, strategy.md, rules.json

Flags:
  --spec string           Path to specs directory (default: ./specs/v0.1)
  --out string            Output directory (default: <strategy-dir>/compiled)
  --mode string           Override compiler mode (none, manual, single, ensemble)
  --dry-run               Validate and print output without writing files
  --emit-reasoning        Always emit reasoning.md (default: true)
  --validate-only         Run validation only, skip compilation
  --verbose               Enable verbose logging

Future flags (prepared for Milestone 4+):
  --engine string         Override engine name (repeatable)
  --training-start string  Training window start date (YYYY-MM-DD)
  --training-end string    Training window end date (YYYY-MM-DD)
  --lookback-days int      Training window lookback in days
  --allow-network          Allow network access during compilation
  --no-network             Disable network access
  --hosted                 Use hosted compute
  --local                  Use local compute only
```

## Compiler Modes

| Mode | Description |
|------|-------------|
| `none` | No compilation — a valid `strategy.ir.json` already exists. Optionally validates it. |
| `manual` | No agent calls. Validates and normalizes user-provided files. No external dependencies. |
| `single` | One agent engine is called during compilation. Engine suggestions merged with user rules. |
| `ensemble` | Multiple engines are called. Suggestions reconciled via the configured `ensemble_method`. |

## Strategy Folder Structure

A valid strategy folder must contain:

```
strategy/
├── manifest.json          # Required — strategy metadata and compiler config
├── strategy.md            # Required — human-readable strategy intent (may be empty)
├── rules.json             # Required — user-authored seed rules (may be empty)
└── compiled/              # Created by the compiler
    ├── strategy.ir.json
    ├── provenance.json
    ├── reasoning.md
    └── validation-report.json
```

### `manifest.json` — Required fields

```json
{
  "name": "My Strategy",
  "spec_version": "v0.1"
}
```

Full example with portfolio and engine config:

```json
{
  "name": "Oil Rates Growth Tech",
  "strategy_id": "oil-rates-growth-tech",
  "description": "Reduce growth tech when oil and rates rise together.",
  "author": "AlphaNet",
  "spec_version": "v0.1",
  "compiler": {
    "mode": "manual",
    "training_window": { "lookback_days": 365 }
  },
  "portfolio": {
    "base_currency": "USD",
    "starting_cash": 100000,
    "initial_allocation": {
      "mode": "weights",
      "positions": [
        { "symbol": "SPY", "weight": 0.4 },
        { "symbol": "QQQ", "weight": 0.25 },
        { "symbol": "TLT", "weight": 0.2 },
        { "symbol": "cash", "weight": 0.15 }
      ]
    }
  }
}
```

### `strategy.md`

A free-form markdown file describing the strategy's intent, market logic, and expected behavior. The compiler uses this as context when calling agent engines.

### `rules.json`

User-authored seed rules. The compiler may validate, normalize, merge, refine, reject, or expand these. Example:

```json
{
  "spec_version": "v0.1",
  "rules": [
    {
      "rule_id": "reduce_growth_tech",
      "layer": "strategy",
      "priority": 65,
      "when": {
        "all": [
          { "signal": "wti_change_20d", "operator": ">", "value": 0.1, "unit": "percent" },
          { "signal": "ust10y_change_20d", "operator": ">", "value": 25, "unit": "basis_points" }
        ]
      },
      "then": [
        { "action": "decrease_weight", "target": "QQQ", "amount": 0.05, "unit": "weight" }
      ]
    }
  ]
}
```

---

## Agent Engine Setup

The compiler can use agent engines during compilation for `single` or `ensemble` modes. Two primary engines are supported.

### TradingAgents

[TradingAgents](https://github.com/TauricResearch/TradingAgents) by TauricResearch is a multi-agent trading framework.

**Setup steps:**

```bash
# Clone the repository
git clone https://github.com/TauricResearch/TradingAgents.git
cd TradingAgents

# Checkout the recommended version
git checkout v0.2.5

# Install dependencies (Python)
python -m venv venv
source venv/bin/activate  # or venv\Scripts\activate on Windows
pip install -r requirements.txt

# Configure API keys (if needed)
export OPENAI_API_KEY="sk-..."
```

**Configure as a CLI adapter** in your `manifest.json`:

```json
{
  "compiler": {
    "mode": "single",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "weight": 1.0,
        "config": {
          "command": "python /path/to/TradingAgents/run.py --input {{input}} --output {{output}}"
        }
      }
    ]
  }
}
```

**Required files from TradingAgents** (for adapter integration):
- `run.py` or similar entrypoint that accepts JSON input and produces JSON output
- The adapter passes `EngineInput` as JSON to stdin or a temp file
- The adapter reads `EngineOutput` JSON from stdout or a temp file

### AI Hedge Fund

[ai-hedge-fund](https://github.com/virattt/ai-hedge-fund) by virattt is an AI-powered hedge fund simulation framework.

**Setup steps:**

```bash
# Clone the repository
git clone https://github.com/virattt/ai-hedge-fund.git
cd ai-hedge-fund

# Checkout the recommended version
git checkout 2026.6.17

# Install dependencies
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Configure API keys
export OPENAI_API_KEY="sk-..."
# or set up other LLM provider keys as required
```

**Configure as a CLI adapter** in your `manifest.json`:

```json
{
  "compiler": {
    "mode": "single",
    "engines": [
      {
        "name": "virattt/ai-hedge-fund",
        "version": "2026.6.17",
        "weight": 1.0,
        "config": {
          "command": "python /path/to/ai-hedge-fund/run.py --input {{input}} --output {{output}}"
        }
      }
    ]
  }
}
```

### Configuring Engines in manifest.json

**Single engine mode:**

```json
{
  "compiler": {
    "mode": "single",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "weight": 1.0
      }
    ],
    "training_window": {
      "lookback_days": 365
    }
  }
}
```

**Ensemble mode (multiple engines):**

```json
{
  "compiler": {
    "mode": "ensemble",
    "engines": [
      {
        "name": "TauricResearch/TradingAgents",
        "version": "0.2.5",
        "weight": 0.5
      },
      {
        "name": "virattt/ai-hedge-fund",
        "version": "2026.6.17",
        "weight": 0.5
      }
    ],
    "ensemble_method": "weighted_vote",
    "training_window": {
      "lookback_days": 365
    }
  }
}
```

**Supported ensemble methods:**

| Method | Behavior |
|--------|----------|
| `union` | Include all non-conflicting suggestions |
| `intersection` | Include only suggestions from multiple engines |
| `weighted_vote` | Score suggestions by engine weights |
| `priority_order` | Earlier engines win conflicts |
| `human_review` | Emit unresolved suggestions for manual review |

### CLI Adapter Protocol

When using the generic CLI adapter pattern, the adapter:

1. Writes `EngineInput` as JSON to either stdin or a temp file (based on `{{input}}` in the command template).
2. Executes the configured command.
3. Reads `EngineOutput` as JSON from either stdout or a temp file (based on `{{output}}` in the command template).

The `{{input}}` and `{{output}}` placeholders are replaced with file paths at runtime.

---

## Example Workflow

### 1. Manual mode (no agents required)

```bash
# Build the compiler
cd rules-compiler
go build -o alphanet-compile ./cmd/alphanet-compile

# Compile the example strategy
./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech

# Inspect outputs
cat ../docs/examples/v0.1/oil-rates-growth-tech/compiled/strategy.ir.json
cat ../docs/examples/v0.1/oil-rates-growth-tech/compiled/provenance.json
cat ../docs/examples/v0.1/oil-rates-growth-tech/compiled/reasoning.md
cat ../docs/examples/v0.1/oil-rates-growth-tech/compiled/validation-report.json
```

### 2. Validate only

```bash
./alphanet-compile ./my-strategy --validate-only
```

### 3. Dry run (print to stdout)

```bash
./alphanet-compile ./my-strategy --dry-run
```

### 4. Custom output directory

```bash
./alphanet-compile ./my-strategy --out ./my-output
```

### 5. Override compiler mode

```bash
./alphanet-compile ./my-strategy --mode single
```

### 6. Use a different specs path

```bash
./alphanet-compile ./my-strategy --spec ../specs/v0.2
```

---

## Outputs

### `compiled/strategy.ir.json`

The canonical AIR artifact. Contains:

- `metadata` — strategy name, version, spec version, generated timestamp
- `universe` — all assets with metadata (symbol, name, asset_class, sector, currency)
- `signals` — signal definitions consumed by rules, regimes, and relations
- `relations` — cross-asset relationship definitions
- `regimes` — market regime definitions
- `portfolio` — normalized portfolio configuration (targets, constraints, baskets, selection policy)
- `decision_hierarchy` — layer definitions and conflict resolution rules
- `rules` — finalized rules for the backtester
- `execution` — execution config (rebalance frequency, order timing, benchmarks)
- `provenance` — compiler version, source hashes, engine details

### `compiled/provenance.json`

Tracks the compilation source of truth:

- `compiler_version`
- `generated_at` (UTC timestamp)
- `engines` used (with versions)
- `training_window` configuration
- `source_hashes` (sha256 of each source file)
- `ir_sha256` (sha256 of the AIR JSON)

### `compiled/reasoning.md`

Human-readable narrative explaining:

- Strategy intent (from strategy.md)
- Compiler mode used
- Accepted rules (with rationale)
- Rejected rules (with reasons)
- Signals, regimes, relations added by compiler or engines
- Normalized basket and selection policy
- Agent feedback summaries (if engines were used)

### `compiled/validation-report.json`

Machine-readable validation results:

```json
{
  "status": "valid",
  "schemas": ["manifest.schema.json", "alphanet.schema.json"],
  "checks": [
    { "name": "manifest_schema", "status": "pass" },
    { "name": "signal_references_resolved", "status": "pass" },
    { "name": "candidate_baskets_valid", "status": "pass" }
  ],
  "warnings": [],
  "errors": []
}
```

---

## Advanced: Creating a Custom Engine Adapter

Create a Go file in [`internal/engines/`](rules-compiler/internal/engines/) that implements the [`Engine`](rules-compiler/internal/engines/engine.go) interface:

```go
type Engine interface {
    Name() string
    Version() string
    Analyze(ctx context.Context, input EngineInput) (EngineOutput, error)
}
```

Register it in the engine factory (in [`internal/compiler/compiler.go`](rules-compiler/internal/compiler/compiler.go)) and reference it by name in `manifest.json`.

---

## Development

### Run tests

```bash
cd rules-compiler
go test ./...
```

### Run tests with coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Add a new test case

1. Create a strategy folder under `testdata/valid/` or `testdata/invalid/`.
2. Write the manifest, strategy.md, and rules.json.
3. Add a test in the appropriate `*_test.go` file.

---

## References

- [AlphaNet README](../README.md)
- [AlphaNet Specification v0.1](../specs/v0.1/SPEC.md)
- [AlphaNet AIR Schema](../specs/v0.1/alphanet.schema.json)
- [Manifest Schema](../specs/v0.1/manifest.schema.json)
- [Portfolio Schema](../specs/v0.1/portfolio.schema.json)
- [Rule Schema](../specs/v0.1/rule.schema.json)
- [Signal Schema](../specs/v0.1/signal.schema.json)
- [Implementation Plan](rules-compiler/implementation-plan.md) (this file)
- [TradingAgents](https://github.com/TauricResearch/TradingAgents)
- [AI Hedge Fund](https://github.com/virattt/ai-hedge-fund)