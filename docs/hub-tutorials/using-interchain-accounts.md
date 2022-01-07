<!--
order: 6
-->

# Using Interchain Accounts

**Interchain Accounts** (ICA) is a standard that allows an account on a *controller* chain to create and securely control an address on a different *host* chain using the Inter Blockchain Protocol (IBC). Transactions that are native to the host chain are wrapped inside an IBC packet and get sent from the Interchain Account on the controller chain, to be executed on the host chain. 

The benefit of ICA is that there is no need to create a custom IBC implementation for each of the unique transactions that a sovereign blockchain might have (trading on a DEX, executing a specific smart contract, etc). Instead, a **generic** implementation allows blockchains to speak to each other, much like contracts can interact on Ethereum or other smart contract platforms.

For example, let’s say that you have an address on the Cosmos Hub (the controller) with OSMO tokens that you wanted to stake on Osmosis (the host). With Interchain Accounts, you can create and control a new address on Osmosis, without requiring a new private key. After sending your tokens to your Interchain Account using a regular IBC token transfer, you can send a wrapped `delegate` transaction over IBC which will then be unwrapped and executed natively on Osmosis.

Blockchains implementing Interchain Accounts can decide which messages they allow a controller chain to execute via a whitelist. At the time of writing, all messages will be allowed on the Cosmos Hub. 

Interchain Accounts will be launched with the Theta Upgrade (expected Q1 2022). If you'd like to learn more before launch, you can use this tutorial with the `gaiad` binary from [this branch](https://github.com/cosmos/gaia/tree/ica-acct-auth).

## How to use
