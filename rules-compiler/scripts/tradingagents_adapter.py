#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


CYAN = "\033[36m"
RESET = "\033[0m"


def colorize(message: str) -> str:
    if os.environ.get("NO_COLOR") or os.environ.get("TERM") == "dumb":
        return message
    return f"{CYAN}{message}{RESET}"


def log(message: str) -> None:
    if os.environ.get("ALPHANET_TA_DEBUG") or os.environ.get("ALPHANET_ENGINE_DEBUG"):
        print(colorize(f"[tradingagents_adapter] {message}"), file=sys.stderr, flush=True)


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
    req = json.loads(raw)
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


def normalize_decision(decision: Any) -> str:
    text = str(decision).strip()
    return text or "unknown"


def normalized_action(decision: Any) -> str:
    text = normalize_decision(decision).lower()
    aliases = {
        "long": "buy",
        "cover": "buy",
        "exit": "sell",
    }
    return aliases.get(text, text)


def is_actionable_decision(decision: Any) -> bool:
    return normalized_action(decision) in {
        "buy",
        "sell",
        "hold",
        "short",
        "neutral",
        "overweight",
        "underweight",
        "trim",
        "reduce",
        "add",
        "dry_run",
    }


def to_signal(symbol: str, date: str, decision: Any) -> dict[str, Any]:
    decision_text = normalize_decision(decision)
    action = normalized_action(decision)
    return {
        "action": "add",
        "signal": {
            "signal_id": f"{symbol.lower()}_tradingagents_decision",
            "family": "agents",
            "type": "trading_signal",
            "name": f"TradingAgents decision for {symbol}",
            "description": f"TradingAgents decision for {symbol} on {date}: {decision_text}",
            "source": {"name": "TradingAgents"},
            "symbol": symbol,
            "date": date,
            "value": action,
            "transform": "level",
            "frequency": "point_in_time",
            "unit": "decision",
            "confidence": 0.7,
            "rationale": decision_text,
            "recommendation": {
                "action": action,
                "rating": action,
                "confidence": 0.7,
                "rationale": decision_text,
            },
        },
        "confidence": 0.7,
        "rationale": decision_text,
    }


def to_jsonable(value: Any, depth: int = 0) -> Any:
    if depth > 6:
        return repr(value)
    if value is None or isinstance(value, (str, int, float, bool)):
        return value
    if isinstance(value, dict):
        return {str(k): to_jsonable(v, depth + 1) for k, v in value.items()}
    if isinstance(value, (list, tuple, set)):
        return [to_jsonable(v, depth + 1) for v in value]
    if hasattr(value, "model_dump"):
        try:
            return to_jsonable(value.model_dump(), depth + 1)
        except Exception:
            pass
    if hasattr(value, "dict"):
        try:
            return to_jsonable(value.dict(), depth + 1)
        except Exception:
            pass
    if hasattr(value, "content"):
        try:
            return str(value.content)
        except Exception:
            pass
    return repr(value)


def stringify(value: Any, max_chars: int = 50000) -> str:
    if value is None:
        return ""
    if isinstance(value, str):
        text = value
    else:
        try:
            text = json.dumps(to_jsonable(value), indent=2, ensure_ascii=False)
        except Exception:
            text = repr(value)
    if len(text) > max_chars:
        return text[:max_chars] + "\n\n...[truncated]..."
    return text


def title(key: str) -> str:
    return key.replace("_", " ").replace("-", " ").title()


def render_section(key: str, value: Any, level: int = 2) -> list[str]:
    heading = "#" * max(2, min(level, 5)) + " " + title(str(key))
    lines: list[str] = []

    if value is None:
        return lines

    if isinstance(value, str):
        text = value.strip()
        if text:
            lines += [heading, "", text, ""]
        return lines

    if isinstance(value, dict):
        lines += [heading, ""]
        emitted = False
        for k, v in value.items():
            sub = render_section(str(k), v, level + 1)
            if sub:
                lines += sub
                emitted = True
        if not emitted:
            lines += ["```json", stringify(value), "```", ""]
        return lines

    if isinstance(value, list):
        if not value:
            return lines
        lines += [heading, ""]
        for idx, item in enumerate(value):
            if isinstance(item, str):
                if item.strip():
                    lines += [item.strip(), ""]
            elif hasattr(item, "content"):
                content = stringify(getattr(item, "content", ""))
                if content.strip():
                    lines += [f"{'#' * max(3, min(level + 1, 5))} Message {idx + 1}", "", content, ""]
            else:
                lines += ["```json", stringify(item, 12000), "```", ""]
        return lines

    text = stringify(value)
    if text.strip():
        lines += [heading, "", text, ""]
    return lines


def extract_report_markdown(symbol: str, date: str, state: Any, decision: Any) -> str:
    captured_at = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    lines: list[str] = [
        f"# TradingAgents Research Report: {symbol}",
        "",
        f"- Symbol: `{symbol}`",
        f"- Analysis date: `{date}`",
        f"- Captured at: `{captured_at}`",
        f"- Final decision: `{normalize_decision(decision)}`",
        "",
    ]

    if state is None:
        lines += ["## Raw State", "", "_TradingAgents did not return graph state._", ""]
        return "\n".join(lines).strip() + "\n"

    if not isinstance(state, dict):
        lines += ["## Raw State", "", "```text", stringify(state), "```", ""]
        return "\n".join(lines).strip() + "\n"

    preferred = [
        "market_report",
        "sentiment_report",
        "news_report",
        "fundamentals_report",
        "investment_debate_state",
        "investment_plan",
        "trader_investment_plan",
        "risk_debate_state",
        "final_trade_decision",
        "messages",
    ]

    lines += ["## TradingAgents Graph State", ""]
    rendered: set[str] = set()
    for key in preferred:
        if key in state:
            section = render_section(key, state[key], 3)
            if section:
                lines += section
                rendered.add(key)

    leftovers = [
        k for k in state.keys()
        if k not in rendered and not str(k).startswith("_")
        and k not in {"company_of_interest", "trade_date", "sender"}
    ]
    if leftovers:
        lines += ["## Additional State", ""]
        for key in leftovers:
            section = render_section(str(key), state[key], 3)
            if section:
                lines += section

    return "\n".join(lines).strip() + "\n"


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
            "reports": [
                {
                    "engine": "TradingAgents",
                    "symbol": s,
                    "date": req["date"],
                    "format": "markdown",
                    "content": f"# TradingAgents Research Report: {s}\n\nDry run.\n",
                }
                for s in req["symbols"]
            ],
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
    reports: list[dict[str, Any]] = []
    notes: list[str] = []

    for symbol in req["symbols"]:
        log(f"start symbol={symbol}")
        state: Any = None
        try:
            graph = TradingAgentsGraph(
                debug=bool(os.environ.get("ALPHANET_TA_GRAPH_DEBUG")),
                config=config,
                selected_analysts=tuple(req["analysts"]),
            )
            state, decision = graph.propagate(symbol, req["date"])
            log(f"done symbol={symbol}")
        except Exception as exc:
            decision = f"ERROR {type(exc).__name__}: {exc}"
            log(f"error symbol={symbol}: {decision}")

        if is_actionable_decision(decision):
            signals.append(to_signal(symbol, req["date"], decision))
        else:
            log(f"skipped signal symbol={symbol}: non-actionable decision {normalize_decision(decision)!r}")
        reports.append({
            "engine": "TradingAgents",
            "symbol": symbol,
            "date": req["date"],
            "format": "markdown",
            "content": extract_report_markdown(symbol, req["date"], state, decision),
        })
        if is_actionable_decision(decision):
            notes.append(f"{symbol}: {normalize_decision(decision)}")
        else:
            notes.append(f"{symbol}: no signal emitted ({normalize_decision(decision)})")

    print(json.dumps({
        "signals": signals,
        "reports": reports,
        "notes": "TradingAgents adapter completed\n" + "\n".join(notes),
    }, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
