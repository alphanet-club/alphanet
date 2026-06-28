# Backtester Data Access Implementation Plan

## Core Rule

Each public data source has its own provider folder:

```text
backtester/data-providers/<source>/
```

Each provider folder contains:

```text
README.md
schema.sql
```

The `schema.sql` is a seed/maintainer schema for initially creating and seeding the public DoltHub database. Normal users usually clone the public database and do not run table creation statements.

---

## Provider Databases

| Source | DoltHub database | Folder |
|---|---|---|
| Alpaca | `alphanet_alpaca` | `backtester/data-providers/alpaca/` |
| FRED | `alphanet_fred` | `backtester/data-providers/fred/` |
| Cboe | `alphanet_cboe` | `backtester/data-providers/cboe/` |
| IMF PortWatch | `alphanet_imf_portwatch` | `backtester/data-providers/imf-portwatch/` |
| Alpha Vantage | `alphanet_alpha_vantage` | `backtester/data-providers/alpha-vantage/` |
| Yahoo Finance | `alphanet_yahoo_finance` | `backtester/data-providers/yahoo-finance/` |

---

## Configurable Public Writes

The backtester must never write to AlphaNet public DoltHub databases unless writes are explicitly enabled.

Enable writes using manifest/backtester config or environment variables.

Default:

```text
public_dolthub_writes.enabled = false
public_dolthub_writes.push_remote = false
```

Environment defaults:

```text
ALPHANET_DOLTHUB_WRITE_ENABLED=false
ALPHANET_DOLTHUB_PUSH_ENABLED=false
```

The official validator may enable writes for approved providers.

Local user runs should default to local-only writes or no writes.

---

## Configurable Local Clone Use

Whether to use a local Dolt clone is configurable per source.

This is required for both local PC and cloud cost optimization.

Each source config should support:

```text
use_local_clone
local_clone_path
remote_mode
```

`remote_mode` values:

```text
fallback
remote_first
disabled
```

---

## Cloud Cost Policy

Cloud workers should decide per source.

Do not hard-code:

```text
always clone
never clone
```

Recommended modes:

```text
auto
remote_first
ephemeral_clone
```

In auto mode:

```text
clone small/reused databases
query remote for large/rarely-used databases
reuse warm clones when available
```

The backtester should record the selected access mode per source in the summary.

---

## Rolling Window

User-facing config stays simple:

```json
{
  "data_loading": {
    "mode": "rolling_window",
    "window_trading_days": 60,
    "memory_limit_mb": 256
  }
}
```

The backtester loads one chunk plus required lookback, evaluates that chunk, releases it, and continues.

---

## Runtime Normalization

Provider schemas are source-specific.

The backtester normalizes provider rows into runtime records in Go code.

Do not cross-join all source data into one shared Dolt database.

---

## Acceptance Criteria

This work is complete when:

1. Provider README and schema are colocated under `backtester/data-providers/<source>/`.
2. Public DoltHub writes are disabled by default.
3. Public DoltHub writes can be enabled by config or environment variables.
4. Local clone use is configurable per source.
5. Cloud access mode is configurable per source.
6. Summary output records source access mode and data commit per source.
