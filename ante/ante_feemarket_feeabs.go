package ante

import (
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DenomResolverImpl is Gaia's implementation of x/feemarket's DenomResolver
type DenomResolverImpl struct {
	FeeabsKeeper  feeabskeeper.Keeper
	StakingKeeper feeabstypes.StakingKeeper
}

var _ feemarkettypes.DenomResolver = &DenomResolverImpl{}

// ConvertToDenom converts any given coin to the native denom of the chain or the other way around.
// Return error if neither of coin.Denom and denom is the native denom of the chain.
// If the denom is the bond denom, convert `coin` to the native denom. return error if coin.Denom is not in the allowed list
// If the denom is not the bond denom, convert the `coin` to the given denom. return error if denom is not in the allowed list
func (r *DenomResolverImpl) ConvertToDenom(ctx sdk.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	bondDenom := r.StakingKeeper.BondDenom(ctx)
	if denom != bondDenom && coin.Denom != bondDenom {
		return sdk.DecCoin{}, ErrNeitherNativeDenom(coin.Denom, denom)
	}
	var amount sdk.Coins
	var hostZoneConfig feeabstypes.HostChainFeeAbsConfig
	var found bool
	var err error

	if denom == bondDenom {
		hostZoneConfig, found = r.FeeabsKeeper.GetHostZoneConfig(ctx, coin.Denom)
		if !found {
			return sdk.DecCoin{}, ErrDenomNotRegistered(coin.Denom)
		}
		amount, err = r.getIBCCoinFromNative(ctx, sdk.NewCoins(sdk.NewCoin(coin.
			Denom, coin.Amount.TruncateInt())), hostZoneConfig)
	} else if coin.Denom == bondDenom {
		hostZoneConfig, found := r.FeeabsKeeper.GetHostZoneConfig(ctx, denom)
		if !found {
			return sdk.DecCoin{}, ErrDenomNotRegistered(denom)
		}
		amount, err = r.FeeabsKeeper.CalculateNativeFromIBCCoins(ctx, sdk.NewCoins(sdk.NewCoin(denom, coin.Amount.TruncateInt())), hostZoneConfig)
	}

	if err != nil {
		return sdk.DecCoin{}, err
	}
	return sdk.NewDecCoinFromDec(denom, amount[0].Amount.ToLegacyDec()), nil
}

// ExtraDenoms should be all denoms that have been registered via governance(host zone)
func (r *DenomResolverImpl) ExtraDenoms(ctx sdk.Context) ([]string, error) {
	allHostZoneConfigs, err := r.FeeabsKeeper.GetAllHostZoneConfig(ctx)
	if err != nil {
		return nil, err
	}
	denoms := make([]string, 0, len(allHostZoneConfigs))
	for _, hostZoneConfig := range allHostZoneConfigs {
		denoms = append(denoms, hostZoneConfig.IbcDenom)
	}
	return denoms, nil
}

// //////////////////////////////////////
// Helper functions for DenomResolver //
// //////////////////////////////////////

func (r *DenomResolverImpl) getIBCCoinFromNative(ctx sdk.Context, nativeCoins sdk.Coins, chainConfig feeabstypes.HostChainFeeAbsConfig) (coins sdk.Coins, err error) {
	if len(nativeCoins) != 1 {
		return sdk.Coins{}, ErrExpectedOneCoin(len(nativeCoins))
	}

	nativeCoin := nativeCoins[0]

	twapRate, err := r.FeeabsKeeper.GetTwapRate(ctx, chainConfig.IbcDenom)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Divide native amount by twap rate to get IBC amount
	ibcAmount := nativeCoin.Amount.ToLegacyDec().Quo(twapRate).RoundInt()
	ibcCoin := sdk.NewCoin(chainConfig.IbcDenom, ibcAmount)

	// Verify the resulting IBC coin
	err = r.verifyIBCCoins(ctx, sdk.NewCoins(ibcCoin))
	if err != nil {
		return sdk.Coins{}, err
	}

	return sdk.NewCoins(ibcCoin), nil
}

// return err if IBC token isn't in allowed_list
func (r *DenomResolverImpl) verifyIBCCoins(ctx sdk.Context, ibcCoins sdk.Coins) error {
	if ibcCoins.Len() != 1 {
		return feeabstypes.ErrInvalidIBCFees
	}

	ibcDenom := ibcCoins[0].Denom
	if r.FeeabsKeeper.HasHostZoneConfig(ctx, ibcDenom) {
		return nil
	}
	return feeabstypes.ErrUnsupportedDenom.Wrapf("unsupported denom: %s", ibcDenom)
}
