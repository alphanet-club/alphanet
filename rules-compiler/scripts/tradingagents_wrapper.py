#!/usr/bin/env python
"""
TradingAgents wrapper for AlphaNet Rules Compiler.

Accepts strategy config via stdin, calls TradingAgents for each symbol,
and emits structured JSON (signals, decisions) that the Go compiler consumes.

The .env file under TRADINGAGENTS_HOME is loaded automatically so API keys
and provider settings come from TradingAgents' own configuration.

Usage:
    echo '{"symbols":["QQQ"],"date":"2024-05-10"}' | python tradingagents_wrapper.py

Environment:
    TRADINGAGENTS_HOME    Path to the cloned TradingAgents repo
"""

import argparse
import json
import os
import subprocess
import sys
import tempfile
from datetime import datetime


def _resolve_ta_path(cli_arg: str | None) -> str | None:
    if cli_arg:
        return cli_arg
    env = os.environ.get("TRADINGAGENTS_HOME")
    if env:
        return os.path.expanduser(env)
    return None


def _find_venv_python(ta_path: str) -> str | None:
    abs_path = os.path.abspath(os.path.expanduser(ta_path))
    for candidate in ["venv/bin/python3", "venv/bin/python", ".venv/bin/python3", ".venv/bin/python"]:
        p = os.path.join(abs_path, candidate)
        if os.path.isfile(p) and os.access(p, os.X_OK):
            return p
    return None


def _load_env_file(ta_path: str, env: dict) -> dict:
    """Load TradingAgents .env into an environment dict."""
    abs_path = os.path.abspath(os.path.expanduser(ta_path))
    env_file = os.path.join(abs_path, ".env")
    if not os.path.isfile(env_file):
        alt = os.path.join(abs_path, ".env.example")
        if os.path.isfile(alt):
            env_file = alt
        else:
            return env
    out = dict(env)
    try:
        with open(env_file) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, _, val = line.partition("=")
                key = key.strip()
                val = val.strip()
                if key not in out:
                    out[key] = val
    except OSError:
        pass
    # Dummy FRED key for macro data import
    if "FRED_API_KEY" not in out:
        out["FRED_API_KEY"] = "dummy"
    return out


def _run_analysis(raw_input: str) -> str:
    """Import TradingAgents, run analysis per symbol, return JSON string."""
    input_data = json.loads(raw_input)
    symbols = input_data.get("symbols", [])
    date = input_data.get("date", datetime.now().strftime("%Y-%m-%d"))
    analysts = tuple(input_data.get("analysts", ["market"]))

    ta_home = os.environ.get("TRADINGAGENTS_HOME")
    if ta_home:
        ab = os.path.abspath(os.path.expanduser(ta_home))
        if ab not in sys.path:
            sys.path.insert(0, ab)

    from tradingagents.default_config import DEFAULT_CONFIG
    from tradingagents.graph.trading_graph import TradingAgentsGraph

    config = DEFAULT_CONFIG.copy()
    # TradingAgents' DEFAULT_CONFIG already applies _apply_env_overrides from
    # TRADINGAGENTS_LLM_PROVIDER, TRADINGAGENTS_DEEP_THINK_LLM, etc.
    # We just set the provider; the user's .env handles model selection.
    provider = os.environ.get("TRADINGAGENTS_LLM_PROVIDER", "openrouter")
    config["llm_provider"] = provider

    # Debug: log the effective config
    deep_think = os.environ.get("TRADINGAGENTS_DEEP_THINK_LLM", "not set")
    quick_think = os.environ.get("TRADINGAGENTS_QUICK_THINK_LLM", "not set")
    backend = os.environ.get("TRADINGAGENTS_LLM_BACKEND_URL", "not set")

    output = {
        "signals": [], "rules": [], "relations": [], "regimes": [],
        "portfolio": None,
        "notes": f"TradingAgents analysis completed for {len(symbols)} symbols on {date}",
    }

    for symbol in symbols:
        try:
            ta = TradingAgentsGraph(debug=False, config=config, selected_analysts=analysts)
            _, decision = ta.propagate(symbol, date)
            decision_str = str(decision)
        except Exception as e:
            decision_str = f"Error: {e}"
        output["signals"].append({
            "action": "add",
            "signal": {
                "signal_id": f"{symbol.lower()}_ta_decision",
                "family": "agents",
                "type": "trading_signal",
                "description": f"TradingAgents decision for {symbol} on {date}: {decision_str}",
                "source": {"name": "TradingAgents", "version": "0.2.5"},
                "symbol": symbol,
                "transform": "level",
                "frequency": "daily",
                "unit": "signal",
            },
            "confidence": 0.7,
            "rationale": decision_str,
        })
        output["notes"] += f"\n- {symbol}: TradingAgents decision = {decision_str}"

    output["notes"] += f"\n[config: provider={provider}, deep_think={deep_think}, quick_think={quick_think}, backend={backend}]"

    return json.dumps(output, indent=2)


def main():
    parser = argparse.ArgumentParser(description="TradingAgents wrapper for AlphaNet")
    parser.add_argument("--output", "-o", help="Write output JSON to this file")
    parser.add_argument("--ta-path", help="Path to the cloned TradingAgents repository")
    parser.add_argument("--input-file", help="Read input JSON from file instead of stdin")
    args = parser.parse_args()

    ta_path = args.ta_path or os.environ.get("TRADINGAGENTS_HOME")

    # 1. Read the raw input (file or stdin)
    if args.input_file:
        with open(args.input_file) as f:
            raw_input = f.read()
    else:
        raw_input = sys.stdin.read()

    if not raw_input.strip():
        print(json.dumps({"signals": [], "rules": [], "relations": [], "regimes": [],
                          "portfolio": None, "notes": "No input data received"}))
        return

    # 2. If a venv is available and we're not already running inside it, delegate
    if ta_path:
        abs_path = os.path.abspath(os.path.expanduser(ta_path))
        venv_python = _find_venv_python(ta_path)

        if venv_python and venv_python != sys.executable:
            ta_env = _load_env_file(abs_path, os.environ)
            ta_env["TRADINGAGENTS_HOME"] = abs_path

            # Write input to temp file to avoid pipe deadlock
            with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as tf:
                tf.write(raw_input)
                tmpfile = tf.name

            try:
                proc = subprocess.run(
                    [venv_python, __file__, "--input-file", tmpfile],
                    capture_output=True, text=True, timeout=20, env=ta_env,
                )
            finally:
                try:
                    os.unlink(tmpfile)
                except OSError:
                    pass

            if proc.returncode != 0:
                print(json.dumps({
                    "signals": [], "rules": [], "relations": [], "regimes": [],
                    "portfolio": None,
                    "notes": f"Error running via venv: {proc.stderr[:500]}",
                }))
            else:
                print(proc.stdout)
            return

        # No venv: inject env vars and add TA to sys.path
        merged = _load_env_file(abs_path, os.environ)
        for k, v in merged.items():
            os.environ.setdefault(k, v)
        if abs_path not in sys.path:
            sys.path.insert(0, abs_path)

    # 3. Run analysis directly
    result = _run_analysis(raw_input)
    print(result)


if __name__ == "__main__":
    main()