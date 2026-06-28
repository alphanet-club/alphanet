#!/usr/bin/env python3
"""Import Alpaca daily stock bars into the alphanet_alpaca Dolt database."""

from __future__ import annotations

import argparse
import datetime as dt
import os
import sys
import uuid
from pathlib import Path
from typing import Any

sys.path.append(str(Path(__file__).resolve().parents[1] / "common"))
from dolt_loader import commit_and_maybe_push, ensure_database, upsert_rows, utc_now  # noqa: E402


DEFAULT_SYMBOLS = {
    "SPY": ("SPY", "SPDR S&P 500 ETF Trust", "ETF"),
    "QQQ": ("QQQ", "Invesco QQQ Trust", "ETF"),
    "DBC": ("DBC", "Invesco DB Commodity Index Tracking Fund", "ETF"),
    "GLD": ("GLD", "SPDR Gold Shares", "ETF"),
    "AMD": ("AMD", "Advanced Micro Devices Inc.", "EQUITY"),
}


def main() -> int:
    args = parse_args()
    api_key = args.api_key or os.environ.get("ALPACA_API_KEY_ID")
    secret_key = args.secret_key or os.environ.get("ALPACA_API_SECRET_KEY")
    if not api_key or not secret_key:
        raise SystemExit(
            "ALPACA_API_KEY_ID and ALPACA_API_SECRET_KEY are required. "
            "Pass --api-key/--secret-key or set the environment variables."
        )

    db_path = args.db.resolve()
    ingestion_id = args.ingestion_id or f"alpaca-{dt.datetime.now(dt.UTC).strftime('%Y%m%dT%H%M%SZ')}-{uuid.uuid4().hex[:8]}"
    started_at = utc_now()

    ensure_database(
        db_path,
        Path(__file__).with_name("schema.sql"),
        remote=args.remote,
        init_schema=args.init_schema,
    )

    symbols = parse_symbols(args.symbol)
    upsert_source_metadata(db_path)
    upsert_rows(
        db_path,
        "symbols",
        ["symbol", "alpaca_symbol", "name", "instrument_type", "exchange", "currency", "active", "notes"],
        symbol_rows(symbols),
        update_columns=["alpaca_symbol", "name", "instrument_type", "exchange", "currency", "active", "notes"],
    )

    rows_read = 0
    rows_written = 0
    status = "success"
    error_message = None

    try:
        bars = fetch_daily_bars(
            api_key,
            secret_key,
            [metadata["alpaca_symbol"] for metadata in symbols.values()],
            args.start,
            args.end,
            feed=args.feed,
            adjustment=args.adjustment,
        )
        rows = bar_rows(symbols, bars, ingestion_id, args.feed, args.adjustment)
        rows_read = len(rows)
        rows_written = upsert_rows(
            db_path,
            "daily_prices",
            [
                "symbol",
                "date",
                "open",
                "high",
                "low",
                "close",
                "adjusted_close",
                "volume",
                "trade_count",
                "vwap",
                "feed",
                "timeframe",
                "adjustment",
                "source_id",
                "ingestion_id",
            ],
            rows,
            update_columns=[
                "open",
                "high",
                "low",
                "close",
                "adjusted_close",
                "volume",
                "trade_count",
                "vwap",
                "timeframe",
                "source_id",
                "ingestion_id",
            ],
        )
    except Exception as exc:
        status = "failed"
        error_message = str(exc)
        raise
    finally:
        upsert_rows(
            db_path,
            "ingestion_runs",
            [
                "ingestion_id",
                "started_at",
                "finished_at",
                "status",
                "request_params",
                "rows_read",
                "rows_written",
                "error_message",
                "metadata",
            ],
            [
                {
                    "ingestion_id": ingestion_id,
                    "started_at": started_at,
                    "finished_at": utc_now(),
                    "status": status,
                    "request_params": {
                        "symbols": sorted(symbols.keys()),
                        "start": args.start,
                        "end": args.end,
                        "feed": args.feed,
                        "timeframe": "1Day",
                        "adjustment": args.adjustment,
                    },
                    "rows_read": rows_read,
                    "rows_written": rows_written,
                    "error_message": error_message,
                    "metadata": {"provider": "alpaca-py"},
                }
            ],
            update_columns=[
                "finished_at",
                "status",
                "request_params",
                "rows_read",
                "rows_written",
                "error_message",
                "metadata",
            ],
        )

    if args.commit or args.push:
        commit_and_maybe_push(
            db_path,
            message=args.commit_message or f"Update Alpaca data through {args.end}",
            push=args.push,
            branch=args.branch,
        )

    print(f"Imported {rows_written} Alpaca rows into {db_path}")
    return 0


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", type=Path, required=True, help="Path to local Dolt database.")
    parser.add_argument("--start", required=True, help="Start date, YYYY-MM-DD.")
    parser.add_argument("--end", required=True, help="End date, YYYY-MM-DD.")
    parser.add_argument(
        "--symbol",
        action="append",
        default=[],
        help="AlphaNet symbol, optionally ALPHA=alpaca_symbol,name,type. Repeatable. Defaults to SPY, QQQ, DBC, GLD, AMD.",
    )
    parser.add_argument("--feed", default="iex", choices=["iex", "sip", "otc"], help="Alpaca market data feed. Free plans generally use iex.")
    parser.add_argument("--adjustment", default="raw", choices=["raw", "split", "dividend", "all"], help="Corporate-action adjustment.")
    parser.add_argument("--api-key", default=None, help="Alpaca API key id. Defaults to ALPACA_API_KEY_ID.")
    parser.add_argument("--secret-key", default=None, help="Alpaca secret key. Defaults to ALPACA_API_SECRET_KEY.")
    parser.add_argument("--remote", default=None, help="Optional origin remote, for example alphanet-club/alphanet_alpaca.")
    parser.add_argument("--init-schema", action="store_true", help="Apply schema.sql before importing.")
    parser.add_argument("--ingestion-id", default=None, help="Override generated ingestion id.")
    parser.add_argument("--commit", action="store_true", help="Create a Dolt commit after import if there are changes.")
    parser.add_argument("--commit-message", default=None, help="Commit message to use with --commit or --push.")
    parser.add_argument("--push", action="store_true", help="Push committed changes to origin.")
    parser.add_argument("--branch", default=None, help="Branch to push. Defaults to current Dolt branch.")
    return parser.parse_args()


def upsert_source_metadata(db_path: Path) -> None:
    upsert_rows(
        db_path,
        "source_metadata",
        ["source_id", "provider_name", "api_docs_url", "terms_url", "notes"],
        [
            {
                "source_id": "alpaca",
                "provider_name": "Alpaca Market Data",
                "api_docs_url": "https://docs.alpaca.markets/reference/stockbars",
                "terms_url": "https://alpaca.markets/terms-conditions",
                "notes": "Imported by AlphaNet Alpaca importer using alpaca-py.",
            }
        ],
        update_columns=["provider_name", "api_docs_url", "terms_url", "notes"],
    )


def parse_symbols(values: list[str]) -> dict[str, dict[str, str]]:
    if not values:
        values = [f"{symbol}={alpaca_symbol},{name},{instrument_type}" for symbol, (alpaca_symbol, name, instrument_type) in DEFAULT_SYMBOLS.items()]

    parsed: dict[str, dict[str, str]] = {}
    for value in values:
        if "=" in value:
            alpha_symbol, payload = value.split("=", 1)
            parts = [part.strip() for part in payload.split(",")]
            alpaca_symbol = parts[0]
            name = parts[1] if len(parts) > 1 and parts[1] else alpha_symbol.upper()
            instrument_type = parts[2] if len(parts) > 2 and parts[2] else "EQUITY"
        else:
            alpha_symbol = value
            alpaca_symbol = value.upper()
            name = value.upper()
            instrument_type = "EQUITY"
        parsed[alpha_symbol.upper()] = {
            "alpaca_symbol": alpaca_symbol.upper(),
            "name": name,
            "instrument_type": instrument_type.upper(),
        }
    return parsed


def symbol_rows(symbols: dict[str, dict[str, str]]) -> list[dict[str, Any]]:
    rows = []
    for symbol, metadata in symbols.items():
        rows.append(
            {
                "symbol": symbol,
                "alpaca_symbol": metadata["alpaca_symbol"],
                "name": metadata["name"],
                "instrument_type": metadata["instrument_type"],
                "exchange": None,
                "currency": "USD",
                "active": True,
                "notes": None,
            }
        )
    return rows


def fetch_daily_bars(
    api_key: str,
    secret_key: str,
    symbols: list[str],
    start: str,
    end: str,
    *,
    feed: str,
    adjustment: str,
) -> dict[str, list[Any]]:
    try:
        from alpaca.data.historical import StockHistoricalDataClient
        from alpaca.data.requests import StockBarsRequest
        from alpaca.data.timeframe import TimeFrame
        from alpaca.data.enums import Adjustment, DataFeed
    except ImportError as exc:
        raise RuntimeError("alpaca-py is required. Install it with: python3 -m pip install alpaca-py") from exc

    client = StockHistoricalDataClient(api_key, secret_key)
    request = StockBarsRequest(
        symbol_or_symbols=symbols,
        timeframe=TimeFrame.Day,
        start=parse_date(start),
        end=parse_date(end) + dt.timedelta(days=1),
        feed=DataFeed(feed),
        adjustment=Adjustment(adjustment),
    )
    response = client.get_stock_bars(request)
    return response.data


def bar_rows(
    symbols: dict[str, dict[str, str]],
    bars_by_alpaca_symbol: dict[str, list[Any]],
    ingestion_id: str,
    feed: str,
    adjustment: str,
) -> list[dict[str, Any]]:
    alpaca_to_alpha = {metadata["alpaca_symbol"]: symbol for symbol, metadata in symbols.items()}
    rows = []
    for alpaca_symbol, bars in bars_by_alpaca_symbol.items():
        alpha_symbol = alpaca_to_alpha.get(alpaca_symbol.upper())
        if not alpha_symbol:
            continue
        for bar in bars:
            date = bar.timestamp.date().isoformat()
            close = decimal_or_none(getattr(bar, "close", None))
            rows.append(
                {
                    "symbol": alpha_symbol,
                    "date": date,
                    "open": decimal_or_none(getattr(bar, "open", None)),
                    "high": decimal_or_none(getattr(bar, "high", None)),
                    "low": decimal_or_none(getattr(bar, "low", None)),
                    "close": close,
                    "adjusted_close": close,
                    "volume": decimal_or_none(getattr(bar, "volume", None)),
                    "trade_count": getattr(bar, "trade_count", None),
                    "vwap": decimal_or_none(getattr(bar, "vwap", None)),
                    "feed": feed,
                    "timeframe": "1Day",
                    "adjustment": adjustment,
                    "source_id": "alpaca",
                    "ingestion_id": ingestion_id,
                }
            )
    return rows


def parse_date(value: str) -> dt.datetime:
    return dt.datetime.strptime(value, "%Y-%m-%d").replace(tzinfo=dt.UTC)


def decimal_or_none(value: Any) -> str | None:
    if value is None:
        return None
    return str(value)


if __name__ == "__main__":
    raise SystemExit(main())
