package v15_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/mock"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gaia/v15/app/helpers"
	v15 "github.com/cosmos/gaia/v15/app/upgrades/v15"
)

type EmptyAppOptions struct{}

func (ao EmptyAppOptions) Get(_ string) interface{} {
	return nil
}

func TestV15UpgradeHandler(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	// set min commission rate to 0
	stakingParams := gaiaApp.StakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.ZeroDec()
	gaiaApp.StakingKeeper.SetParams(ctx, stakingParams)

	// confirm all commissions are 0
	stakingKeeper := gaiaApp.StakingKeeper

	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		require.Equal(t, val.Commission.CommissionRates.Rate, sdk.ZeroDec(), "non-zero previous commission rate for validator %s", val.GetOperator())
	}

	// pre-test min commission rate is 0
	require.Equal(t, stakingKeeper.GetParams(ctx).MinCommissionRate, sdk.ZeroDec(), "non-zero previous min commission rate")

	// run the test and confirm the values have been updated
	v15.V15UpgradeHandler(ctx, &gaiaApp.AppKeepers)

	newStakingParams := gaiaApp.StakingKeeper.GetParams(ctx)
	require.NotEqual(t, newStakingParams.MinCommissionRate, sdk.ZeroDec(), "failed to update min commission rate")
	require.Equal(t, newStakingParams.MinCommissionRate, sdk.NewDecWithPrec(5, 2), "failed to update min commission rate")

	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		require.Equal(t, val.Commission.CommissionRates.Rate, newStakingParams.MinCommissionRate, "failed to update update commission rate for validator %s", val.GetOperator())
	}

	// set one of the validators commission rate to 10% and ensure it is not updated
	updateValCommission := sdk.NewDecWithPrec(10, 2)
	updateVal := stakingKeeper.GetAllValidators(ctx)[0]
	updateVal.Commission.CommissionRates.Rate = updateValCommission
	stakingKeeper.SetValidator(ctx, updateVal)

	v15.V15UpgradeHandler(ctx, &gaiaApp.AppKeepers)
	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		if updateVal.OperatorAddress == val.OperatorAddress {
			require.Equal(t, val.Commission.CommissionRates.Rate, updateValCommission, "should not update commission rate for validator %s", val.GetOperator())
		} else {
			require.Equal(t, val.Commission.CommissionRates.Rate, newStakingParams.MinCommissionRate, "failed to update update commission rate for validator %s", val.GetOperator())
		}
	}
}

func TestValidatorSigningInfosFix(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	slashingKeeper := gaiaApp.SlashingKeeper

	signingInfosNum := 8
	emptyAddrCtr := 0

	// create some dummy signing infos, half of which with an empty address field
	for i := 0; i < signingInfosNum; i++ {
		pubKey, err := mock.NewPV().GetPubKey()
		require.NoError(t, err)

		consAddr := sdk.ConsAddress(pubKey.Address())
		info := slashingtypes.NewValidatorSigningInfo(
			consAddr,
			0,
			0,
			time.Unix(0, 0),
			false,
			0,
		)

		if i <= signingInfosNum/2 {
			info.Address = ""
			emptyAddrCtr++
		}

		slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info)
		require.NoError(t, err)
	}

	// check signing info were correctly created
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		if info.Address == "" {
			emptyAddrCtr--
		}

		return false
	})
	require.Zero(t, emptyAddrCtr)

	// upgrade signing infos
	v15.UpgradeValidatorSigningInfos(ctx, slashingKeeper)

	// check that all signing info have the address field correctly updated
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
		require.NotEmpty(t, info.Address)
		require.Equal(t, address.String(), info.Address)

		return false
	})
}
