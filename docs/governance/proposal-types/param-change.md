---
order: 4
parent:
  order: 1
---

# Parameter Change

This Cosmos Hub educational documentation aims to outline the [Hub's parameters](#params-wiki), describe their functions, and describe the potential implications of modifying each parameter. This documentation also aims to provide guidelines for creating and assessing parameter-change proposals.



Drafting and submitting a parameter-change governance proposal involves two kinds of risk: losing proposal deposit amounts and the potential to alter the function of the Cosmos Hub network in an undesirable way. 

## What parameters can be changed?

The complete parameters of the Cosmos Hub are split up into different modules, each of which has its own set of parameters. Any of them can be updated with a Param Change Proposal. If you are technically inclined, this is the full [list of modules](https://github.com/cosmos/cosmos-sdk/tree/master/x) in the Cosmos SDK. The Cosmos Hub is built using the Cosmos SDK, but not all available modules are in use on the Hub.

There are currently 8 modules active in the Cosmos Hub with parameters that may be altered via governance proposal. New modules may be introduced in the future.
1. [auth](./params-change/Auth.md) - Authentication of accounts and transactions
2. [bank](./params-change/Bank.md) - Token transfer functionalities
3. [gov](./params-change/Governance.md) - On-chain governance proposals and voting
4. [staking](./params-change/Staking.md) - Proof-of-stake layer
5. [slashing](./params-change/Slashing.md) - Validator punishment mechanisms
6. [distribution](./params-change/Distribution.md) - Fee distribution and staking token provision distribution
7. [crisis](./params-change/Crisis.md) - Halting the blockchain under certain circumstances (ie. if an invariant is broken)
8. [mint](./params-change/Mint.md) - Creation of new units of staking token
<!-- markdown-link-check-disable -->

## What are the current parameter values?
<!-- markdown-link-check-enable -->
There are ways to query the current settings for each module's parameter(s). Some can be queried with the command line program [`gaiad`](../../getting-started/installation.md).

You can begin by using the command `gaia q [module] -h` to get help about the subcommands for the module you want to query. For example, `gaiad q staking params --chain-id <chain-id> --node <node-id>` returns the settings of relevant parameters:

```
bond_denom: uatom
historical_entries: 10000
max_entries: 7
max_validators: 175
unbonding_time: 1814400s
```

If a parameter-change proposal is successful, the change takes effect immediately upon completion of the voting period.


**Note:** You cannot currently query the `bank` module's parameter, which is `sendenabled`. You also cannot query the `crisis` module's parameters.


## Why create a parameter change proposal?
Parameters are what govern many aspects of the chain's behaviour. As circumstances and attitudes change, sometimes you might want to change a parameter to bring the chain's behaviour in line with community opinion. For example, the Cosmos Hub launched with 100 active validators and there have been 3 proposals to date that have increased the `MaxValidators` parameter. At the time of writing, the active set contains 175 validators.

The Cosmos Hub has been viewed as a slow-moving, highly secure chain and that is reflected in some of its other parameters, such as a 21 day unbonding period and 14 day voting period. These are quite long compared to other chains in the Cosmos Ecosystem

## Risks in parameter change proposals 
Because parameters dictate some of the ways in which the chain operates, changing them can have an impact beyond what is immediately obvious. 

For example, reducing the unbonding period might seem like the only effect is in how quickly delegators can liquidate their assets. It might also have a much greater impact on the overall security of the network that would be hard to realize at first glance.

This is one of the reasons that having a thorough discussion before going on-chain is so important - talking through the impacts of a proposal is a great way to avoid unintended effects.

## Credits

This documentation was originally created by Gavin Birch ([Figment Networks](https://figment.io)). Its development was supported by funding approved on January 29, 2020 by the Cosmos Hub via Community Spend [Proposal 23](https://cosmoshub-3.bigdipper.live/proposals/23) ([full Proposal PDF here](https://ipfs.io/ipfs/QmSMGEoY2dfxADPfgoAsJxjjC6hwpSNx1dXAqePiCEMCbY)). In late 2021 and early 2022 significant updates were made by [Hypha Worker Co-op](https://hypha.coop/), especially @dcwalk and @lexaMichaelides.  🙏

**Special thanks** to the following for providing credible information:
- Aleks (All in Bits; Fission Labs) for answering countless questions about these parameters
- Alessio (All in Bits) for explaining how [`SigVerifyCostED25519`](https://hub.cosmos.network/main/governance/proposal-types/params-change/Auth.html#4-sigverifycosted25519) & [`SigVerifyCostSecp256k1`](https://hub.cosmos.network/main/governance/proposal-types/params-change/Auth.html#5-sigverifycostsecp256k1) work, and detailed answers to my many questions
- Vidor for volunteering to explain [`ConstantFee`](https://hub.cosmos.network/main/governance/proposal-types/params-change/Crisis.html#1-constantfee) and answering my many questions in detail
- Hyung (B-Harvest) for volunteering how [`InflationRateChange`](https://hub.cosmos.network/main/governance/proposal-types/params-change/Mint.html#2-inflationratechange) works
- Joe (Chorus One) for explaining the security details involved with using full nodes for transactions
- Sunny (All in Bits; Sikka) for volunteering an explanation of the purpose of [`withdrawaddrenabled`](https://hub.cosmos.network/main/governance/proposal-types/params-change/Distribution.html#4-withdrawaddrenabled)
