package v21

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/gaia/v20/app/keepers"
	providerkeeper "github.com/cosmos/interchain-security/v6/x/ccv/provider/keeper"
	types2 "github.com/cosmos/interchain-security/v6/x/ccv/provider/types"
)

// Neutron and Stride denoms that were not whitelisted but the consumer rewards pool contains amounts of those denoms.
// Price in $ for each denom corresponds to an approximation fo the current amount stored in the consumer rewards pool
// as of 27.09.2024. Only denoms with amounts more than $10 are included.
const (
	NeutronUusdc = "ibc/4E0D0854C0F846150FA8389D75EA5B5129B17703D7F4992D0356B4FE7C013D42" // ~$40
	NeutronUtia  = "ibc/7054742D02E4F28B7DB5B44D97A496CF5AD16C2AE6948028A5FD57DCE7C5E271" // ~$300

	StrideStutia  = "ibc/17DABEBAC71C388DA064A3D54FB7E68BAF0687965EC39DEADA1FB78C0F1447E6" // ~$18,000
	StrideStadym  = "ibc/3F0A41ECB6FAF27E315583DBF39B5B69A7149D23959A0E4B319F7EF5C618DCD7" // ~$800
	StrideStaISLM = "ibc/61A6F21D6AFF9835F66056461F1CAE24AA3323820259856B485FE7C063CA4FA6" // ~$1650
	StrideStuband = "ibc/E9401AC885592AC2023E0FB9BA7C8BC66D346CEE04CED8E9F545F3C25290708A" // ~$300
	StrideStadydx = "ibc/EEFD952A6DE346F2649039E99A16430B05FFEDF628A4DE99F34BB4B5F6A9346E" // ~$21,000
	StrideStusaga = "ibc/F918765AC289257B35DECC52BD92EBCDBA3C139658BD6F2670D70A6E10B97F58" // ~$300
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v21.
// It performs module migrations, as well as the following tasks:
// - Initializes the MaxProviderConsensusValidators parameter in the provider module to 180.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		ctx.Logger().Info("distributing rewards of Neutron and Stride unaccounted denoms")
		err = DistributeNeutronAndStrideUnaccountedDenoms(ctx, keepers.ProviderKeeper, keepers.BankKeeper, keepers.AccountKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "could not distribute rewards of Neutron and Stride unaccounted denoms")
		}

		ctx.Logger().Info("Upgrade v21 complete")
		return vm, nil
	}
}

// DistributeDenoms distributes all the `denoms` that reside in the  `address` and are meant for the chain with `consumerId`
func DistributeDenoms(ctx sdk.Context, providerKeeper providerkeeper.Keeper, bankKeeper bankkeeper.Keeper, address sdk.AccAddress, consumerId string, denoms []string) error {
	for _, denom := range denoms {
		coinRewards := bankKeeper.GetBalance(ctx, address, denom)
		decCoinRewards := sdk.DecCoins{sdk.DecCoin{Denom: coinRewards.Denom, Amount: math.LegacyNewDecFromInt(coinRewards.Amount)}}
		consumerRewardsAllocation := types2.ConsumerRewardsAllocation{Rewards: decCoinRewards}

		isDenomAllowlisted := providerKeeper.ConsumerRewardDenomExists(ctx, denom)

		// allowlist the denom that distribution can take place
		providerKeeper.SetConsumerRewardDenom(ctx, denom)

		err := providerKeeper.SetConsumerRewardsAllocationByDenom(ctx, consumerId, denom, consumerRewardsAllocation)
		if err != nil {
			return err
		}

		// call `BeginBlockRD` to actually perform the distribution
		providerKeeper.BeginBlockRD(ctx)

		// if you were not allowlisted before, revert to initial state
		if !isDenomAllowlisted {
			providerKeeper.DeleteConsumerRewardDenom(ctx, denom)
		}
	}
	return nil
}

// HasExpectedChainIdSanityCheck return true if the chain with the provided `consumerId` is of a chain with the `expectedChainId`
func HasExpectedChainIdSanityCheck(ctx sdk.Context, providerKeeper providerkeeper.Keeper, consumerId string, expectedChainId string) bool {
	actualChainId, err := providerKeeper.GetConsumerChainId(ctx, consumerId)
	if err != nil {
		return false
	}
	if expectedChainId != actualChainId {
		return false
	}
	return true
}

// DistributeNeutronAndStrideUnaccountedDenoms distributed previously unaccounted denoms to the Stride and Neutron consumer chains
func DistributeNeutronAndStrideUnaccountedDenoms(ctx sdk.Context, providerKeeper providerkeeper.Keeper, bankKeeper bankkeeper.Keeper, accountKeeper authkeeper.AccountKeeper) error {
	consumerRewardsPoolAddress := accountKeeper.GetModuleAccount(ctx, types2.ConsumerRewardsPool).GetAddress()

	const NeutronConsumerId = "0"
	const NeutronChainId = "neutron-1"

	if !HasExpectedChainIdSanityCheck(ctx, providerKeeper, NeutronConsumerId, NeutronChainId) {
		return fmt.Errorf("failed sanity check: consumer id (%s) does not correspond to chain id (%s)", NeutronConsumerId, NeutronChainId)
	}

	neutronUnaccountedDenoms := []string{NeutronUusdc, NeutronUtia}
	err := DistributeDenoms(ctx, providerKeeper, bankKeeper, consumerRewardsPoolAddress, NeutronConsumerId, neutronUnaccountedDenoms)
	if err != nil {
		return fmt.Errorf("cannot distribute rewards for consumer id (%s): %w", NeutronConsumerId, err)
	}

	const StrideConsumerId = "1"
	const StrideChainId = "stride-1"

	if !HasExpectedChainIdSanityCheck(ctx, providerKeeper, StrideConsumerId, StrideChainId) {
		return fmt.Errorf("failed sanity check: consumer id (%s) does not correspond to chain id (%s)", StrideConsumerId, StrideChainId)
	}

	strideUnaccountedDenoms := []string{StrideStutia, StrideStadym, StrideStaISLM, StrideStuband, StrideStadydx, StrideStusaga}
	err = DistributeDenoms(ctx, providerKeeper, bankKeeper, consumerRewardsPoolAddress, StrideConsumerId, strideUnaccountedDenoms)
	if err != nil {
		return fmt.Errorf("cannot distribute rewards for consumer id (%s): %w", StrideConsumerId, err)
	}

	return nil
}
