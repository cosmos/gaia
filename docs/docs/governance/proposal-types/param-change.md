---
title: Parameter Changes
order: 4
---

This documentation aims to provide guidelines for creating and assessing parameter-change proposals.

Drafting and submitting a parameter-change governance proposal involves two kinds of risk: losing proposal deposit amounts and the potential to alter the function of the Cosmos Hub network in an undesirable way. 

## What parameters can be changed?

The complete parameters of the Cosmos Hub are split up into different modules, each of which has its own set of parameters. Most parameters can be updated by submitting a governance proposal.

List of modules whose parameters can be changed via governance:
* x/auth
* x/bank
* x/distribution
* x/evidence
* x/feegrant
* x/gov
* x/mint
* x/slashing
* x/staking
* ibc-go/transfer
* interchain-security/provider

Each cosmos-sdk module uses `MsgUpdateParams` for providing parameter changes. You can learn more about it in the cosmos-sdk documentation of each module (e.g. https://docs.cosmos.network/v0.47/build/modules/staking#msgupdateparams)

## What are the current parameter values?
<!-- markdown-link-check-enable -->
There are ways to query the current settings for each module's parameter(s). Some can be queried with the command line program [`gaiad`](../../getting-started/installation).

You can begin by using the command `gaiad q [module] -h` to get help about the subcommands for the module you want to query. For example, `gaiad q staking params` returns the settings of relevant parameters:

```sh
bond_denom: uatom
historical_entries: 10000
max_entries: 7
max_validators: 180
unbonding_time: 1814400s
```

If a parameter-change proposal is successful, the change takes effect immediately upon completion of the voting period.

**Note:** You cannot currently query the `bank` module's parameter, which is `sendenabled`. You also cannot query the `crisis` module's parameters.

## Why create a parameter change proposal?

Parameters are what govern many aspects of the chain's behaviour. As circumstances and attitudes change, sometimes you might want to change a parameter to bring the chain's behaviour in line with community opinion. For example, the Cosmos Hub launched with 100 active validators and there have been 4 proposals to date that have increased the `MaxValidators` parameter. At the time of writing, the active set contains 180 validators.

The Cosmos Hub has been viewed as a slow-moving, highly secure chain and that is reflected in some of its other parameters, such as a 21 day unbonding period and 14 day voting period. These are quite long compared to other chains in the Cosmos Ecosystem

## Risks in parameter change proposals

Because parameters dictate some of the ways in which the chain operates, changing them can have an impact beyond what is immediately obvious. 

For example, reducing the unbonding period might seem like the only effect is in how quickly delegators can liquidate their assets. It might also have a much greater impact on the overall security of the network that would be hard to realize at first glance.

This is one of the reasons that having a thorough discussion before going on-chain is so important - talking through the impacts of a proposal is a great way to avoid unintended effects.
