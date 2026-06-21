#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any


def log(message: str) -> None:
    if os.environ.get("ALPHANET_TA_DEBUG"):
        print(f"[tradingagents_adapter] {message}", file=sys.stderr, flush=True)


def load_dotenv(root: Path) -> None:
    env_file = root / ".env"
    if not env_file.exists():
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

    analysts = req.get("analysts") or ["market"]
    if not isinstance(analysts, list):
        raise SystemExit("request.analysts must be a list")
    analysts = [str(a).strip() for a in analysts if str(a).strip()]

    return {"symbols": symbols, "date": date, "analysts": analysts}


def to_signal(symbol: str, date: str, decision: Any) -> dict[str, Any]:
    decision_text = str(decision).strip()
    return {
        "action": "add",
        "signal": {
            "signal_id": f"{symbol.lower()}_tradingagents_decision",
            "family": "agents",
            "type": "trading_signal",
            "description": f"TradingAgents decision for {symbol} on {date}: {decision_text}",
            "source": {"name": "TradingAgents"},
            "symbol": symbol,
            "transform": "level",
            "frequency": "daily",
            "unit": "signal",
        },
        "confidence": 0.7,
        "rationale": decision_text,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="AlphaNet TradingAgents adapter")
    parser.add_argument("--ta-home", default=os.environ.get("TRADINGAGENTS_HOME", ""))
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    raw = sys.stdin.read()
    if not raw.strip():
        raise SystemExit("expected JSON request on stdin")

    req = normalize_request(raw)

    if args.dry_run:
        print(json.dumps({
            "signals": [to_signal(s, req["date"], "DRY_RUN") for s in req["symbols"]],
            "notes": f"dry run for {len(req['symbols'])} symbols",
        }, indent=2))
        return 0

    if not args.ta_home:
        raise SystemExit("TRADINGAGENTS_HOME or --ta-home is required")

    ta_home = Path(args.ta_home).expanduser().resolve()
    if not ta_home.exists():
        raise SystemExit(f"TradingAgents home not found: {ta_home}")

    load_dotenv(ta_home)
    os.environ["TRADINGAGENTS_HOME"] = str(ta_home)
    os.environ.setdefault("FRED_API_KEY", "dummy")
    sys.path.insert(0, str(ta_home))

    log(f"ta_home={ta_home}")
    log(f"symbols={req['symbols']} date={req['date']} analysts={req['analysts']}")

    try:
        from tradingagents.default_config import DEFAULT_CONFIG
        from tradingagents.graph.trading_graph import TradingAgentsGraph
    except Exception as exc:
        raise SystemExit(f"failed to import TradingAgents from {ta_home}: {type(exc).__name__}: {exc}") from exc

    config = DEFAULT_CONFIG.copy()
    if os.environ.get("TRADINGAGENTS_LLM_PROVIDER"):
        config["llm_provider"] = os.environ["TRADINGAGENTS_LLM_PROVIDER"]

    signals: list[dict[str, Any]] = []
    notes: list[str] = []

    for symbol in req["symbols"]:
        log(f"start symbol={symbol}")
        try:
            graph = TradingAgentsGraph(
                debug=bool(os.environ.get("ALPHANET_TA_GRAPH_DEBUG")),
                config=config,
                selected_analysts=tuple(req["analysts"]),
            )
            _, decision = graph.propagate(symbol, req["date"])
            log(f"done symbol={symbol}")
        except Exception as exc:
            decision = f"ERROR {type(exc).__name__}: {exc}"
            log(f"error symbol={symbol}: {decision}")

        signals.append(to_signal(symbol, req["date"], decision))
        notes.append(f"{symbol}: {decision}")

    print(json.dumps({
        "signals": signals,
        "notes": "TradingAgents adapter completed\n" + "\n".join(notes),
    }, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
