---
title: Cosmos SDK LSM
order: 3
---

## Deprecation

As of the `v24.x` release of Gaia, the Cosmos SDK based Liquid Staking Module is deprecated. The `v24` release line 
will still have all the types from the forked SDK version, but all the API endpoints will be disabled.

## Migration

The following outlines the key differences between the Cosmos SDK LSM and the x/liquid module, including message and 
API removals between the two, as well as parameter and limit changes. In general an effort was made to keep as much 
as possible the same between the two. Cases where removals were made were largely due to no longer be able to 
properly track state from outside the staking module.

The state associated with each account currently on a live network which is using the SDK LSM will be migrated to 
use the liquid module. There should be no change to the existing liquid staking denoms, tokenize share records, or 
rewards associated with any owner of a tokenize share record.

### Removals
- MsgUnbondValidator
- MsgValidatorBond

Both the message types and the APIs for Validator Bonds have been removed. The `min_self_delegation` field in the 
SDK's staking module, which had previously been marked as deprecated, has been restored to its upstream state.

Note that if you created a validator after the field was deprecated, you may end up with `0` self delegation and a 
`min_self_delegation` of `1`. This should be minimally invasive and will only affect you if you try to  call 
`EditValidator` or potentially in an `Unjail` scenario. The resolution for this is to self delegate a minimal amount 
of stake.

- Limits changes

The `GlobalLiquidStakingCap` and `ValidatorLiquidStakingCap` are now the only protocol limits and are enforced 
solely on tokenized shares through the x/liquid module itself. Previously this tracked delegations via ICA as 
counting towards these limits, but the new tallies do not take that into account.

Additionally, the `liquid_shares` field of the `Validator` object returned in staking module queries will no longer 
be updated. Its removal is planned in v25, but as of v24 that value will still show up, but will not be updated and 
is not used in limits calculations. To query the same information, you can use the `QueryLiquidValidator` request 
added to the x/liquid module.

### Migrations
All messages and APIs have been moved.

For protobuf, the types within the distribution and staking modules have all been moved to
```protobuf
gaia.liquid.v1beta1.*
```

For API paths, the module-specific paths have all been moved to
```
/gaia/liquid/v1beta1/
```

CLI interactions with the liquid module should use `gaiad q liquid` or `gaiad tx liquid`, etc.
