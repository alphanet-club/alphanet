# Data Provider Adapters

Each public data provider has its own folder:

```text
backtester/data-providers/<source>/
```

Each provider folder contains:

```text
README.md
schema.sql
```

The schema is colocated with provider documentation because every source has its own DoltHub database, own schema, and own connection.

Normal users usually clone the public DoltHub database and do not run these schema files.
