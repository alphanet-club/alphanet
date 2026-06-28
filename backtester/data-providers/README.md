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

## Python Importer Dependencies

Provider importer scripts use Poetry for Python dependencies.

From this directory:

```bash
poetry install
poetry run python alpaca/importer.py --help
```

Shared dependency declarations live in:

```text
backtester/data-providers/pyproject.toml
backtester/data-providers/poetry.lock
```

In restricted environments, keep Poetry's cache and virtualenv inside this
folder:

```bash
POETRY_CONFIG_DIR=.poetry-config \
POETRY_CACHE_DIR=.poetry-cache \
POETRY_VIRTUALENVS_IN_PROJECT=true \
poetry install
```
