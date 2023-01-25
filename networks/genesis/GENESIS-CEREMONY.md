# Genesis Validator Ceremony

Welcome to the Cosmos Hub Genesis Validator Ceremony!

**NOTE**: The Ceremony has been completed. These instructions are now stale.
They are left here for archival purposes. The `penultimate_genesis.json` they
refer to has been updated to match the final recommended genesis.json.

## What is it?

This *is not* the launch of the Cosmos Hub. Before a blockchain like the
Cosmos Hub can launch, it needs to determine an initial validator set.

This is a ceremony to establish a decentralized initial validator set
that can be recommended for the Genesis State of the Cosmos Network.
This validator set is computed from the set of signed `gentx` transactions with non-zero ATOMs submitted during this genesis ceremony.

Before you consider participating in this ceremony, please read the entire
document.

Genesis transactions will be collected on Github in this repository and checked for validity by an automated script.
Genesis file collection will terminate on 12 March 2019 23:00 GMT. The final recommended genesis file will be published shortly after that time.

By participating in this ceremony and submitting a gen-tx, you are making a commitment to your fellow Cosmonauts
that you will be around to bring your validator online by the recommended genesis time of 13 March 2019 23:00 GMT to launch the network. Note that you can start `gaiad` 
with the recommended genesis file before that time and, assuming you configure it successfully, it will automatically start the peer-to-peer and consensus processes once the genesis timestamp is reached.

Please keep the following things in mind.

1. This process is intended for technically inclined people who have participated in Cosmos testnets and Game of Stakes. If you aren't already familiar with this process, you are advised against participating due to the risks involved. There is no need for you to participate if you feel unprepared - 
 you can create a validator or stake ATOMs any time after launch.
2. ATOMs staked during genesis will be at risk of 5% slashing if your validator double signs. If you accidentally misconfigure your validator setup, this can easily happen, and slashed ATOMs are not expected to be recoverable by any means. Additionally, if you double-sign, your validator will be [tombstoned](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/slashing/07_tombstone.md) and you will be required to change operator and signing keys.
3. ATOMs staked during genesis or after will be locked up as part of the defense against long range attacks for 3 weeks. They can be re-delegated or undelegated, but will not be transferrable until a hard-fork enables transfers.
   

## Genesis File

**WARNING: THIS IS NOT THE FINAL RECOMMENDATION FOR THE GENESIS FILE**

This repository contains a work-in-progress recommendation for the genesis file called [`penultimate_genesis.json`](./penultimate_genesis.json).
It **IS NOT** the final recommended genesis file.
If you find an error in this genesis file, please contact us
immediately at "genesis at interchain dot io".

To understand how this file was compiled, please see [GENESIS.md](GENESIS.md).

A final recommendation will be available shortly, including a justification for
all components of the genesis file and scripts to recompute it.

Anyone with an ATOM allocation in the [`penultimate_genesis.json`](./penultimate_genesis.json) who intends to participate in the genesis ceremony must submit a pull request
containing a valid `gen-tx` to this repository in the `/gentx` folder with a file name like `<moniker>.json`.

## Instructions

Generally the steps to create a validator are as follows:

1. [Install Gaiad and Gaiacli version v0.33.0](https://github.com/cosmos/cosmos-sdk/blob/master/docs/gaia/installation.md)

2. [Setup your fundraiser keys](https://github.com/cosmos/cosmos-sdk/blob/master/docs/gaia/delegator-guide-cli.md#restoring-an-account-from-the-fundraiser)

3. Download the [genesis file](https://raw.githubusercontent.com/cosmos/launch/master/penultimate_genesis.json) to `~/.gaiad/config/genesis.json`

4. Sign a genesis transaction:

```bash
gaiad gentx \
  --amount <amount_of_delegation_uatom> \
  --commission-rate <commission_rate> \
  --commission-max-rate <commission_max_rate> \
  --commission-max-change-rate <commission_max_change_rate> \
  --pubkey <consensus_pubkey> \
  --name <key_name>
```

This will produce a file in the ~/.gaiad/config/gentx/ folder that has a name with the format `gentx-<node_id>.json`. The content of the file should have a structure as follows:

```json
{
  "type": "auth/StdTx",
  "value": {
    "msg": [
      {
        "type": "cosmos-sdk/MsgCreateValidator",
        "value": {
          "description": {
            "moniker": "<moniker>",
            "identity": "",
            "website": "",
            "details": ""
          },
          "commission": {
            "rate": "<commission_rate>",
            "max_rate": "<commission_max_rate>",
            "max_change_rate": "<commission_max_change_rate>"
          },
          "min_self_delegation": "1",
          "delegator_address": "cosmos1msz843gguwhqx804cdc97n22c4lllfkk39qlnc",
          "validator_address": "cosmosvaloper1msz843gguwhqx804cdc97n22c4lllfkk5352lt",
          "pubkey": "<consensus_pubkey>",
          "value": {
            "denom": "uatom",
            "amount": "100000000000"
          }
        }
      }
    ],
    "fee": {
      "amount": null,
      "gas": "200000"
    },
    "signatures": [
      {
        "pub_key": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AlT62zuYGlZGUG3Yv0RtIFoPTzVY4N+WEFmBvz1syjws"
        },
        "signature": ""
      }
    ],
    "memo": ""
  }
}
```

__**NOTE**__: If you would like to override the memo field use the `--ip` and `--node-id` flags for the `gaiad gentx` command above.

Finally, to participate in this ceremony, copy this file to the `gentx` folder in this repo
and submit a pull request:

```
cp ~/.gaiad/config/gentx/gentx-<node_id>.json ./gentx/<moniker>.json
```

We will only accept self delegation transactions up to 100,000 atoms for genesis. We expect 1-5% of the ATOM allocation to
be staked via genesis transactions.

## A Note about your Validator Signing Key

Your validator signing private key lives at `~/.gaiad/config/priv_validator_key.json`. If this key is stolen, an attacker would be able to make
your validator double sign, causing a slash of 5% of your atoms and the [tombstoning](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/slashing/07_tombstone.md) of your validator. If you are interested in how to better protect this key please see the [`tendermint/kms`](https://github.com/tendermint/kms) (_*use at your own risk*_) repo. We will have a complete guide for how to secure this file soon after launch.

## Next Steps

Wait for the Interchain Foundation (ICF) to publish a final recommendation for the
Genesis Block Release Software and be ready to come online at the recommended
time.

The ICF will recommend a particular genesis file and software version, but there
is no guarantee a network will ever start from it - nodes and validators may
never come online, the community may disregard the recommendation and choose
different genesis files, and/or they may modify the software in arbitrary ways. Such
outcomes and many more are outside the ICF's control and completely in the hands
of the community

On initialization of the software, the Cosmos Hub Bonded Proof-of-Stake system will kick in to
determine the initial validator set (max 100 validators) from the set of `gentx` transactions.
More than 2/3 of the voting power of this set must be online and participating in consensus
in order to create the first block and start the Cosmos Hub.

We expect and hope that ATOM holders will exercise discretion in initial staking to ensure the network
does not ever become excessively centralized as we move steadily to the target of 66% ATOMs staked. This is
a first of its kind experiment in bootstrapping a decentralized network. Other proof of stake networks have
bootstrapped with the aid of a foundation or other administrator. We hope to bootstrap as a decentralized community, building on the shared experiences of many many testnets.

See the [blog
post](https://blog.cosmos.network/the-3-phases-of-the-cosmos-hub-mainnet-fdff3a68c4c0) 
for more details on the three phases of launch.


# Disclaimer


The Cosmos Hub is *highly* experimental software. In these early days, we can
expect to have issues, updates, and bugs. The existing tools require advanced
technical skills and involve risks which are outside of the control of the
Interchain Foundation and/or the Tendermint team (see also the risk section in
the Interchain Cosmos Contribution Terms). Any use of this open source Apache
2.0 licensed software is done at your *own risk and on a “AS IS” basis, without
warranties or conditions of any kind*, and any and all liability of the
Interchain Foundation and/or the Tendermint team for damages arising in
connection to the software is excluded. **Please exercise extreme caution!**

Furthermore, it must be noted that it remains in the community's discretion to adopt or not
to adopt the Genesis State that Interchain Foundation (ICF) recommends within the Genesis Block
Software. Therefore, ICF *cannot* guarantee that (i) ATOMs will be created and
(ii) the recommended allocation as set forth herein will actually take place.
