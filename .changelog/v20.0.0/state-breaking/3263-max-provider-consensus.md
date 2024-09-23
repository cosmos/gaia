- Set the `MaxProviderConsensusValidators` parameter of the provider module to 180. 
  This parameter will be used to govern the number of validators participating in consensus,
  and takes over this role from the `MaxValidators` parameter of the staking module.
  ([\#3263](https://github.com/cosmos/gaia/pull/3263))