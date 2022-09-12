# Group Module

**The `group` module facilitates creating and managing multisig accounts and enables members to configure group policies that can submit and vote on specific message execution proposals.**

## Overview
A group is a set of accounts and associated weights. Group accounts may then be instantiated by creating a new decision policy. Each decision policy can have a different threshold for voting on and passing proposals.

See the following resources for additional information on using the `group` module.

- [Cosmos SDK Group Module Docs](https://docs.cosmos.network/v0.46/modules/group/)
- [Cosmos Developer Portal Group Tutorial](https://tutorials.cosmos.network/tutorials/understanding-group/)

##### Group Administration
Groups can be initialized with or without a decision policy. Groups must have an administrator account, but that account does not necessarily have to be a group member. The group may self-administer by setting the administrator to a group policy account. Upon creating a group, each member is assigned a weight that will correspond to voting power with respect to policy proposals.

The administrator can update group members, replace themselves, and update decision policies.

##### Policies
There are two types of decision policies, threshold and percentage decision policies. A threshold decision policy sets the minimum voting power cast in favor of a proposal for it to pass. A percentage decision policy requires a minimum percentage of voting power cast in favor of a policy in order for it to pass. Policies can be updated by a group admin and are the actual abstraction of multi-signature accounts.

##### Proposals
Proposals are submitted on behalf of an individual decision policy account and voted on. Proposals facilitate multiple arbitrary message execution as long as the decision policy account has the authority to execute them. For example, a decision policy account can't perform a bank send if it doesn't carry a balance.

## Group Lifecycle

**1. Group Creation**

A group is initialized by the administrator. Both options exist to create the group with or without its first decision policy.

**2. Policy Creation**

The group admin may create one or more decision policies. The decision policy acts as the multi-sig account from which proposals that pass are executed. Policies describe the conditions and the voting period of any proposals submitted on their behalf.

**3. Proposal Submission**

Proposals can be submitted by one or more group members and must be signed by the group admin. Proposals may contain one or more messages and must be attached to a specific group policy.

**4. Proposal Voting**

Once a group proposal is submitted, the voting period begins. Group members may vote `VOTE_OPTION_YES`, `VOTE_OPTION_NO`, `VOTE_OPTION_ABSTAIN`, and `VOTE_OPTION_NO_WITH_VETO`.

**5. Proposal Execution**

Even if a proposal passes, its messages cannot be executed automatically because of the gas calculation that needs to be run at the moment of execution.

## Usage

### Administer a Group

#### Group Creation
To create a new group without a decision policy, run `gaiad tx group create-group [admin] [metadata] [members-json-file] [flags]`. The admin account does not also have to be a group member.

> Note: Metadata cannot be omitted but can be left blank in the form of an empty string.

**members.json**
```json
{
  "members": [
    {
      "address": "cosmos1symvu...",
      "weight": "1",
      "metadata": "Bob"
    },
    {
      "address": "cosmos1xzudh...",
      "weight": "1",
      "metadata": "Alice"
    }
  ]
}
```
#### Updating a Group
Members can voluntarily leave a group by executing `gaiad tx group leave-group [member-address] [group-id] [flags]`.

Group Administrators can also remove members from the group by running `gaiad tx group update-group-members [admin] [group-id] [members-json-file] [flags]` and updating the `members.json` by setting the user's weight to `0`.

**members.json**
```json
{
  "members": [
    {
      "address": "cosmos1symvu...",
      "weight": "0",
      "metadata": "Bob"
    },
    {
      "address": "cosmos1xzudh...",
      "weight": "1",
      "metadata": "Alice"
    }
  ]
}
```

### Creating a Decision Policy
To start submitting and voting on proposals, a group must have at least one decision policy.

Run `gaiad tx group create-group-policy [admin] [group-id] [metadata] [decision-policy-json-file] [flags]` to create a group decision policy and see below for examples of both decision policies.

#### Threshold Decision Policy
**policy.json**
```json
{
  "@type": "/cosmos.group.v1.ThresholdDecisionPolicy",
  "threshold": "1",
  "windows": {
    "min_execution_period": "0s",
    "voting_period": "120s"
  }
}
```

#### Percentage Decision Policy
**policy.json**
```json
{
  "@type": "/cosmos.group.v1.PercentageDecisionPolicy",
  "percentage": "0.5",
  "windows": {
    "min_execution_period": "0s",
    "voting_period": "30s"
  }
}
```

### Submitting a Proposal
Once a group has at least one decision policy, members can begin to submit proposals on behalf of a specific policy account. Run `gaiad tx group submit-proposal [proposal_json_file] [flags]` to submit a group proposal, and see below for an example of the `proposal.json`

> Note: Messages in the proposal can only be executed if the group policy account on the proposal has the authority to do so. In the below example, the policy account would need to have sufficient balance to execute the bank send message.

**proposal.json**
```json
{
	"group_policy_address": "cosmos1afk9z...",
	"messages": [
  	{
  		"@type": "/cosmos.bank.v1beta1.MsgSend",
  		"from_address": "cosmos1afk9z...",
  		"to_address": "cosmos1symvu...",
  		"amount":[{"denom": "uatom","amount": "10000000"}]
  	}
	],
	"metadata": "Send 10 atom from Group to Bob",
	"proposers": [
    "cosmos1xzudh..."
  ]
}
```

### Voting & Executing a Proposal
The rules for a proposal's voting process depend on the attached decision policy; this includes `voting_period` and whether a specific percentage or voting weight threshold is necessary to pass a proposal.

Once a proposal has been successfully submitted and the voting period has begun, members can submit their votes by running `gaiad tx group vote [proposal-id] [voter-address] [vote-option] [metadata] [flags]`.

If a proposal passes, then a group member must trigger the proposal execution by running `gaiad tx group exec [proposal-id] [flags]`.

### HTTP & gRPC
Queries via the gRPC and REST endpoints are available if enabled in `app.toml`. See the [REST](https://docs.cosmos.network/v0.46/modules/group/05_client.html#rest) and [gRPC](https://docs.cosmos.network/v0.46/modules/group/05_client.html#grpc) specs in the SDK docs about both api definitions.
