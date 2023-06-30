---
order: 5
title: Emergency Upgrades
---

## **Introduction**

This article goes into the general flow of events for an emergency upgrade, pre-upgrade and during the upgrade for all validators.

Emergency upgrades are an unfortunate necessity for all validators when running a PoS node.

When a vunerability is discovered and responsibily disclosed, the vunerability will be reported to one of the channels described in the [Security.md](../../SECURITY.md).

Once the vulnerability is patched and tested, the core team will reach out to all validators to update their nodes with the latest version.

### **Cosmos Hub Upgrade Information**

Depending on the severity of the vulnerability, the core teams and validators can take different actions.

The core teams will try reach out **48 hours** or more before a scheduled upgrade, so that validators have time to inspect the changes and also adjust their internal planning accordingly.

If the change is **non** state-breaking then the patch can be applied in a un-cordinated manner by validators asap.

If the change is **state-breaking**, that will mean that the upgrade has to be co-ordinated for all validators at a certain block or halt-height.

#### **Co-ordinated Upgrades**

A co-ordinated upgrade will require updating your node configuration to the correct halt-height. The code teams will communicate the correct height for all validators via Email and [Discord](https://discord.com/channels/669268347736686612/798937713474142229).

At the halt-height, you will swap out the old binaries for the new ones, while monitoring your node. Be sure to check-in on the [Cosmos Developers Discord](https://discord.gg/cosmosnetwork) on the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229), as notifications will be made pre/during/post upgrade on **this** channel.

If you do have issues with the upgrade, please bring them up in the [Discord channel](https://discord.com/channels/669268347736686612/798937713474142229) where the core teams will be monitoring progress and will help you resolve any issues that arise.

Once the threshold of 67% voting power is reached by the upgraded validators, the Cosmos Hub network will start producing blocks. Note that this process might take between 5-20 mins depending on the required migrations for the upgrade and voting power present. Please be patient and check the [Discord channel](https://discord.com/channels/669268347736686612/798937713474142229) for any updates.

If there are any serious issues with the upgrade then you will notified on Discord to rollback the changes.

#### **Rolling back an upgrade**

During the network upgrade, core Cosmos teams will be monitoring the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229)  channel and communicating with validators on the status of the upgrade.

During this time, the core teams will listen to validator needs to determine if the upgrade is experiencing unintended challenges. In the event of unexpected challenges, the core teams, after conferring with validators, may choose to declare that the upgrade will be skipped.

**Note**: There is no particular need to restore a state snapshot prior to the upgrade height, unless specifically directed by core teams.

### **Vulnerability Levels**

This sections discusses the actions to be taken for different vulnerability levels.

#### **Critical level vulnerabilities**

For critical level vulnerabilities, the core teams will usually start a private patching effort with a number of key validators. By the time an emergency patch is released, the expectation is that a good proportion of the validator set have already upgraded. Vulnerabilities at this level should be updated immediately. We might ask validators to shut down the network until a patch is released, due to the potential consquences arising from the identified vulnerability.

#### **High level vulnerabilities**

For high level vulnerabilities, the core teams will reach out to validators to upgrade the chain asap.

#### **Medium level vulnerabilities**

For medium level vulnerabilities, the team will release a patch and may consider an out-of-band (emergency) release. We encourage validators to upgrade asap.

#### **Low level vulnerabilities**

Fixes for low level vulnerabilities are periodically rolled into point releases or if state breaking they are rolled into the next minor or major release.

## **What to expect before the upgrade?**

Once a patch is ready, the core teams will reach out to all the validators via a number of pre-established public or private channels depending on the severity of the vulnerability.

Validators **MUST HAVE** their contact details on file by registering their teams on [Cosmos Developers Discord](https://discord.gg/cosmosnetwork) and joining the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) channel. This is necessary so that validators don't miss important updates and get jailed for missing blocks or have other operational issues.

The core teams are happy to have a private Telegram channel with each validator team. Please reach out to the Discord admins to help set this up.

## **Main Contact Points**

The primary contact points for the Cosmos Hub are the core team at Informal Systems and the [Cosmos Developers Discord](https://discord.gg/cosmosnetwork) on the [cosmos-hub-validators-verified](https://discord.com/channels/669268347736686612/798937713474142229) channel.

## **Community**

Discuss the finer details of being a validator on our community Discord and sign up for the Cosmos newsletter to get regular updates:

* [Cosmos Developers Discord](https://discord.gg/cosmosnetwork)
* [Newsletter](https://cosmos.network/updates/signup/)
