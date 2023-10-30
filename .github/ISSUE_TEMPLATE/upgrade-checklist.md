---
name: Cosmos Hub Upgrade Checklist
about: Create a checklist for an upgrade
labels: epic, needs-triage
---

## Cosmos Hub Upgrade to Gaia <Version>

<!-- TODO: Replace <Version> by the actual Gaia version -->

```[tasklist]
### After Cutting Release Candidate
- [ ] Coordinate with Hypha to test release candidate
- [ ] Create proposal text draft
- [ ] Post proposal text draft on forum
- [ ] Upgrade release and replicated security testnets (note: on Wednesdays)
- [ ] Review post-upgrade status of affected features if necessary
```

```[tasklist]
### Before Proposal Submission (TODO sync on a call)
- [ ] Cut final release
- [ ] Predict block height for target date
- [ ] Update/proofread proposal text
- [ ] Transfer deposit amount (i.e., 250 ATOMs) to submitter wallet 
- [ ] Create upgrade docs (with disclaimer upgrade prop still being voted on)
- [ ] Coordinate with marketing/comms to prep communication channels/posts
```

```[tasklist]
### Voting Period
- [ ] Estimate threshold of validators that are aware of proposal and have voted or confirmed their vote
- [ ] Coordinate with marketing/comms to update on voting progress (and any change in upgrade time)
```

```[tasklist]
## Proposal Passed 
- [ ] Determine "on-call" team: available on Discord in [#cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) during upgrade 
- [ ] Coordinate with marketing/comms on who will be available, increase regular upgrade time updates and validator outreach
- [ ] Prep Gaia docs: `docs/getting-started/quickstart.md`, `docs/hub-tutorials/join-mainnet.md`, `docs/migration/` (open PR)
- [ ] Prep chain-registry update: [cosmoshub/chain.json](https://github.com/toschdev/chain-registry/blob/master/cosmoshub/chain.json) (open PR)
- [ ] Prep [cosmos mainnet repo](https://github.com/cosmos/mainnet) update (open PR)
- [ ] Prep internal statesync node for upgrade (confirm cosmovisor configured) 
- [ ] Reach out to main dependency teams -- Comet, IBC, SDK -- for assistance during the upgrade (#gaia-release-warroom on Slack)
```

```[tasklist]
## During Upgrade (note: on Wednesdays at 15:00 UTC)
- [ ] Available on Discord in [#cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229)
- [ ] Available on Twitter / Slack / Telegram 
```

```[tasklist]
## Post Upgrade 
- [ ] Merge PRs for Gaia docs & chain-registry update
- [ ] FAQ: collect issues on upgrade from discord
- [ ] Hold validator feedback session
```
