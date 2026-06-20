# Stooq Provider

    DoltHub database:

    ```text
    alphanet_stooq
    ```

    Local folder:

    ```text
    backtester/data-providers/stooq/
    ```

    Schema sample:

    ```text
    backtester/data-providers/stooq/schema.sql
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

    ETF, equity, index, and broad commodity benchmark prices.

    ## Source URLs

    - `https://stooq.com/db/h/`
- `https://stooq.com/q/d/l/?s={stooq_symbol}&i=d`

    ## Authentication

    No API key required for simple CSV download path.

    ## Official Scoring Status

    Allowed for official scoring once ingested into the public AlphaNet Dolt database.

    ## Adapter Notes

    Map AlphaNet symbols to Stooq symbols such as SPY -> spy.us, QQQ -> qqq.us, DBC -> dbc.us.

    ## Runtime Role

    The provider adapter should convert this source-specific schema into normalized runtime records used by the backtester.

    Cross-provider normalization happens in Go code, not by joining all data into one shared database.
