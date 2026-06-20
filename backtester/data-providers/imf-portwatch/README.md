# IMF PortWatch Provider

    DoltHub database:

    ```text
    alphanet_imf_portwatch
    ```

    Local folder:

    ```text
    backtester/data-providers/imf-portwatch/
    ```

    Schema sample:

    ```text
    backtester/data-providers/imf-portwatch/schema.sql
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

    Shipping, port, and chokepoint observations.

    ## Source URLs

    - `https://portwatch.imf.org/`
- `https://portwatch.imf.org/pages/data-and-methodology`
- `https://portwatch.imf.org/search?collection=dataset`

    ## Authentication

    No API key expected for public datasets.

    ## Official Scoring Status

    Allowed for official scoring once ingested into the public AlphaNet Dolt database.

    ## Adapter Notes

    Store exact dataset URL in ingestion_runs because endpoint details may change.

    ## Runtime Role

    The provider adapter should convert this source-specific schema into normalized runtime records used by the backtester.

    Cross-provider normalization happens in Go code, not by joining all data into one shared database.
