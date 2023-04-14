---
name: Cosmos Hub Upgrade Checklist
about: Create a checklist for an upgrade
labels: epic, needs-triage
---

## Cosmos Hub Upgrade Epic

# <Upgrade Name>

**Create an issue for each item** and mark complete once it has been done.

<!-- TODO: add time estimates for comms -->

```[tasklist]
### Communication (during entire lifecycle)
- [ ]  Signaling proposal (before development starts)
- [ ]  Testnet blog post - target validators on Cosmos Medium
- [ ]  Tweet link to testnet upgrade blog - @ cosmohub
- [ ]  Testnet upgrade info (discord only)
- [ ]  Tweet updates on proposal status - @ cosmohub
- [ ]  Mainnet blog post - target wider audience on Cosmos Medium
- [ ]  Tweet link to mainnet upgrade blog - @ cosmos @ cosmoshub  
- [ ]  Link to mainnet upgrade instructions (all channels - Discord, Telegram, Slack)
- [ ]  Tweet upgrade countdown during voting period - @ cosmos @ cosmoshub  
- [ ]  Tweet upgrade success story - @ cosmos @ cosmoshub  
```

```[tasklist]
### Library dependencies
- [ ]  Upgrade to SDK version <SDK VERSION>
- [ ]  Upgrade to IBC version <IBC VERSION>
- [ ]  Upgrade to ICS version <ICS VERSION>
- [ ]  Upgrade to PFM version <PFM VERSION>
- [ ]  Upgrade to Liquidity version <Liquidity VERSION>
- [ ]  Integrate new modules ([checklist](https://github.com/cosmos/hub-eng/blob/main/module_qa/module_checklist.md))
```

```[tasklist]
### Testnet
- [ ]  Communication prep
- [ ]  Docs:
  - [ ]  [testnets](https://github.com/cosmos/testnets) updated with most recent rc
  - [ ]  [join-testnet](https://github.com/cosmos/gaia/blob/main/docs/hub-tutorials/join-testnet.md)
- [ ]  Release candidate
- [ ]  Create testnet proposal
- [ ]  Run testnet for one week
- [ ]  Final Release
```

```[tasklist]
### Docs
- On release branch 
  - [ ] Quickstart in `docs/getting-started/quickstart.md`
  - [ ] Join mainnet in `docs/hub-tutorials/join-mainnet.md`
  - [ ] Migration docs in `docs/migration/`
  - [ ] Update `CHANGELOG.md`
    - [ ]  Breaking REST api changes
    - [ ]  Breaking CLI api changes
- Post Upgrade  
  - [ ] [chain-registry.json](https://github.com/cosmos/chain-registry/blob/master/cosmoshub/chain.json)
  - [ ] Update [cosmos mainnet repo](https://github.com/cosmos/mainnet)
```

```[tasklist]
### Mainnet Proposal
- [ ]  Predict block height for target date
- [ ]  Create forum post
- [ ]  Submit on-chain proposal
```


