<!--
order: 2
-->

# ADR 001: Interchain Accounts

## Changelog

- 2022-02-04: added content
- 2022-01-19: init
- 2023-06-28: mark as rejected

## Status

REJECTED Not Implemented

**Reason:** The IBC team decided to integrate this functionality directly into their codebase and maintain it, because multiple users require it. 

## Abstract

This is the Core Interchain Accounts Module. It allows the Cosmos Hub to act as a host chain with interchain accounts that are controlled by external IBC connected "Controller" blockchains. Candidate chains include Umee, Quicksilver, Sommelier. It is also a necessary component for a Authentication Module that allows the Cosmos Hub to act as a Controller chain as well. This will be recorded in a separate ADR.

## Rationale

This allows the Hub to participate in advanced cross-chain defi operations, like Liquid Staking and various protocol controlled value applications.

## Desired Outcome

The hub can be used trustlessly as a host chain in the configuration of Interchain Accounts.

## Consequences

There has been preliminary work done to understand if this increases any security feature of the Cosmos Hub. One thought was that this capability is similar to contract to contract interactions which are possible on virtual machine blockchains like EVM chains. Those interactions introduced a new attack vector, called a re-entrancy bug, which was the culprit of "The DAO hack on Ethereum". We believe there is no risk of these kinds of attacks with Interchain Accounts because they require the interactions to be atomic and Interchain Accounts are asynchronous.

#### Backwards Compatibility

This is the first of its kind.

#### Forward Compatibility

There are future releases of Interchain Accounts which are expected to be backwards compatible.

## Technical Specification

[ICS-27 Spec](https://github.com/cosmos/ibc/blob/master/spec/app/ics-027-interchain-accounts/README.md)

## Development

- Integration requirements
  - Development has occurred in [IBC-go](https://github.com/cosmos/ibc-go) and progress tracked on the project board there.
- Testing (Simulations, Core Team Testing, Partner Testing)
  - Simulations and Core Team tested this module
- Audits (Internal Dev review, Third-party review, Bug Bounty)
  - An internal audit, an audit from Informal Systems, and an audit from Trail of Bits all took place with fixes made to all findings.
- Networks (Testnets, Productionnets, Mainnets)
  - Testnets

## Governance [optional]

- **Needs Signaling Proposal**
- Core Community Governance
  - N/A
- Steering Community
  - N/A. Possibly Aditya Srinpal, Sean King, Bez?
- Timelines & Roadmap
  - Expected to be released as part of IBC 3.0 in Feb 2022 (currently in [beta release](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0-beta1))

## Project Integrations [optional]

- Gaia Integrations
  - [PR](https://github.com/cosmos/gaia/pull/1150)
- Integration Partner
  - IBC Team

#### Downstream User Impact Report

(Needs to be created)

#### Upstream Partner Impact Report

(Needs to be created)

#### Inter-module Dependence Report

(Needs to be created)

## Support

[Documentation](https://ibc.cosmos.network/main/apps/interchain-accounts/overview.html)

## Additional Research & References

- [Why Interchain Accounts Change Everything for Cosmos Interoperability](https://medium.com/chainapsis/why-interchain-accounts-change-everything-for-cosmos-interoperability-59c19032bf11)
- [Interchain Account Auth Module Demo Repo](https://github.com/cosmos/interchain-accounts)