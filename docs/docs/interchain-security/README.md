---
title: Interchain Security
order: 3
---

The [Interchain Security](https://cosmos.github.io/interchain-security/) feature brings to the Cosmos Hub a shared security model, where the Cosmos Hub validators, also validate consumer chains. This is valuable for consumer chains, as consumer chains can focus on product-market fit, rather than business and operational agreements in bringing together a validator set. As part of this agreement, consumer chains pay for the security by distributing a portion of the consumer chain revenue to Hub token holders.

All potential chains are onboarded as consumer chains, via Hub Governance, with the feedback from the Hub community.

## ICS features

[Partial Set Security](https://cosmos.github.io/interchain-security/features/partial-set-security) and [Power Shaping](https://cosmos.github.io/interchain-security/features/power-shaping) bring benefits for both the consumer chains and validators:

### Top-N consumer chains

Validators inside the top-N percent of voting power are required to validate the consumer chain.

e.g. `top-95` means that the 95% of the validators (by voting power) are required to run the consumer chain binary

### Opt-in consumer chains

Only validators that opt to running a consumer chains are required to run the chain binary and become eligible for consumer chain rewards distribution.

### Parameter customization

Consumer chains gain the ability to customize the validator set to their needs:

- define allow/denylists
- set maximum number of validators
- set validator power cap

## Notable consumer chains

Currently the Cosmos Hub has the following two Consumer Chains.

### Neutron

[Neutron](https://neutron.org/), is a smart contracting platform, that was the first consumer chain onboarded.  
Neutron was onboarded as a consumer chain in May 2023, see Hub [proposal 792](https://www.mintscan.io/cosmos/proposals/792) for more details.

### Stride

[Stride](https://www.stride.zone/), is a liquid staking provider, which aims to unlock liquidity for Cosmos Hub token holders.  
Stride was onboarded as a consumer chain in July 2023, see Hub [proposal 799](https://www.mintscan.io/cosmos/proposals/799) for more details.

## Resources

For more information visit:

- [Interchain Security docs](https://cosmos.github.io/interchain-security)
