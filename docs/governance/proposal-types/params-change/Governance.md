# `gov` subspace
The `gov` module is responsible for on-chain governance proposals and voting functionality.

<table>
    <tr>
        <th>Key</th>
        <th>Value</th>
    </tr>
    <tr v-for="(v,k) in $themeConfig.currentParameters.gov">
        <td><a :href="'#'+k"><code>{{ k }}</code></a></td>
        <td><code>{{ v }}</code></td>
    </tr>
</table>

The `gov` module is responsible for the on-chain governance system. In this system, holders of the native staking token of the chain may vote on proposals on a 1-token per 1-vote basis. The module supports:

- **Proposal submission**: Users can submit proposals with a deposit. Once the minimum deposit is reached, proposal enters voting period
- **Vote**: Participants can vote on proposals that reached MinDeposit
- **Inheritance and penalties**: Delegators inherit their validator's vote if they don't vote themselves.
- **Claiming deposit**: Users that deposited on proposals can recover their deposits if the proposal was accepted OR if the proposal never entered voting period.

## Governance notes on parameters

### `depositparams`

#### `min_deposit`
**The minimum deposit required for a proposal to enter the [voting period](#votingperiod), in micro-ATOMs.**

* on-chain value: `{{ $themeConfig.currentParameters.gov.depositparams.min_deposit }}`
* [Proposal 47](https://www.mintscan.io/cosmos/proposals/47) change: `64000000` `uatom`
* `cosmoshub-4` default: `512000000` `uatom`
* `cosmoshub-3` default: `512000000` `uatom`

Prior to a governance proposal entering the [voting period](#votingperiod) (ie. for the proposal to be voted upon), there must be at least a minimum number of ATOMs deposited. Anyone may contribute to this deposit. Deposits of passed and failed proposals are returned to the contributors. Deposits are burned when proposals 1) [expire](#maxdepositperiod), 2) fail to reach [quorum](#quorum), or 3) are [vetoed](#veto). This parameter subkey value represents the minimum deposit required for a proposal to enter the [voting period](#votingperiod) in micro-ATOMs, where `512000000uatom` is equivalent to 512 ATOM.

##### Decreasing the value of `mindeposit`
Decreasing the value of the `mindeposit` subkey will enable governance proposals to enter the [voting period](#votingperiod) with fewer ATOMs at risk. This will likely increase the volume of new governance proposals.

##### Increasing the value of `mindeposit`
Increasing the value of the `mindeposit` subkey will require risking a greater number of ATOMs before governance proposals may enter the [voting period](#votingperiod). This will likely decrease the volume of new governance proposals.

#### `max_deposit_period`
**The maximum amount of time that a proposal can accept deposit contributions before expiring, in nanoseconds.**

* on-chain value: `{{ $themeConfig.currentParameters.gov.depositparams.max_deposit_period}}`
* `cosmoshub-4` default: `1209600000000000`
* `cosmoshub-3` default: `1209600000000000`

Prior to a governance proposal entering the [voting period](#votingperiod), there must be at least a minimum number of ATOMs deposited. This parameter subkey value represents the maximum amount of time that the proposal has to reach the minimum deposit amount before expiring. The maximum amount of time that a proposal can accept deposit contributions before expiring is currently `1209600000000000` nanoseconds or 14 days. If the proposal expires, any deposit amounts will be burned.

##### Decreasing the value of `maxdepositperiod`
Decreasing the value of the `maxdepositperiod` subkey will decrease the time for deposit contributions to governance proposals. This will likely decrease the time that some proposals remain visible and potentially decrease the likelihood that they will enter the [voting period](#votingperiod). This may increase the likelihood that proposals will expire and have their deposits burned.

##### Increasing the value of `maxdepositperiod`
Increasing the value of the `maxdepositperiod` subkey will extend the time for deposit contributions to governance proposals. This will likely increase the time that some proposals remain visible and potentially increase the likelihood that they will enter the [voting period](#votingperiod). This may decrease the likelihood that proposals will expire and have their deposits burned.

##### Notes
Currently most network explorers (eg. Hubble, Big Dipper, Mintscan) give the same visibility to proposals in the deposit period as those in the [voting period](#votingperiod). This means that a proposal with a small deposit (eg. 0.001 ATOM) will have the same visibility as those with a full 512 ATOM deposit in the voting period.

### `votingparams`
#### `votingperiod`
**The maximum amount of time that a proposal can accept votes before the voting period concludes, in nanoseconds.**

* on-chain value: `{{ $themeConfig.currentParameters.gov.votingparams.voting_period}}`
* `cosmoshub-4` default: `1209600000000000`
* `cosmoshub-3` default: `1209600000000000`

Once a governance proposal enters the voting period, there is a maximum period of time that may elapse before the voting period concludes. This parameter subkey value represents the maximum amount of time that the proposal has to accept votes, which is currently `1209600000000000` nanoseconds or 14 days. If the proposal vote does not reach quorum ((ie. 40% of the network's voting power is participating) before this time, any deposit amounts will be burned and the proposal's outcome will not be considered to be valid. Voters may change their vote any number of times before the voting period ends. This voting period is currently the same for any kind of governance proposal.

##### Decreasing the value of `votingperiod`
Decreasing the value of the `votingperiod` subkey will decrease the time for voting on governance proposals. This will likely:
1. decrease the proportion of the network that participates in voting, and
2. decrease the likelihood that quorum will be reached. 

##### Increasing the value of `votingperiod`
Increasing the value of the `votingperiod` subkey will increase the time for voting on governance proposals. This may:
1. increase the proportion of the network that participates in voting, and
2. increase the likelihood that quorum will be reached. 

##### Notes
Historically, off-chain discussions and engagement appears to be have been greater occurred during the voting period of a governance proposal than when the proposal is posted off-chain as a draft. A non-trivial amount of the voting power has voted in the second week of the voting period. Proposals 23, 19, and 13 each had approximately 80% network participation or more.

### `tallyparams`
#### `quorum`
**The minimum proportion of network voting power required for a governance proposal's outcome to be considered valid.**

* on-chain value: `{{ $themeConfig.currentParameters.gov.tallyparams.quorum}}`
* `cosmoshub-4` default: `0.400000000000000000`
* `cosmoshub-3` default: `0.400000000000000000`

Quorum is required for the outcome of a governance proposal vote to be considered valid and for deposit contributors to recover their deposit amounts, and this parameter subkey value represents the minimum value for quorum. Voting power, whether backing a vote of 'yes', 'abstain', 'no', or 'no-with-veto', counts toward quorum. If the proposal vote does not reach quorum (ie. 40% of the network's voting power is participating) before this time, any deposit amounts will be burned and the proposal outcome will not be considered to be valid.

##### Decreasing the value of `quorum`
Decreasing the value of the `quorum` subkey will enable a smaller proportion of the network to legitimize the outcome of a proposal. This increases the risk that a decision will be made with a smaller proportion of ATOM-stakers' positions being represented, while decreasing the risk that a proposal will be considered invalid. This will likely decrease the risk of a proposal's deposit being burned. 

##### Increasing the value of `quorum`
Increasing the value of the `quorum` subkey will require a larger proportion of the network to legitimize the outcome of a proposal. This decreases the risk that a decision will be made with a smaller proportion of ATOM-stakers' positions being represented, while increasing the risk that a proposal will be considered invalid. This will likely increase the risk of a proposal's deposit being burned.

#### `threshold`
**The minimum proportion of participating voting power required for a governance proposal to pass.**

* on-chain value: `{{ $themeConfig.currentParameters.gov.tallyparams.threshold}}`
* `cosmoshub-4` default: `0.500000000000000000`
* `cosmoshub-3` default: `0.500000000000000000`

A simple majority 'yes' vote (ie. 50% of participating voting power) is required for a governance proposal vote to pass. Though necessary, a simple majority 'yes' vote may not be sufficient to pass a proposal in two scenarios:
1. Failure to reach [quorum](#quorum) of 40% network power or
2. A 'no-with-veto' vote of 33.4% of participating voting power or greater.

If a governance proposal passes, deposit amounts are returned to contributors. If a text-based proposal passes, nothing is enacted automatically, but there is a social expectation that participants will co-ordinate to enact the commitments signalled in the proposal. If a parameter change proposal passes, the protocol parameter will automatically change immediately after the [voting period](#votingperiod) ends, and without the need to run new software. If a community-spend proposal passes, the Community Pool balance will decrease by the number of ATOMs indicated in the proposal and the recipient's address will increase by this same number of ATOMs immediately after the voting period ends.

##### Decreasing the value of `threshold`
Decreasing the value of the `threshold` subkey will decrease the proportion of voting power required to pass a proposal. This may:
1. increase the likelihood that a proposal will pass, and
2. increase the likelihood that a minority group will effect changes to the network.

##### Increasing the value of `threshold`
Increasing the value of the `threshold` subkey will increase the proportion of voting power required to pass a proposal. This may:
1. decrease the likelihood that a proposal will pass, and
2. decrease the likelihood that a minority group will effect changes to the network.

#### `veto_threshold`
**The minimum proportion of participating voting power to veto (ie. fail) a governance proposal.**

* on-chain value: `{{ $themeConfig.currentParameters.gov.tallyparams.veto_threshold}}`
* `cosmoshub-4` default: `0.334000000000000000`
* `cosmoshub-3` default: `0.334000000000000000`

Though a simple majority 'yes' vote (ie. 50% of participating voting power) is required for a governance proposal vote to pass, a 'no-with-veto' vote of 33.4% of participating voting power or greater can override this outcome and cause the proposal to fail. This enables a minority group representing greater than 1/3 of voting power to fail a proposal that would otherwise pass.

##### Decreasing the value of `veto_threshold`
Decreasing the value of the `veto_threshold` subkey will decrease the proportion of participating voting power required to veto. This will likely:
1. enable a smaller minority group to prevent proposals from passing, and
2. decrease the likelihood that contentious proposals will pass. 


##### Increasing the value of `veto_threshold`
Increasing the value of the `veto_threshold` subkey will increase the proportion of participating voting power required to veto. This will require a larger minority group to prevent proposals from passing, and will likely increase the likelihood that contentious proposals will pass.