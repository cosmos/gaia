---
name: Cosmos Hub Upgrade Checklist
about: Create a checklist for an upgrade

---

## Cosmos Hub Upgrade Epic

# <Upgrade Name>

Check mark each item if it has been completed and ready for the upgrade. If not added yet, create an issue for each item and mark complete once it has been done and integrated.

- [ ]  Upgrade to SDK version <SDK VERSION>
- [ ]  Upgrade to IBC version <IBC VERSION>
- [ ]  Integrate new modules ([checklist](https://github.com/cosmos/hub-eng/blob/main/module_qa/module_checklist.md))

- [ ]  Testnet
  - [ ]  Communication prep
  - [ ]  Docs:
    - [ ]  https://github.com/cosmos/testnets
    - [ ]  `gaia/docs/hub-tutorials/join-testnet.md`
  - [ ]  Release candidate
  - [ ]  Create testnet proposal
  - [ ]  Run testnet for one week
  - [ ]  Final Release

- [ ]  Docs:
  - [ ]  Prepare upgrade-docs branch to merge post upgrade
  - [ ]  https://github.com/cosmos/gaia
    - [ ]  `gaia/docs/hub-tutorials/join-mainnet.md`
    - [ ]  `gaia/docs/getting-started/quickstart.md`
    - [ ]  `README.md`
    - [ ]  `changelog.md`
      - [ ]  Breaking REST api changes
      - [ ]  Breaking CLI api changes
  - [ ]  https://github.com/cosmos/mainnet

- [ ]  Mainnet Proposal
  - [ ]  Predict block height for target date
  - [ ]  Create forum post
  - [ ]  Submit on-chain proposal

- [ ]  Communication
  - [ ]  Blogposts
    - [ ]  Testnet - target validators on Interchain Medium
    - [ ]  Mainnet - target wider audience on Cosmos Medium
  - [ ]  Twitter
    - [ ]  Link to testnet upgrade blog - @ cosmohub
    - [ ]  Link to mainnet upgrade blog - @ cosmos @ cosmoshub
    - [ ]  Upgrade countdown - @ cosmos @ cosmoshub
    - [ ]  Upgrade success story - @ cosmos @ cosmoshub
  - [ ]  Discord, Telegram, Slack
    - [ ]  Testnet upgrade info (discord only)
    - [ ]  Link to mainnet upgrade instructions (all channels)
  - [ ]  Validator email list
    - [ ]  Upgrade instructions + link to mainnet blog - one week before upgrade
  - [ ]  Update chain registry after upgrade
