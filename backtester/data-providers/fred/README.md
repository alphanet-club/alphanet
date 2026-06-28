# FRED Provider

DoltHub database:

```text
alphanet_fred
```

Local folder:

```text
backtester/data-providers/fred/
```

Schema sample:

```text
backtester/data-providers/fred/schema.sql
```

## Important Usage Model

Most users should not create these tables themselves.

In normal local use, users should clone the public AlphaNet DoltHub database
for this provider and point the backtester at that clone.

Typical user workflow:

```bash
# From the root:
dolt clone alphanet-club/alphanet_fred ./data/dolthub/alphanet_fred
```

## Purpose

Macro, rates, WTI, broad dollar, and fallback volatility representation.

## Source URLs

- `https://fred.stlouisfed.org/docs/api/fred/`
- `https://fred.stlouisfed.org/docs/api/fred/series_observations.html`
- `https://api.stlouisfed.org/fred/series/observations?series_id={series_id}&api_key={api_key}&file_type=json`

## Authentication

Requires a FRED API key:

```bash
export FRED_API_KEY="..."
```

## Importer

Install provider importer dependencies from the repository root:

```bash
cd backtester/data-providers
poetry install
```

View importer options:

```bash
cd backtester/data-providers
poetry run python fred/importer.py --help
```

Seed a local Dolt database:

```bash
cd backtester/data-providers
poetry run python fred/importer.py \
  --db ../../data/dolthub/alphanet_fred \
  --remote alphanet-club/alphanet_fred \
  --init-schema \
  --start 2026-01-01 \
  --end 2026-01-31
```

Commit and push to DoltHub explicitly:

```bash
cd backtester/data-providers
poetry run python fred/importer.py \
  --db ../../data/dolthub/alphanet_fred \
  --remote alphanet-club/alphanet_fred \
  --init-schema \
  --start 2026-01-01 \
  --end 2026-01-31 \
  --commit \
  --push \
  --branch main
```

By default the importer seeds `DGS10`, `DGS2`, `DCOILWTICO`, `DTWEXBGS`,
and `VIXCLS`.

Add or override series with repeatable `--series` flags:

```bash
--series DGS10=US10Y,"10-Year Treasury Constant Maturity Rate",rates
--series FEDFUNDS=FEDFUNDS,"Effective Federal Funds Rate",rates
```

The script masks API keys in `ingestion_runs.request_url` and writes only to
the local Dolt clone unless `--commit` and `--push` are explicitly provided.

## Official Scoring Status

Allowed for official scoring once ingested into the public AlphaNet Dolt database.

## Runtime Role

The provider adapter should convert this source-specific schema into normalized
runtime records used by the backtester.

Cross-provider normalization happens in Go code, not by joining all data into
one shared database.
