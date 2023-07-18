# CHANGELOG

## v11.0.0

*July 18, 2023*

### API BREAKING

- [GlobalFee](x/globalfee)
  - Add `bypass-min-fee-msg-types` and `maxTotalBypassMinFeeMsgGagUsage` to
    globalfee params. `bypass-min-fee-msg-types` in `config/app.toml` is
    deprecated ([\#2424](https://github.com/cosmos/gaia/pull/2424))

### BUG FIXES

- Fix logic bug in `GovPreventSpamDecorator` that allows bypassing the 
  `MinInitialDeposit` requirement 
  ([a759409](https://github.com/cosmos/gaia/commit/a759409c9da2780663244308b430a7847b95139b))

### DEPENDENCIES

- Bump [PFM](https://github.com/strangelove-ventures/packet-forward-middleware) to 
  [v4.0.5](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v4.0.5)
  ([\#2185](https://github.com/cosmos/gaia/issues/2185))
- Bump [Interchain-Security](https://github.com/cosmos/interchain-security) to
  [v2.0.0](https://github.com/cosmos/interchain-security/releases/tag/v2.0.0)
  ([\#2616](https://github.com/cosmos/gaia/pull/2616))
- Bump [Liquidity](https://github.com/Gravity-Devs/liquidity) to 
  [v1.6.0-forced-withdrawal](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.6.0-forced-withdrawal) 
  ([\#2652](https://github.com/cosmos/gaia/pull/2652))

### STATE BREAKING

- General
  - Fix logic bug in `GovPreventSpamDecorator` that allows bypassing the
    `MinInitialDeposit` requirement
    ([a759409](https://github.com/cosmos/gaia/commit/a759409c9da2780663244308b430a7847b95139b))
  - Bump [Interchain-Security](https://github.com/cosmos/interchain-security) to
    [v2.0.0](https://github.com/cosmos/interchain-security/releases/tag/v2.0.0)
    ([\#2616](https://github.com/cosmos/gaia/pull/2616))
  - Bump [Liquidity](https://github.com/Gravity-Devs/liquidity) to
    [v1.6.0-forced-withdrawal](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.6.0-forced-withdrawal)
    ([\#2652](https://github.com/cosmos/gaia/pull/2652))
- [GlobalFee](x/globalfee)
  - Create the upgrade handler and params migration for the new Gloabal Fee module
    parameters introduced in [#2424](https://github.com/cosmos/gaia/pull/2424)
    ([\#2352](https://github.com/cosmos/gaia/pull/2352))
  - Add `bypass-min-fee-msg-types` and `maxTotalBypassMinFeeMsgGagUsage` to
    globalfee params ([\#2424](https://github.com/cosmos/gaia/pull/2424))
  - Update Global Fee's AnteHandler to check tx fees against the network min gas
    prices in DeliverTx mode ([\#2447](https://github.com/cosmos/gaia/pull/2447))

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

