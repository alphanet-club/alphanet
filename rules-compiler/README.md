# AlphaNet Rules Compiler

The rules compiler converts an authored strategy directory into AlphaNet AIR:

```text
manifest.json + strategy.md + rules.json
        |
        v
rules-compiler
        |
        v
compiled/strategy.ir.json
```

The backtester consumes only the compiled AIR. Agent engines are compile-time helpers only.

## Modes

- `none`: validate/load sources without emitting AIR.
- `manual`: compile the authored manifest/rules deterministically.
- `single`: run the first enabled engine, then compile.
- `ensemble`: run all enabled engines, then compile.

## TradingAgents integration

TradingAgents is invoked through a small adapter:

```text
Go compiler -> TradingAgents venv Python -> scripts/tradingagents_adapter.py -> JSON
```

This avoids recursive wrapper execution and hidden wrapper timeouts.

Test the adapter directly:

```bash
cd rules-compiler

echo '{"symbols":["QQQ"],"date":"2025-12-31","analysts":["market"]}' | TRADINGAGENTS_HOME=~/Github/TradingAgents   ALPHANET_TA_DEBUG=1   ~/Github/TradingAgents/venv/bin/python   scripts/tradingagents_adapter.py --dry-run
```

Then run the real adapter:

```bash
echo '{"symbols":["QQQ"],"date":"2025-12-31","analysts":["market"]}' | TRADINGAGENTS_HOME=~/Github/TradingAgents   ALPHANET_TA_DEBUG=1   ~/Github/TradingAgents/venv/bin/python   scripts/tradingagents_adapter.py
```

The Go engine auto-detects `TRADINGAGENTS_HOME/venv/bin/python` when `TRADINGAGENTS_HOME` is set.

Optional engine config:

```json
{
  "name": "TauricResearch/TradingAgents",
  "version": "0.2.5",
  "symbols": ["QQQ"],
  "config": {
    "ta_path": "~/Github/TradingAgents",
    "python": "~/Github/TradingAgents/venv/bin/python",
    "adapter_script": "scripts/tradingagents_adapter.py",
    "analysts": ["market"],
    "max_symbols": 1
  }
}
```

## Build

```bash
cd rules-compiler
go build -a -o alphanet-compile ./cmd/alphanet-compile
```

## Compile example

```bash
TRADINGAGENTS_HOME=~/Github/TradingAgents ALPHANET_TA_DEBUG=1 ALPHANET_TA_MAX_SYMBOLS=1 ./alphanet-compile ../docs/examples/v0.1/oil-rates-growth-tech --mode single --verbose
```

Do not commit:

- `rules-compiler/alphanet-compile`
- failed/transient generated compiled outputs
- local implementation notes that contradict `rules-compiler-plan.md`
