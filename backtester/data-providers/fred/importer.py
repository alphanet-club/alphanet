#!/usr/bin/env python3
"""Import FRED observations into the alphanet_fred Dolt database."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import sys
import urllib.parse
import urllib.request
import urllib.error
import uuid
from pathlib import Path
from typing import Any

sys.path.append(str(Path(__file__).resolve().parents[1] / "common"))
from dolt_loader import commit_and_maybe_push, ensure_database, upsert_rows, utc_now  # noqa: E402


DEFAULT_SERIES = {
    "DGS10": {
        "title": "Market Yield on U.S. Treasury Securities at 10-Year Constant Maturity",
        "category": "rates",
        "alpha_symbol": "US10Y",
    },
    "DGS2": {
        "title": "Market Yield on U.S. Treasury Securities at 2-Year Constant Maturity",
        "category": "rates",
        "alpha_symbol": "US2Y",
    },
    "DCOILWTICO": {
        "title": "Crude Oil Prices: West Texas Intermediate (WTI)",
        "category": "commodities",
        "alpha_symbol": "WTI",
    },
    "DTWEXBGS": {
        "title": "Nominal Broad U.S. Dollar Index",
        "category": "fx",
        "alpha_symbol": "USD_BROAD",
    },
    "VIXCLS": {
        "title": "CBOE Volatility Index: VIX",
        "category": "volatility",
        "alpha_symbol": "VIX",
    },
}


def main() -> int:
    args = parse_args()
    api_key = args.api_key or os.environ.get("FRED_API_KEY")
    if not api_key:
        raise SystemExit("FRED_API_KEY is required. Pass --api-key or set the environment variable.")

    db_path = args.db.resolve()
    ingestion_id = args.ingestion_id or f"fred-{dt.datetime.now(dt.UTC).strftime('%Y%m%dT%H%M%SZ')}-{uuid.uuid4().hex[:8]}"
    started_at = utc_now()

    ensure_database(
        db_path,
        Path(__file__).with_name("schema.sql"),
        remote=args.remote,
        init_schema=args.init_schema,
    )

    series = parse_series(args.series)
    rows_read = 0
    rows_written = 0
    request_urls: list[str] = []
    status = "success"
    error_message = None

    try:
        observations_by_series: dict[str, list[dict[str, Any]]] = {}
        for series_id in sorted(series):
            request_url = build_observations_url(series_id, api_key, args.start, args.end)
            request_urls.append(mask_api_key(request_url))
            observations = fetch_observations(series_id, request_url, ingestion_id)
            observations_by_series[series_id] = observations
            rows_read += len(observations)

        upsert_source_metadata(db_path)
        upsert_rows(
            db_path,
            "series_catalog",
            [
                "series_id",
                "title",
                "category",
                "frequency",
                "units",
                "seasonal_adjustment",
                "source_url",
                "alpha_symbol",
                "notes",
            ],
            series_catalog_rows(series),
            update_columns=[
                "title",
                "category",
                "frequency",
                "units",
                "seasonal_adjustment",
                "source_url",
                "alpha_symbol",
                "notes",
            ],
        )
        upsert_rows(
            db_path,
            "series_aliases",
            ["alias", "series_id", "meaning"],
            alias_rows(series),
            update_columns=["series_id", "meaning"],
        )

        for observations in observations_by_series.values():
            rows_written += upsert_rows(
                db_path,
                "observations",
                [
                    "series_id",
                    "date",
                    "value",
                    "realtime_start",
                    "realtime_end",
                    "source_id",
                    "ingestion_id",
                ],
                observations,
                update_columns=["value", "source_id", "ingestion_id"],
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
                "request_url",
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
                    "request_url": "\n".join(request_urls),
                    "request_params": {"observation_start": args.start, "observation_end": args.end},
                    "rows_read": rows_read,
                    "rows_written": rows_written,
                    "error_message": error_message,
                    "metadata": {"series": sorted(series)},
                }
            ],
            update_columns=[
                "finished_at",
                "status",
                "request_url",
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
            message=args.commit_message or f"Update FRED data through {args.end}",
            push=args.push,
            branch=args.branch,
        )

    print(f"Imported {rows_written} FRED rows into {db_path}")
    return 0


def upsert_source_metadata(db_path: Path) -> None:
    upsert_rows(
        db_path,
        "source_metadata",
        ["source_id", "provider_name", "api_docs_url", "terms_url", "notes"],
        [
            {
                "source_id": "fred",
                "provider_name": "Federal Reserve Economic Data",
                "api_docs_url": "https://fred.stlouisfed.org/docs/api/fred/",
                "terms_url": "https://fred.stlouisfed.org/docs/api/terms_of_use.html",
                "notes": "Imported by AlphaNet FRED importer.",
            }
        ],
        update_columns=["provider_name", "api_docs_url", "terms_url", "notes"],
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", type=Path, required=True, help="Path to local Dolt database.")
    parser.add_argument("--start", required=True, help="Start date, YYYY-MM-DD.")
    parser.add_argument("--end", required=True, help="End date, YYYY-MM-DD.")
    parser.add_argument(
        "--series",
        action="append",
        default=[],
        help="FRED series id, optionally SERIES=alpha_symbol,title,category. Repeatable. Defaults to the AlphaNet seed set.",
    )
    parser.add_argument("--api-key", default=None, help="FRED API key. Defaults to FRED_API_KEY.")
    parser.add_argument("--remote", default=None, help="Optional origin remote, for example alphanet-club/alphanet_fred.")
    parser.add_argument("--init-schema", action="store_true", help="Apply schema.sql before importing.")
    parser.add_argument("--ingestion-id", default=None, help="Override generated ingestion id.")
    parser.add_argument("--commit", action="store_true", help="Create a Dolt commit after import if there are changes.")
    parser.add_argument("--commit-message", default=None, help="Commit message to use with --commit or --push.")
    parser.add_argument("--push", action="store_true", help="Push committed changes to origin.")
    parser.add_argument("--branch", default=None, help="Branch to push. Defaults to current Dolt branch.")
    return parser.parse_args()


def parse_series(values: list[str]) -> dict[str, dict[str, str | None]]:
    if not values:
        return DEFAULT_SERIES

    parsed: dict[str, dict[str, str | None]] = {}
    for value in values:
        if "=" in value:
            series_id, payload = value.split("=", 1)
            parts = [part.strip() for part in payload.split(",")]
            alpha_symbol = parts[0] if parts and parts[0] else None
            title = parts[1] if len(parts) > 1 and parts[1] else series_id
            category = parts[2] if len(parts) > 2 and parts[2] else None
        else:
            series_id = value
            alpha_symbol = None
            title = value
            category = None
        parsed[series_id.upper()] = {"title": title, "category": category, "alpha_symbol": alpha_symbol}
    return parsed


def series_catalog_rows(series: dict[str, dict[str, str | None]]) -> list[dict[str, Any]]:
    rows = []
    for series_id, metadata in series.items():
        rows.append(
            {
                "series_id": series_id,
                "title": metadata["title"],
                "category": metadata.get("category"),
                "frequency": None,
                "units": None,
                "seasonal_adjustment": None,
                "source_url": f"https://fred.stlouisfed.org/series/{series_id}",
                "alpha_symbol": metadata.get("alpha_symbol"),
                "notes": None,
            }
        )
    return rows


def alias_rows(series: dict[str, dict[str, str | None]]) -> list[dict[str, Any]]:
    rows = []
    for series_id, metadata in series.items():
        alpha_symbol = metadata.get("alpha_symbol")
        if alpha_symbol:
            rows.append({"alias": alpha_symbol, "series_id": series_id, "meaning": metadata["title"]})
    return rows


def build_observations_url(series_id: str, api_key: str, start: str, end: str) -> str:
    params = {
        "series_id": series_id,
        "api_key": api_key,
        "file_type": "json",
        "observation_start": start,
        "observation_end": end,
    }
    return "https://api.stlouisfed.org/fred/series/observations?" + urllib.parse.urlencode(params)


def fetch_observations(series_id: str, request_url: str, ingestion_id: str) -> list[dict[str, Any]]:
    try:
        with urllib.request.urlopen(request_url, timeout=30) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"FRED API HTTP {exc.code} for {series_id}: {extract_error_message(body)}") from exc

    if "error_code" in payload:
        raise RuntimeError(f"FRED API error {payload['error_code']}: {payload.get('error_message')}")

    rows = []
    for observation in payload.get("observations", []):
        rows.append(
            {
                "series_id": series_id,
                "date": observation["date"],
                "value": decimal_or_none(observation.get("value")),
                "realtime_start": observation.get("realtime_start") or "1776-07-04",
                "realtime_end": observation.get("realtime_end") or "9999-12-31",
                "source_id": "fred",
                "ingestion_id": ingestion_id,
            }
        )
    return rows


def extract_error_message(body: str) -> str:
    try:
        payload = json.loads(body)
    except json.JSONDecodeError:
        return body.strip()[:500] or "empty response body"
    if isinstance(payload, dict):
        return str(payload.get("error_message") or payload.get("message") or payload)
    return str(payload)


def decimal_or_none(value: str | None) -> str | None:
    if value is None:
        return None
    value = value.strip()
    if not value or value == "." or value.lower() in {"null", "nan", "none"}:
        return None
    return value


def mask_api_key(url: str) -> str:
    parsed = urllib.parse.urlparse(url)
    params = urllib.parse.parse_qsl(parsed.query, keep_blank_values=True)
    masked = [(key, "***" if key == "api_key" else value) for key, value in params]
    return urllib.parse.urlunparse(parsed._replace(query=urllib.parse.urlencode(masked)))


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as exc:
        print(f"error: {exc}", file=sys.stderr)
        raise SystemExit(1)
