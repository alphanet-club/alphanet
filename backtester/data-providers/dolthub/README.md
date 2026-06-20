# DoltHub Provider

## Purpose

DoltHub is AlphaNet's public versioned data layer.

The backtester should prefer local Dolt clones for speed and reproducibility.

If local data is missing, the backtester can use AlphaNet public DoltHub databases as the next source before falling back to public APIs.

## Public Databases

```text
alphanet_market_prices
alphanet_macro_fred
alphanet_volatility_cboe
alphanet_shipping_portwatch
alphanet_fx_commodities
```

## Local Clone Workflow

```bash
mkdir -p data/dolthub
cd data/dolthub

dolt clone alphanet-club/alphanet_market_prices
dolt clone alphanet-club/alphanet_macro_fred
dolt clone alphanet-club/alphanet_volatility_cboe
dolt clone alphanet-club/alphanet_shipping_portwatch
dolt clone alphanet-club/alphanet_fx_commodities
```

## Update Workflow

```bash
cd data/dolthub/alphanet_market_prices
dolt pull
```

## Official Scoring

Official scoring should record the Dolt commit hash for every source database used by a backtest.

Public writes require explicit maintainer credentials and should never happen silently.
