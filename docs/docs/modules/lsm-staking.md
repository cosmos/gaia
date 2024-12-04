---
title: LSM Staking
order: 1
---

The `x/staking` module used by the Hub includes extensions that enable liquid staking
You can read more about it in our [LSM documentation](https://github.com/cosmos/cosmos-sdk/tree/v0.50.9-lsm/x/staking#totalliquidstakedtokens).

## What are LSM shares

LSM shares are derivatives of the delegation shares. They are tied to a delegator and a validator pair and they represent the underlying delegation shares.
By issuing LSM shares, the underlying staked ATOM can become "liquid" while still being slashable. The LSM shares are tokens that can be used in various DeFi protocols and transferred between users or between chains via IBC.

LSM shares are not fungible (as they are tied to a delegator/validator pair) and are issued by the Hub directly and thus don't depend on the security of any entity other than the Cosmos Hub itself.

## Benefits

By tokenizing your staked ATOM into LSM shares, you maintain the benefits of staking while gaining flexibility in using these shares in DeFi protocols and platforms.

The LSM shares issued by the Hub are powering liquid staking derivatives like stATOM or dATOM and they are the backbone of the Hydro platform.
