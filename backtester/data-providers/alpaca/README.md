# Alpaca Provider

DoltHub database:

```text
alphanet_alpaca
```

Local folder:

```text
backtester/data-providers/alpaca/
```

Schema sample:

```text
backtester/data-providers/alpaca/schema.sql
```

## Important Usage Model

Most users should not create these tables themselves.

In normal local use, users should clone the public AlphaNet DoltHub database
for this provider and point the backtester at that clone.

Typical user workflow:

```bash
# From the root:
dolt clone alphanet-club/alphanet_alpaca ./data/dolthub/alphanet_alpaca
```

## Purpose

Equity and ETF daily OHLCV market data.

## Source URLs

- `https://alpaca.markets/data`
- `https://docs.alpaca.markets/us/reference/api-references`
- `https://github.com/alpacahq/alpaca-py`

## Authentication

Requires Alpaca market data credentials:

```bash
export ALPACA_API_KEY_ID="..."
export ALPACA_API_SECRET_KEY="..."
```

The importer uses Alpaca's `iex` feed by default for free-tier compatibility.

## Importer

Install provider importer dependencies from the repository root:

```bash
cd backtester/data-providers
poetry install
```

View importer options:

```bash
cd backtester/data-providers
poetry run python alpaca/importer.py --help
```

Seed a local Dolt database:

```bash
cd backtester/data-providers
poetry run python alpaca/importer.py \
  --db ../../data/dolthub/alphanet_alpaca \
  --remote alphanet-club/alphanet_alpaca \
  --init-schema \
  --start 2026-01-01 \
  --end 2026-01-31
```

Commit and push to DoltHub explicitly:

```bash
cd backtester/data-providers
poetry run python alpaca/importer.py \
  --db ../../data/dolthub/alphanet_alpaca \
  --remote alphanet-club/alphanet_alpaca \
  --init-schema \
  --start 2026-01-01 \
  --end 2026-01-31 \
  --commit \
  --push \
  --branch main
```

By default the importer seeds `SPY`, `QQQ`, `DBC`, `GLD`, and `AMD`.

Add or override symbols with repeatable `--symbol` flags:

```bash
--symbol MSFT=MSFT,"Microsoft Corporation",EQUITY
```

## Official Scoring Status

Candidate market-price provider for official scoring once the data is ingested
into the public AlphaNet Dolt database and licensing/redistribution terms are
confirmed.

## Runtime Role

The provider adapter should convert this source-specific schema into normalized
runtime records used by the backtester.

Cross-provider normalization happens in Go code, not by joining all data into
one shared database.
