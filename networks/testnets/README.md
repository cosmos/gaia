### 📌 The `theta-testnet-001` testnet will remain the primary Cosmos Hub testnet following the v7-Theta upgrade. It will be used to test v8-Rho, v9-Lambda, and further upgrades.

# Cosmos Hub Testnets

This repository contains current and archived genesis files, scripts, and configurations for Cosmos Hub testnets. Key network information is present here, but please check out the tutorial in the [Cosmos Hub documentation](https://hub.cosmos.network/main/hub-tutorials/join-testnet.html) for step-by-step instructions on how to join whichever testnet is currently running. You can also find legacy network information [here](legacy/).

## Cosmos Hub Testnet Plan

The goals of the Cosmos Hub testnet program are to:

-  Build confidence among validators and full node operators to test upcoming software upgrades.
-  Provide developers with a low-risk production-like environment to test integrations with Cosmos Hub.
-  Make historical hub software releases and states readily available for testing and querying.

Beyond these goals, testnets could also become a site for R&D for new development and governance approaches in a fast-moving and live context.

### [Public Testnet](public/)

The public testnet targets validators who want to participate in a simulated chain upgrade before the mainnet upgrade takes place. Shortly after a new Gaia version is available, we submit a software upgrade proposal, vote on it, and update all nodes with the new binary at the halt height specified in the proposal.

Up until the `Vega` testnet, our approach was to deploy a testnet for each Gaia upgrade.

**Starting with the `theta-testnet-001` testnet, we have moved to a persistent testnet model. This testnet will stay online and remain the primary Cosmos Hub testnet after the `v7-Theta` upgrade, including for the v8-Rho and v9-Lambda upgrades, and beyond.**

Based on our experience with `Vega`, we have configured the public testnet so that:
* Testnet coordinators will operate 4+ validators with combined voting power exceeding 75% total power.
* These validators will require an addition of ~550M bonded test ATOM (current bonded ATOM are ~180M) and a corresponding increase in total supply.
* Tesnet coordinators control a faucet with >100M liquid tokens.
* Testnet coordinators can reward validators with limited edition secondary tokens that are named after their release (`Theta`, `Rho`, `Epsilon`, `Lambda`). The testnet will have a fixed supply of 1000 each of such tokens.

### [Replicated Security Persistent Testnet](replicated-security/)

The Replicated Security testnet provides a public platform to explore:
- Launching and stopping consumer chains
- Interchain Security features
- Relayer operations
- Integrations (block explorers, monitors, etc.)

We have configured this testnet so that:
* Testnet coordinators operate 3+ validators with a combined voting power exceeding 75% total power.
* Testnet coordinators control a faucet with >100M liquid tokens.

### [Developer Testnet](devnet/)

The devnet is specially set up for block explorers, wallets, exchanges, and other integrators, to give early endpoints for you to test against. We update the node binaries to the latest branch of the [`cosmos/gaia`](https://github.com/cosmos/gaia) repo to give you the most current software version.


### [Local Testnet](local/)

A local testnet can be set up to experiment in a local single-validator environment.
