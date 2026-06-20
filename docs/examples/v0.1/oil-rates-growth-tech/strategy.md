# Oil Rates Growth Tech Strategy

## Intent

This strategy reduces growth technology exposure when oil prices and long-term interest rates rise together.

The central belief is:

> Rising oil plus rising long rates can create pressure on long-duration growth equities, especially when liquidity is already tight.

## Core Market Logic

The strategy watches for three related conditions:

1. WTI crude oil rising over a medium-term window.
2. U.S. 10-year Treasury yields rising over the same period.
3. A tight-liquidity regime becoming active.

When those conditions align, the strategy should:

- reduce exposure to broad growth technology assets such as `QQQ`
- reduce exposure to high-beta semiconductor or AI-related names such as `NVDA` and `AMD`
- raise cash
- optionally increase duration exposure through `TLT` if volatility and portfolio constraints allow it

## Risk Management Philosophy

The strategy should not aggressively short growth technology.

Instead, it should behave as a portfolio-aware risk reducer:

- trim overweight growth technology
- raise cash toward target
- avoid violating concentration limits
- avoid excessive turnover
- respect max single-position limits
- respect portfolio-level cash minimums

## Expected Behavior

The strategy is expected to underperform during strong risk-on technology rallies where oil and rates rise without hurting equity sentiment.

It is expected to help reduce drawdowns during tightening or inflationary shocks when growth technology weakens.

## Assets

Primary assets:

- `QQQ`
- `NVDA`
- `AMD`
- `TLT`
- `USO`
- `SPY`

## Example Rule

If:

- `wti_change_20d > 10%`
- `ust10y_change_20d > 25 basis points`
- `tight_liquidity` regime is active

Then:

- decrease `QQQ` weight by 5%
- decrease `NVDA` weight by 3%
- increase cash by 8%

The portfolio engine may scale or reject these actions if they violate constraints.
