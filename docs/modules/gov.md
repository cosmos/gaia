# Gov Module

**The `gov` module enables token holders to participate in on-chain governance.**

## Overview

Users can submit and vote on the following types of governance proposals:

1. Software Upgrade Proposal
2. Signaling Proposal
3. Param Change Proposal
4. Community Spend Proposal
5. IBC Client Upgrade Proposal

See the following resources for additional information on using the `gov` module.
- [Cosmos SDK Gov Module Docs](https://docs.cosmos.network/v0.46/modules/gov/)
- [Hub Gov Tutorial]()

#### Forum

The **[Cosmos Hub Forum](https://forum.cosmos.network/)** is the main discussion space for Cosmos Hub Governance. Community members can discuss, share thoughts, and solicit feedback for ideas, drafts and upcoming or active proposals. For more information on active Cosmos Hub governance, visit the [proposal directory](https://forum.cosmos.network/c/hub-proposals/25), or the [forum participation guide](https://forum.cosmos.network/t/start-here-participating-in-the-forum/5993) to get started.

#### Submission Updates

Additionally the new gov module now accommodates proposals with multiple arbitrary message execution. For `v0.46` of the Cosmos SDK, only `MsgSend`, `MsgSoftwareUpgrade`, and `MsgCancelUpgrade` are supported by executing `submit-proposal`. Signaling proposals can also be submitted with the new api by leaving the messages field empty and populating the metadata field with the corresponding proposal text. The remaining proposal types must use `submit-legacy-proposal` until more message types are supported.

This means that a single proposal could execute multiple bank send messages, or both a software upgrade and bank send message.

**Note:** It is only possible to execute a MsgSend transaction if the `gov` module account holds a token balance. Theoretically there could be a community spend proposal to fund the `gov` module account, or assign staking rewards to the account as well.

#### Voting

Once a proposal has been submitted and the token deposit has reached the `min_deposit` amount, the voting period begins. Users can vote with a single option like `yes`, `no`, `no_with_veto`, and `abstain`. Users can also split their vote with `weighted-vote`. This means a token holder could submit a weighted vote for a proposal with 0.5 of their vote counting as `yes`, 0.1 `no`, and 0.4 `abstain`.

## Governance Lifecycle

**Network Parameters**

The network parameters define the time and threshold constraints of the governance lifecycle on the Cosmos Hub.

- `voting_period` - Denominated in nanoseconds; the amount of time from the end of the deposit period that token holders have to vote on an active proposal.
- `tally_params` - Vote tally Parameters
  - `quorum` - Minimum percentage of total voting power that must be cast in order for a proposal to be valid.
  - `threshold` - Minimum percentage of `yes` out of the total votes for the proposal to pass.
  - `veto_threshold` - Minimum percentage of total votes cast for the proposal to be vetoed and fail.
- `deposit_params` - Deposit Parameters
  - `min_deposit` - Minimum denominated token amount deposited to a proposal in order for it to move to voting.
  - `max_deposit_period` - Maximum amount of time denominated in nanoseconds that a proposal can accept deposits.


Run `gaiad q gov params` to retrieve the current network parameters for the Cosmos Hub

```json
{
  "voting_params": {
    "voting_period": "1209600000000000"
  },
  "tally_params": {
    "quorum": "0.400000000000000000",
    "threshold": "0.500000000000000000",
    "veto_threshold": "0.334000000000000000"
  },
  "deposit_params": {
    "min_deposit": [
      {
        "denom": "uatom",
        "amount": "64000000"
      }
    ],
    "max_deposit_period": "1209600000000000"
  }
}
```

**1 - Proposal Submission**

After discussion on the [Cosmos Hub Forum](https://forum.cosmos.network/c/hub-proposals/25), a proposal is ready for submission. Users can submit their proposal with a partial or full deposit.

**2 - Deposit Period**

The current deposit period duration is 2 weeks. A proposal must have passed the `min_deposit` threshold of 64 atoms to be voted on.

**2 - Voting Period**

The current voting period duration is 2 weeks. A proposal must have met a 40% quorum of total voting power and 50% yes votes out of the total number of votes to pass. Conversely, a proposal can fail if it surpasses a 33% veto out of total number of votes. 

## Usage

### New Proposals
To submit a new `gov` proposal, expect to run `gaiad tx gov submit-proposal [path/to/proposal.json] [flags]`. See below for a few examples of different proposal json definitions:

**`MsgSend` Proposal**
```json
{
  "messages": [
    {
      "@type": "/cosmos.bank.v1beta1.MsgSend",
      "from_address": "cosmos10....",
      "to_address": "cosmos1w...",
      "amount":[{"denom": "uatom","amount": "100"}]
    }
  ],
  "metadata": "VGVzdGluZyAxLCAyLCAzIQ==",
  "deposit": "5000uatom"
}
```

**`SoftwareUpgrade` Proposal**
> Note: The authority in this proposal must be the gov module address.

```json
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
      "plan": {
        "name": "rho",
        "height": 20,
        "info": "More information..."
      }
    }
  ],
  "metadata": "VGVzdGluZyAxLCAyLCAzIQ==",
  "deposit": "5000uatom"
}
```

**`CancelSoftwareUpgrade` Proposal**

```json
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
      "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
    }
  ],
  "metadata": "VGVzdGluZyAxLCAyLCAzIQ==",
  "deposit": "5000uatom"
}
```

**Signaling Proposal**

A simple signaling/text proposal can use the new `submit-proposal` by populating the metadata field with the appropriate text and leaving the `messages` collection empty.

```json
{
  "messages": [],
  "metadata": "Proposal Text Link",
  "deposit": "1000000uatom"
}
```

### Legacy Proposals
For the remaining proposal types, it is necessary to use the `legacy-proposal` api. The available sub-commands are:
```
  cancel-software-upgrade Cancel the current software upgrade proposal
  community-pool-spend    Submit a community pool spend proposal
  ibc-upgrade             Submit an IBC upgrade proposal
  param-change            Submit a parameter change proposal
  software-upgrade        Submit a software upgrade proposal
  update-client           Submit an update IBC client proposal
```

It is possible to submit legacy proposals with each field in the command. However, it is also possible to pass a json file with the proposal field definitions instead. See the following legacy proposal cli api definitions:

**Community Spend Proposal**

`gaiad tx gov submit-legacy-proposal community-pool-spend "path/to/proposal.json" --from signing_key [flags]`
```json
{
  "title": "Community Pool Spend",
  "description": "Fund Gov Module Account",
  "recipient": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
  "amount": "10000000uatom",
  "deposit": "1000000uatom"
}
```

**Param Change**

`gaiad tx gov submit-legacy-proposal param-change "path/to/proposal.json" --from signing_key [flags]`

```json
{
  "title": "Staking Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ],
  "deposit": "1000000uatom"
}
```

**IBC Client Update**

`gaiad tx gov submit-legacy-proposal update-client [subject-client-id] [substitute-client-id] [flags]`

**IBC Upgrade**
See the [IBC Upgrade via Gov Proposal Docs](https://ibc.cosmos.network/main/ibc/proposals.html) for more a more comprehensive guide.

` gaiad tx gov submit-legacy-proposal ibc-upgrade [name] [height] [path/to/upgraded_client_state.json] [flags]`

### Queries

Current CLI Queries expose the following information. For more information on using the query cli api visit the [SDK Gov Module Docs](https://docs.cosmos.network/v0.46/modules/gov/07_client.html#query)
```
  deposit     Query details of a deposit
  deposits    Query deposits on a proposal
  param       Query the parameters (voting|tallying|deposit) of the governance process
  params      Query the parameters of the governance process
  proposal    Query details of a single proposal
  proposals   Query proposals with optional filters
  proposer    Query the proposer of a governance proposal
  tally       Get the tally of a proposal vote
  vote        Query details of a single vote
  votes       Query votes on a proposal
```

### HTTP & gRPC
Queries via the gRPC and REST endpoints are available if enabled in `app.toml`. See the [REST](https://docs.cosmos.network/v0.46/modules/gov/07_client.html#rest) and [gRPC](https://docs.cosmos.network/v0.46/modules/gov/07_client.html#grpc) specs in the SDK docs about both api definitions.
