# Yahoo Finance Provider

    DoltHub database:

    ```text
    alphanet_yahoo_finance
    ```

    Local folder:

    ```text
    backtester/data-providers/yahoo-finance/
    ```

    Schema sample:

    ```text
    backtester/data-providers/yahoo-finance/schema.sql
    ```

    ## Important Usage Model

Most users should not create these tables themselves.

In normal local use, users should clone the public AlphaNet DoltHub database for this provider and point the backtester at that clone.

The `schema.sql` file in this folder is primarily for:

- AlphaNet maintainers creating the initial public DoltHub database
- seeding a new public provider database
- integration tests
- local development of ingestion adapters
- documenting the expected source-specific table shape

Typical user workflow:

```bash
dolt clone alphanet-club/<provider_database>
```

Then configure the backtester to use that local clone or choose remote/public access for that source.


    ## Purpose

    Optional research fallback only.

    ## Source URLs

    - No stable official URL documented for v1. Optional fallback only.

    ## Authentication

    No official AlphaNet credential requirement in v1.

    ## Official Scoring Status

    Not required for official scoring in v1.

    ## Adapter Notes

    Disabled by default. Use only when explicitly enabled by user.

    ## Runtime Role

    The provider adapter should convert this source-specific schema into normalized runtime records used by the backtester.

    Cross-provider normalization happens in Go code, not by joining all data into one shared database.
