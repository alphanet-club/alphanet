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

## Data Requirement Planning

Before loading historical data, the backtester should analyze the compiled AIR
and create a `DataRequirementPlan`.

The plan should be derived from:

- backtest start and end dates
- executable rules
- signal interests and signal transforms
- transform lookback windows
- initial allocation
- candidate baskets
- selection policies
- benchmark requirements
- execution price requirements

The plan should answer:

```text
which provider
which table or dataset
which symbols or series
which fields
which date range
which lookback-expanded date range
which transforms require the data
estimated rows and storage
```

The planner must account for signal lookback. If a processing window starts on
`2020-06-01` and a required transform needs 252 trading days of history, the
data requirement starts before `2020-06-01`.

Lookback expansion should be computed per requirement where possible rather
than using one global maximum for every source.

The `DataRequirementPlan` does not introduce a new shared cache database.
Provider-specific Dolt clones remain the local cache and source workspace.

---

## Provider Dolt Clones as Cache

Each provider's local Dolt clone is the cache for that provider.

Do not normalize all provider data into one local cache database.

The resolver should satisfy each planned provider requirement using this order:

```text
1. Query local provider Dolt clone.
2. Detect missing symbols, series, fields, or date ranges.
3. Query the public remote Dolt database for only the missing data.
4. If still incomplete, run the provider importer into the local provider clone.
```

Remote SQL queries should be gap-filling queries, not broad blind fetches.

Importer fallback should request only the missing slice when the provider
supports scoped imports.

Cloud backtests should avoid committing imported rolling-window slices by
default. Local commits and public pushes require explicit write configuration.

---

## Rolling Window

User-facing config stays simple:

```json
{
  "data_loading": {
    "mode": "rolling_window",
    "window_trading_days_target": 60,
    "max_memory_mb": 256,
    "max_storage_mb": 2048,
    "prune_after_window": false
  }
}
```

`window_trading_days_target` is a target, not a guarantee.

The backtester may shrink a window when estimated memory or disk usage would
exceed configured limits. It may expand a window when requirements are small
and doing so is more efficient.

Remote Dolt SQL result sets should count against `max_memory_mb` because they
are loaded into memory unless the resolver explicitly imports or streams them
into a local provider Dolt clone.

Remote result memory should be released after each rolling window is evaluated.
Only data imported into a local provider Dolt clone should persist beyond the
current window.

For each rolling window:

```text
1. Expand the window by required signal lookback.
2. Build provider-specific requirements for that expanded window.
3. Fill local provider Dolt clone gaps from local data, remote SQL, or importer fallback.
4. Evaluate the unexpanded processing window.
5. Release remote query result memory for the completed window.
6. Optionally prune local provider rows that are no longer needed.
```

Pruning should be configurable. When `prune_after_window` is enabled, pruning
must reclaim disk space enough to keep the provider clone within its configured
storage budget. The implementation may use row deletion, ephemeral branches,
temporary clones, Dolt garbage collection, compaction, or clone rotation, but it
must verify actual on-disk usage after pruning. If the storage budget cannot be
met, the planner must shrink future windows, avoid additional imports, or fail
with a clear storage-budget error.

---

## Storage Budgets

Backtester configuration should support a maximum local storage budget.

Example:

```json
{
  "data_loading": {
    "max_storage_mb": 2048,
    "source_storage_policy": "weighted_by_requirement",
    "prune_after_window": true
  }
}
```

The default `weighted_by_requirement` policy should compute provider budgets
using:

```text
provider_weight = required_rows * average_row_width * reuse_factor
provider_budget = max_storage_mb * provider_weight / sum(provider_weights)
```

`required_rows` is the planner's estimated row count for the provider across
the current run or window.

`average_row_width` is the estimated encoded row size for the provider's
required tables and fields.

`reuse_factor` should increase when provider data is reused across many
windows, symbols, transforms, benchmarks, or portfolio operations.

A source with many required symbols, fields, rows, or repeated lookback reuse
should receive a larger share of the budget than a compact macro source.

Local users may configure a large storage budget and build a durable local
cache. Cloud validators may use smaller budgets unless a user pays for a larger
instance or data tier.

---

## Public Database Rebalancing

AlphaNet should periodically rebalance what data is preloaded into public
provider Dolt databases based on usage patterns and strategy needs.

Frequently used symbols, series, date ranges, and official example strategy
requirements should be prioritized so common backtests remain affordable.

Less common or expensive data can remain available through remote queries or
provider importer fallback.

This rebalancing policy should be usage-informed and can change over time as
the community's backtesting demand changes.

---

## Data Plan Summary Output

Backtest summaries should record the data plan and actual resolution behavior.

Example:

```json
{
  "data_plan": {
    "mode": "rolling_window",
    "window_trading_days_target": 60,
    "max_memory_mb": 256,
    "max_storage_mb": 2048,
    "prune_after_window": true,
    "sources": [
      {
        "source": "alpaca",
        "database": "alphanet_alpaca",
        "required_symbols": 21,
        "required_fields": ["open", "high", "low", "close", "adjusted_close", "volume"],
        "access_modes_used": ["local_dolt", "remote_sql", "importer"],
        "rows_from_local": 120000,
        "rows_from_remote": 5000,
        "rows_imported": 700
      }
    ]
  }
}
```

This output is required for reproducibility, debugging, cloud cost analysis, and
future data-usage billing.

---

## Cost-Control Design Checklist

The data access design should explicitly cover these cost and scale risks:

1. Lookback expansion: every rolling window must expand by the lookback required
   for each signal, transform, benchmark, and execution price requirement.
2. Dolt pruning requirement: when `prune_after_window` is enabled, pruning must
   actually reclaim enough disk space to keep provider clones within their
   storage budgets, verified by measured on-disk usage.
3. Ephemeral imports: cloud backtests should avoid committing imported
   rolling-window slices by default; local commits and public pushes require
   explicit write configuration.
4. Gap-first remote access: local provider clones should be checked first, then
   remote SQL should fetch only missing symbols, series, fields, or date ranges.
5. Adaptive windows: `window_trading_days_target` is a target. The planner may
   shrink or expand windows based on lookback, `max_memory_mb`, and
   `max_storage_mb`.
6. Weighted source budgets: the default provider budget formula is
   `provider_weight = required_rows * average_row_width * reuse_factor`, then
   `provider_budget = max_storage_mb * provider_weight / sum(provider_weights)`.
7. Summary observability: summaries should record the data plan, source commits,
   access modes, rows from local, rows from remote, imported rows, memory limits,
   storage limits, and pruning policy.

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
6. Backtester creates a `DataRequirementPlan` from compiled AIR before loading data.
7. Rolling windows expand by required signal lookback.
8. Local provider Dolt clones are used as the provider cache.
9. Remote SQL and importer fallback fill only missing requirement gaps.
10. Memory limits, storage limits, and pruning policy are configurable.
11. Summary output records source access mode, data commit, data plan, and data resolution statistics per source.
