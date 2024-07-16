- Add the [feemarket module](https://github.com/skip-mev/feemarket) and set the initial params to the following values. ([\#3028](https://github.com/cosmos/gaia/pull/3028) and [\#3164](https://github.com/cosmos/gaia/pull/3164))
  ```
  FeeDenom = "uatom"
  DistributeFees = false // burn base fees
  MinBaseGasPrice = 0.005 // same as previously enforced by `x/globalfee`
  MaxBlockUtilization = 30_000_000 // the default value 
  ```
  