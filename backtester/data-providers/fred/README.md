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

    Macro, rates, WTI, broad dollar, fallback volatility, and XAUUSD representation.

    ## Source URLs

    - `https://fred.stlouisfed.org/docs/api/fred/`
- `https://fred.stlouisfed.org/docs/api/fred/series_observations.html`
- `https://api.stlouisfed.org/fred/series/observations?series_id={series_id}&api_key={api_key}&file_type=json`

    ## Authentication

    Requires FRED_API_KEY for API ingestion.

    ## Official Scoring Status

    Allowed for official scoring once ingested into the public AlphaNet Dolt database.

    ## Adapter Notes

    AlphaNet XAUUSD maps to FRED GOLDPMGBD228NLBM in v1.

    ## Runtime Role

    The provider adapter should convert this source-specific schema into normalized runtime records used by the backtester.

    Cross-provider normalization happens in Go code, not by joining all data into one shared database.
