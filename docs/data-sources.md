# AlphaNet Data Sources and DoltHub Workflow

## Purpose

AlphaNet uses public data for deterministic backtesting.

This data is required by the backtester, not by the rules compiler.

The rules compiler may use user-provided rules, strategy text, research notes, LLMs, external investment models, or external agent engines. Those systems may use their own data access mechanisms. The compiled AlphaNet IR should not depend on those sources at backtest time.

---

## Core v1 Data Source Rule

Each public data source has:

```text
1. Its own DoltHub database.
2. Its own schema.
3. Its own connection.
4. Its own provider folder under backtester/data-providers/<source>/.
```

Do not combine public source data into one universal Dolt database.

Do not keep schemas disconnected from provider docs.

Each provider folder contains:

```text
README.md
schema.sql
```

The schema is a seed/maintainer schema for creating the public database. Most users should clone the public DoltHub database and use it directly.

---

## Provider Databases

```text
backtester/data-providers/alpaca/          -> alphanet_alpaca
backtester/data-providers/fred/            -> alphanet_fred
backtester/data-providers/cboe/            -> alphanet_cboe
backtester/data-providers/imf-portwatch/   -> alphanet_imf_portwatch
backtester/data-providers/alpha-vantage/   -> alphanet_alpha_vantage
backtester/data-providers/yahoo-finance/   -> alphanet_yahoo_finance
```

Yahoo Finance is optional and not required for official scoring in v1.

---

## Source Priority

The backtester should resolve data in this priority order:

```text
1. Local Dolt clone for that specific source
2. AlphaNet public DoltHub database for that specific source
3. Public API or public downloadable source through the provider importer
4. Local CSV files, when explicitly configured as fixtures or overrides
```

The backtester should query local provider Dolt clones first, detect gaps, use
remote Dolt SQL for missing slices, and call provider importers only when data
is still unavailable.

---

## Public DoltHub Writes

Writes to AlphaNet public DoltHub databases must be explicitly enabled.

They must never happen silently.

Writes can be enabled by either:

```text
1. manifest/backtester config
2. backtester environment variables
```

Recommended config:

```json
{
  "public_dolthub_writes": {
    "enabled": false,
    "allow_sources": [],
    "require_clean_working_set": true,
    "require_auth": true,
    "commit_message_template": "Update {source} data from {provider}",
    "push_remote": false
  }
}
```

Recommended environment variables:

```text
ALPHANET_DOLTHUB_WRITE_ENABLED=false
ALPHANET_DOLTHUB_PUSH_ENABLED=false
ALPHANET_DOLTHUB_ALLOWED_SOURCES=alpaca,fred,cboe
```

Default behavior:

```text
read public data
write local clone if configured
do not push to public DoltHub
```

Official validator behavior may enable public writes explicitly.

---

## Per-Source Local Clone Policy

Whether to use a local Dolt clone must be configurable per source.

This matters for both local and cloud runs.

Recommended config shape:

```json
{
    "sources": {
    "alpaca": {
      "database": "alphanet_alpaca",
      "use_local_clone": true,
      "local_clone_path": "./data/dolthub/alphanet_alpaca",
      "remote_mode": "fallback"
    },
    "fred": {
      "database": "alphanet_fred",
      "use_local_clone": true,
      "local_clone_path": "./data/dolthub/alphanet_fred",
      "remote_mode": "fallback"
    },
    "imf_portwatch": {
      "database": "alphanet_imf_portwatch",
      "use_local_clone": false,
      "remote_mode": "remote_first"
    }
  }
}
```

Supported `remote_mode` values:

```text
fallback       use local clone first, remote if missing
remote_first   query remote before local clone
disabled       never use remote
```

---

## Cloud Data Access Policy

Do not make a universal rule that cloud workers never clone.

Do not make a universal rule that cloud workers always clone.

Hosted cloud mode should be configurable per source and can default to:

```text
auto
```

In `auto` mode, the runner decides source-by-source.

Clone locally when:

```text
- database is small
- many range queries will hit that source
- repeated chunks need the same provider
- official scoring requires exact committed data locally
```

Use remote/public access when:

```text
- one or two series are needed
- the source database is large
- the run is short
- startup time or storage cost dominates
```

Backtest summaries should record the selected source mode.

---

## Local Clone Workflow

```bash
mkdir -p data/dolthub
cd data/dolthub

dolt clone alphanet-club/alphanet_alpaca
dolt clone alphanet-club/alphanet_fred
dolt clone alphanet-club/alphanet_cboe
dolt clone alphanet-club/alphanet_imf_portwatch
dolt clone alphanet-club/alphanet_alpha_vantage
```

Optional:

```bash
dolt clone alphanet-club/alphanet_yahoo_finance
```

---

## User-Facing Data Loading Settings

Users should only need these memory/performance knobs:

```json
{
  "data_loading": {
    "mode": "rolling_window",
    "window_trading_days_target": 60,
    "max_memory_mb": 256,
    "max_storage_mb": 2048,
    "source_storage_policy": "weighted_by_requirement",
    "prune_after_window": false
  }
}
```

`window_trading_days_target` is a target. The backtester may shrink or expand
actual windows based on inferred lookback, memory limits, and storage limits.

The backtester handles lookback inference, adaptive chunk planning, gap
detection, range queries, storage budgeting, and memory checks internally.

The default `weighted_by_requirement` storage policy uses:

```text
provider_weight = required_rows * average_row_width * reuse_factor
provider_budget = max_storage_mb * provider_weight / sum(provider_weights)
```

Remote Dolt SQL result sets count against `max_memory_mb` while the current
rolling window is being evaluated. That memory should be released after each
window unless the data was imported into a local provider Dolt clone.

Provider-specific local Dolt clones remain the cache. The backtester should not
normalize all provider data into one shared cache database.

When `prune_after_window` is enabled, completed window data must be removed in
a way that actually reclaims enough disk space to keep provider clones within
their configured storage budgets. The backtester should verify measured
on-disk usage after pruning.

---

## Public Database Rebalancing

AlphaNet should periodically rebalance preloaded public provider Dolt data based
on public usage and official strategy needs.

Commonly used symbols, series, date ranges, and example-strategy requirements
should be prioritized to keep routine backtests affordable.

Less common data can remain available through remote SQL or provider importer
fallback.

---

## Data Versioning in Backtest Results

Backtest summaries should record source database commits independently.

Example:

```json
{
  "data_commits": [
    {
      "source": "alpaca",
      "database": "alphanet_alpaca",
      "commit": "abc123",
      "access_mode": "local_clone"
    },
    {
      "source": "fred",
      "database": "alphanet_fred",
      "commit": "def456",
      "access_mode": "remote_first"
    }
  ]
}
```
