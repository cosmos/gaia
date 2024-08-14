- Set the `MaxValidators` parameter of the staking module to 200, which is the current number of 180 plus 20.
  This is done as a result of introducing the inactive-validators feature of Interchain Security, 
  which entails that the number of validators participating in consensus will be governed by the 
  `MaxProviderConsensusValidators` parameter in the provider module.
  ([\#3263](https://github.com/cosmos/gaia/pull/3263))