package v21

import (
	"context"
	"fmt"

	providerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/provider/keeper"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govparams "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/cosmos/gaia/v23/app/keepers"
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

		ctx.Logger().Info("allocating rewards of Neutron and Stride unaccounted denoms")
		err = AllocateNeutronAndStrideUnaccountedDenoms(ctx, keepers.ProviderKeeper, keepers.BankKeeper, keepers.AccountKeeper)
		if err != nil {
			// migration can only work on cosmoshub-4
			// all testchains except for mainnet export fork will fail this
			ctx.Logger().Error("Error allocating rewards of Neutron and Stride unaccounted denoms:", "message", err.Error())
		}

		err = InitializeConstitutionCollection(ctx, *keepers.GovKeeper)
		if err != nil {
			ctx.Logger().Error("Error initializing Constitution Collection:", "message", err.Error())
		}

		err = InitializeGovParams(ctx, *keepers.GovKeeper)
		if err != nil {
			ctx.Logger().Error("Error initializing Gov Params:", "message", err.Error())
		}

		ctx.Logger().Info("Upgrade v21 complete")
		return vm, nil
	}
}

// AllocateRewards allocates all the `denoms` that reside in the  `address` and are meant for the chain with `consumerID`
func AllocateRewards(ctx sdk.Context, providerKeeper providerkeeper.Keeper, bankKeeper bankkeeper.Keeper, address sdk.AccAddress, consumerID string, denoms []string) error {
	for _, denom := range denoms {
		coinRewards := bankKeeper.GetBalance(ctx, address, denom)
		decCoinRewards := sdk.DecCoins{sdk.DecCoin{Denom: coinRewards.Denom, Amount: math.LegacyNewDecFromInt(coinRewards.Amount)}}
		consumerRewardsAllocation := providertypes.ConsumerRewardsAllocation{Rewards: decCoinRewards}

		err := providerKeeper.SetConsumerRewardsAllocationByDenom(ctx, consumerID, denom, consumerRewardsAllocation)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasexpectedChainIDSanityCheck returns true if the chain with the provided `consumerID` is of a chain with the `expectedChainID`
func HasExpectedChainIDSanityCheck(ctx sdk.Context, providerKeeper providerkeeper.Keeper, consumerID string, expectedChainID string) bool {
	actualChainID, err := providerKeeper.GetConsumerChainId(ctx, consumerID)
	if err != nil {
		return false
	}
	if expectedChainID != actualChainID {
		return false
	}
	return true
}

// AllocateNeutronAndStrideUnaccountedDenoms allocates previously unaccounted denoms to the Stride and Neutron consumer chains
func AllocateNeutronAndStrideUnaccountedDenoms(ctx sdk.Context, providerKeeper providerkeeper.Keeper, bankKeeper bankkeeper.Keeper, accountKeeper authkeeper.AccountKeeper) error {
	consumerRewardsPoolAddress := accountKeeper.GetModuleAccount(ctx, providertypes.ConsumerRewardsPool).GetAddress()

	const NeutronconsumerID = "0"
	const NeutronChainID = "neutron-1"

	if !HasExpectedChainIDSanityCheck(ctx, providerKeeper, NeutronconsumerID, NeutronChainID) {
		return fmt.Errorf("failed sanity check: consumer id (%s) does not correspond to chain id (%s)", NeutronconsumerID, NeutronChainID)
	}

	neutronUnaccountedDenoms := []string{NeutronUusdc, NeutronUtia}
	err := AllocateRewards(ctx, providerKeeper, bankKeeper, consumerRewardsPoolAddress, NeutronconsumerID, neutronUnaccountedDenoms)
	if err != nil {
		return fmt.Errorf("cannot allocate rewards for consumer id (%s): %w", NeutronconsumerID, err)
	}

	const StrideconsumerID = "1"
	const StrideChainID = "stride-1"

	if !HasExpectedChainIDSanityCheck(ctx, providerKeeper, StrideconsumerID, StrideChainID) {
		return fmt.Errorf("failed sanity check: consumer id (%s) does not correspond to chain id (%s)", StrideconsumerID, StrideChainID)
	}

	strideUnaccountedDenoms := []string{StrideStutia, StrideStadym, StrideStaISLM, StrideStuband, StrideStadydx, StrideStusaga}
	err = AllocateRewards(ctx, providerKeeper, bankKeeper, consumerRewardsPoolAddress, StrideconsumerID, strideUnaccountedDenoms)
	if err != nil {
		return fmt.Errorf("cannot allocate rewards for consumer id (%s): %w", StrideconsumerID, err)
	}

	return nil
}

// setting the default constitution for the chain
// this is in line with cosmos-sdk v5 gov migration: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/gov/migrations/v5/store.go#L57
func InitializeConstitutionCollection(ctx sdk.Context, govKeeper govkeeper.Keeper) error {
	return govKeeper.Constitution.Set(ctx, "This chain has no constitution.")
}

func InitializeGovParams(ctx sdk.Context, govKeeper govkeeper.Keeper) error {
	params, err := govKeeper.Params.Get(ctx)
	if err != nil {
		return err
	}

	params.ProposalCancelRatio = govparams.DefaultProposalCancelRatio.String()
	params.ProposalCancelDest = govparams.DefaultProposalCancelDestAddress

	return govKeeper.Params.Set(ctx, params)
}
