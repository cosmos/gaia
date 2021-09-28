# Formatting a Params Change Proposal

**Note:** Changes to the [`gov` module](/Governance.md) are different from the other kinds of parameter changes because `gov` has subkeys, [as discussed here](https://github.com/cosmos/cosmos-sdk/issues/5800). Only the `key` part of the JSON file is different for `gov` parameter-change proposals.

For parameter-change proposals, there are seven (7) components:
1. **Title** - the distinguishing name of the proposal, typically the way the that explorers list proposals
2. **Description** - the body of the proposal that further describes what is being proposed and details surrounding the proposal
3. **Subspace** - the Cosmos Hub module with the parameter that is being changed
4. **Key** - the parameter that will be changed
5. **Value** - the value of the parameter that will be changed by the governance mechanism
6. **Denom** - `uatom` (micro-ATOM) will be the type of asset used as the deposit
7. **Amount** - the amount that will be contributed to the deposit (in micro-ATOMs "uatom") from the account submitting the proposal

### Examples

In this simple example ([below](#testnet-example)), a network explorer will list the governance proposal by its title: "Increase the minimum deposit amount for governance proposals." When a user selects the proposal, they'll see the proposal’s description. A nearly identical proposal [can be found on the gaia-13007 testnet here](https://hubble.figment.network/cosmos/chains/gaia-13007/governance/proposals/30).

Not all explorers will show the proposed parameter changes that are coded into the proposal, so ensure that you verify that the description aligns with what the governance proposal is programmed to enact. If the description says that a certain parameter will be increased, it should also be programmed to do that, but it's possible that that's not the case (accidentally or otherwise).

You can query the proposal details with the gaiad command-line interface using this command: `gaiad q gov proposal 30 --chain-id gaia-13007 --node 45.77.218.219:26657`

You use can also use [Hubble](https://hubble.figment.network/cosmos/chains/gaia-13007/transactions/B5AB56719ADB7117445F6E191E1FCE775135832AFE6C9922B8703AADBC4B13F3?format=json) or gaiad to query the transaction that I sent to create a similar proposal on-chain in full detail: `gaiad q tx B5AB56719ADB7117445F6E191E1FCE775135832AFE6C9922B8703AADBC4B13F3 --chain-id gaia-13007 --node 45.77.218.219:26657`

#### Testnet Example: changing a parameter from the `gov` module
```
{
  "title": "Increase the minimum deposit amount for governance proposals",
  "description": "If successful, this parameter-change governance proposal that will change the minimum deposit from 0.1 to 0.2 testnet ATOMs.",
  "changes": [
    {
      "subspace": "gov",
      "key": "depositparams",
      "value": {"mindeposit":"200000umuon"}
    }
  ],
  "deposit": "100000umuon"
}
```
The deposit `denom` is `uatom` and `amount` is `100000`. Since 1,000,000 micro-ATOM is equal to 1 ATOM, a deposit of 0.1 ATOM will be included with this proposal. The gaia-13007 testnet currently has a 0.1 ATOM minimum deposit, so this will put the proposal directly into the voting period. There is a minimum deposit required for a proposal to enter the voting period, and anyone may contribute to this deposit within a 14-day period. If the minimum deposit isn't reached before this time, the deposit amounts will be burned. Deposit amounts will also be burned if quorum isn't met in the vote or if the proposal is vetoed.

### Mainnet example: 
To date, the Cosmos Hub's parameters have not been changed by a parameter-change governance proposal. This is a hypothetical example of the JSON file that would be used with a command line transaction to create a new proposal. This is an example of a proposal that changes two parameters, and both parameters are from the [`slashing` module](Slashing.md). A single parameter-change governance proposal can reportedly change any number of parameters.

```
{
  "title": "Parameter changes for validator downtime",
  "description": "If passed, this governance proposal will do two things:\n\n1. Increase the slashing penalty for downtime from 0.01% to 0.50%\n2. Decrease the window \n\nIf this proposal passes, validators must sign at least 5% of 5,000 blocks, which is 250 blocks. That means that a validator that misses 4,750 consecutive blocks will be considered by the system to have committed a liveness violation, where previously 9,500 consecutive blocks would need to have been missed to violate these system rules. Assuming 7s block times, validators offline for approximately 9.25 consecutive hours (instead of ~18.5 hours) will be slashed 0.5% (instead of 0.01%).",
  "changes": [
    {
      "subspace": "slashing",
      "key": "SlashFractionDowntime",
      "value": 0.005000000000000000
    }
{
      "subspace": "slashing",
      "key": "SignedBlocksWindow",
      "value": 5000
    }
  ],
  "deposit": "512000000uatom"
}
```
**Note:** in the JSON file, `\n` creates a new line.

It's worth noting that this example proposal doesn't provide reasoning/justification for these changes. Consider consulting the [parameter-change best practices documentation](submitting.md) for guidance on the contents of a parameter-change proposal.

