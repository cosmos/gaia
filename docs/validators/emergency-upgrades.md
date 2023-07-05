---
order: 5
title: Emergency Upgrades
---

## **Introduction**

This article goes into the general flow of events for an emergency upgrade, pre-upgrade and during the upgrade for all validators.

Emergency upgrades are an unfortunate necessity for all validators when running a Proof-of-Stake node. When a vunerability is discovered and responsibily disclosed, the vunerability will be reported to one of the channels mentioned in the [Security.md](../../SECURITY.md).

The process for updates has the following steps:

- **Heads-up** - An annoucement from the relevant teams is made to validators, on the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) channel, that a priority fix is incoming. As a result, validators should reserve resources to handle the changes in the next 1-3 of days. The messaging could also establish an official private communications channel for co-ordination.
- **Release** - An official release will be made on GitHub on the Cosmos org. This should confirm the authenticity of the fix for validators. This will be an actual release of the source code that validators can inspect.
- **Deployment** - Validator coordination and one-to-one communications to help them push through the changes.
- **Advisory** - An advisory is released when all the impacted parties are upgraded and generally comes from the relevant teams responsible for the release **OR** the team responsible for patching the source code.


### **Cosmos Hub Upgrade Information**

Depending on the severity of the vulnerability, the relevant teams and validators can take different actions.

The relevant teams will try to reach out to validators as soon as possible, so that internal planning can be adjusted accordingly.
Validators will have time to inspect the changes circa **24 hours** or more before a planned upgrade.

If the change is **does not** require co-ordination, then the patch can be applied by validators **asap**.

If the change requires co-ordination, then all validators will need to set a block or halt-height as highlighted in the communications.

#### **Co-ordinated Upgrades**

A co-ordinated upgrade will require updating your node configuration to the correct halt-height. The relevant teams will communicate the correct height for all validators via Email and [Discord](https://discord.com/channels/669268347736686612/798937713474142229).

At the halt-height, you will swap out the old binaries for the new ones, while monitoring your node. Be sure to check-in on the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) channel on [Cosmos Developers Discord](https://discord.gg/cosmosnetwork), as notifications will be made pre/during/post upgrade on **this** channel.

If you do have issues with the upgrade, please bring them up in the [Discord channel](https://discord.com/channels/669268347736686612/798937713474142229) where the core teams will be monitoring progress and will help you resolve any issues that arise.

Once the threshold of 67% voting power of upgraded validators is reached, the Cosmos Hub network will start producing blocks. Note that this process might take between 5-20 mins depending on the required migrations for the upgrade and voting power present. Please be patient and check the [Discord channel](https://discord.com/channels/669268347736686612/798937713474142229) for any updates.

If there are any serious issues with the upgrade then you will notified on Discord to rollback the changes.

#### **Rolling back an upgrade**

During the network upgrade, the relevant Cosmos teams will be monitoring the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229)  channel and communicating with validators on the status of the upgrade.

During this time, the relevant teams will listen to validator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the relevant teams, after conferring with validators, may choose to declare that the upgrade will be skipped.

**Note**: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by relevant teams.

### **Vulnerability Levels**

This section discusses the actions to be taken for different vulnerability levels.

#### **Critical level vulnerabilities**

For critical level vulnerabilities, the relevant teams will usually start a private patching effort with a number of key validators. By the time an emergency patch is released, the expectation is that a good proportion of the validator set have already upgraded. Vulnerabilities at this level should be updated immediately. We might ask validators to shut down the network until a patch is released, due to the potential consquences arising from the identified vulnerability.

#### **High level vulnerabilities**

For high level vulnerabilities, the relevant teams will reach out to validators to upgrade the chain asap.

#### **Medium level vulnerabilities**

For medium level vulnerabilities, the team will release a patch and may consider an out-of-band (emergency) release. We encourage validators to upgrade asap.

#### **Low level vulnerabilities**

Fixes for low level vulnerabilities are periodically rolled into point releases or if state breaking, they are rolled into the next minor or major release.

## **What to expect before the upgrade?**

Once a patch is ready, the relevant teams will reach out to all the validators via a number of pre-established public or private channels depending on the severity of the vulnerability.

Validators **MUST HAVE** their contact details on file by registering their teams on [Cosmos Developers Discord](https://discord.gg/cosmosnetwork) and join the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) channel. This is necessary so that validators don't miss important updates and get jailed for missing blocks or have other operational issues.

## **Main Contact Points**

The primary contact points for the Cosmos Hub are the relevant teams at Informal Systems and the [Cosmos Developers Discord](https://discord.gg/cosmosnetwork) on the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) channel.

## **Community**

Discuss the finer details of being a validator on our community Discord and sign up for the Cosmos newsletter to get regular updates:

* [Cosmos Developers Discord](https://discord.gg/cosmosnetwork)
* [Newsletter](https://cosmos.network/updates/signup/)
