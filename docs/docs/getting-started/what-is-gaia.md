---
title: What is Gaia?
order: 1
---

`gaia` is the name of the Cosmos SDK application for the Cosmos Hub.

The Gaia Daemon and command-line interface (CLI) is `gaiad`.  `gaiad` runs a full-node of the `gaia` application.

`gaia` is built on the Cosmos SDK using the following modules:

- `x/auth`: Accounts and signatures.
- `x/bank`: Token transfers.
- `x/staking`: Staking logic.
- `x/mint`: Inflation logic.
- `x/distribution`: Fee distribution logic.
- `x/slashing`: Slashing logic.
- `x/gov`: Governance logic.
- `ibc-go/modules`: Inter-blockchain communication. Hosted in the `cosmos/ibc-go` repository.
- `x/params`: Handles app-level parameters.
- `x/globalfee`: Manages a number of fee parameters at the network and node level.

About the Cosmos Hub: The Cosmos Hub is the first Hub to be launched in the Cosmos Network. The role of a Hub is to facilitate transfers between blockchains. If a blockchain connects to a Hub via IBC, it automatically gains access to all the other blockchains that are connected to it. The Cosmos Hub is a public Proof-of-Stake chain. Its staking token is called the Atom. With the integration of Interchain Security, the Cosmos Hub acts as the provider chain, for the Atom Economic Zone.

Next, learn how to [install Gaia](./installation.mdx).
