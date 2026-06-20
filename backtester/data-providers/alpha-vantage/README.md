# Alpha Vantage Provider

    DoltHub database:

    ```text
    alphanet_alpha_vantage
    ```

    Local folder:

    ```text
    backtester/data-providers/alpha-vantage/
    ```

    Schema sample:

    ```text
    backtester/data-providers/alpha-vantage/schema.sql
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

    Secondary adjusted-price, FX, and commodity fallback.

    ## Source URLs

    - `https://www.alphavantage.co/documentation/`
- `https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol={symbol}&outputsize=full&apikey={api_key}`

    ## Authentication

    Requires ALPHA_VANTAGE_API_KEY.

    ## Official Scoring Status

    Allowed only when public AlphaNet ingestion has recorded and committed the data. Do not rely on live API calls for official scoring.

    ## Adapter Notes

    Respect rate limits and record function_name and request params in ingestion_runs.

    ## Runtime Role

    The provider adapter should convert this source-specific schema into normalized runtime records used by the backtester.

    Cross-provider normalization happens in Go code, not by joining all data into one shared database.
