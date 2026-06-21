#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Any


CYAN = "\033[36m"
RESET = "\033[0m"


def colorize(message: str) -> str:
    if os.environ.get("NO_COLOR") or os.environ.get("TERM") == "dumb":
        return message
    return f"{CYAN}{message}{RESET}"


def log(message: str) -> None:
    if os.environ.get("ALPHANET_AHF_DEBUG") or os.environ.get("ALPHANET_ENGINE_DEBUG"):
        print(colorize(f"[ai_hedge_fund_adapter] {message}"), file=sys.stderr, flush=True)


def load_dotenv(root: Path) -> None:
    env_file = root / ".env"
    if not env_file.exists():
        log(f"no .env found at {env_file}")
        return
    for raw in env_file.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), value.strip().strip('"').strip("'"))


def normalize_request(raw: str) -> dict[str, Any]:
    try:
        req = json.loads(raw)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"invalid JSON on stdin: {exc}") from exc

    symbols = req.get("symbols") or []
    if not isinstance(symbols, list):
        raise SystemExit("request.symbols must be a list")
    symbols = [str(s).strip().upper() for s in symbols if str(s).strip()]
    symbols = [s for s in symbols if s != "CASH"]

    date = str(req.get("date") or "").strip()
    if not date:
        raise SystemExit("request.date is required")

    return {"symbols": symbols, "date": date}


def infer_decision(text: str) -> str:
    lowered = text.lower()
    for word in ("buy", "sell", "short", "hold", "neutral"):
        if re.search(rf"\b{word}\b", lowered):
            return word.capitalize()
    return ""


def is_actionable_decision(decision: str) -> bool:
    return decision.strip().lower() in {"buy", "sell", "short", "hold", "neutral"}


def to_signal(symbol: str, date: str, decision: str, rationale: str) -> dict[str, Any]:
    return {
        "action": "add",
        "signal": {
            "signal_id": f"{symbol.lower()}_ai_hedge_fund_decision",
            "family": "agents",
            "type": "trading_signal",
            "description": f"AI Hedge Fund decision for {symbol} on {date}: {decision}",
            "source": {"name": "ai-hedge-fund"},
            "symbol": symbol,
            "transform": "level",
            "frequency": "daily",
            "unit": "signal",
        },
        "confidence": 0.65,
        "rationale": rationale[-4000:],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="AlphaNet AI Hedge Fund adapter")
    parser.add_argument("--ahf-home", default=os.environ.get("AI_HEDGE_FUND_HOME", ""))
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--use-poetry", action="store_true", default=bool(os.environ.get("ALPHANET_AHF_USE_POETRY")))
    parser.add_argument("--ollama", action="store_true", default=bool(os.environ.get("ALPHANET_AHF_OLLAMA")))
    args = parser.parse_args()

    raw = sys.stdin.read()
    if not raw.strip():
        raise SystemExit("expected JSON request on stdin")

    req = normalize_request(raw)

    if args.dry_run:
        print(json.dumps({
            "signals": [
                to_signal(s, req["date"], "DRY_RUN", "AI Hedge Fund dry run")
                for s in req["symbols"]
            ],
            "notes": f"dry run for {len(req['symbols'])} symbols",
        }, indent=2))
        return 0

    if not args.ahf_home:
        raise SystemExit("AI_HEDGE_FUND_HOME or --ahf-home is required")

    ahf_home = Path(args.ahf_home).expanduser().resolve()
    if not ahf_home.exists():
        raise SystemExit(f"AI Hedge Fund home not found: {ahf_home}")

    load_dotenv(ahf_home)
    os.environ["AI_HEDGE_FUND_HOME"] = str(ahf_home)

    tickers = ",".join(req["symbols"])
    if args.use_poetry:
        cmd = ["poetry", "run", "python", "src/main.py"]
    else:
        cmd = [sys.executable, "src/main.py"]
    cmd += ["--ticker", tickers, "--end-date", req["date"]]
    if args.ollama:
        cmd.append("--ollama")

    log(f"ahf_home={ahf_home}")
    log(f"cmd={' '.join(cmd)}")

    proc = subprocess.run(cmd, cwd=ahf_home, capture_output=True, text=True, env=os.environ.copy())
    combined = "\n".join(x for x in [proc.stdout, proc.stderr] if x)

    if proc.returncode != 0:
        raise SystemExit(f"AI Hedge Fund exited {proc.returncode}:\n{combined[-4000:]}")

    decision = infer_decision(combined)
    signals = []
    if is_actionable_decision(decision):
        signals = [to_signal(symbol, req["date"], decision, combined) for symbol in req["symbols"]]
    else:
        log("skipped signals: no actionable decision found")

    print(json.dumps({
        "signals": signals,
        "notes": "AI Hedge Fund adapter completed\n"
        + (f"Decision: {decision}\n" if decision else "No actionable decision found; no signal emitted.\n")
        + combined[-4000:],
    }, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
