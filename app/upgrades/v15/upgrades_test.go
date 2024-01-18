package v15_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v15/app/helpers"
	v15 "github.com/cosmos/gaia/v15/app/upgrades/v15"
)

func TestUpgradeMinCommissionRate(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	// set min commission rate to 0
	stakingParams := gaiaApp.StakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.ZeroDec()
	err := gaiaApp.StakingKeeper.SetParams(ctx, stakingParams)
	require.NoError(t, err)

	stakingKeeper := gaiaApp.StakingKeeper
	valNum := len(stakingKeeper.GetAllValidators(ctx))

	// create 3 new validators
	for i := 0; i < 3; i++ {
		pk := ed25519.GenPrivKeyFromSecret([]byte{uint8(i)}).PubKey()
		val, err := stakingtypes.NewValidator(
			sdk.ValAddress(pk.Address()),
			pk,
			stakingtypes.Description{},
		)
		require.NoError(t, err)
		// set random commission rate
		val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(tmrand.Int63n(100), 2)
		stakingKeeper.SetValidator(ctx, val)
		valNum++
	}

	validators := stakingKeeper.GetAllValidators(ctx)
	require.Equal(t, valNum, len(validators))

	// pre-test min commission rate is 0
	require.Equal(t, stakingKeeper.GetParams(ctx).MinCommissionRate, sdk.ZeroDec(), "non-zero previous min commission rate")

	// run the test and confirm the values have been updated
	v15.UpgradeMinCommissionRate(ctx, *stakingKeeper)

	newStakingParams := stakingKeeper.GetParams(ctx)
	require.NotEqual(t, newStakingParams.MinCommissionRate, sdk.ZeroDec(), "failed to update min commission rate")
	require.Equal(t, newStakingParams.MinCommissionRate, sdk.NewDecWithPrec(5, 2), "failed to update min commission rate")

	for _, val := range stakingKeeper.GetAllValidators(ctx) {
		require.True(t, val.Commission.CommissionRates.Rate.GTE(newStakingParams.MinCommissionRate), "failed to update update commission rate for validator %s", val.GetOperator())
	}
}

func TestUpgradeSigningInfos(t *testing.T) {
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
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		if info.Address == "" {
			emptyAddrCtr--
		}

		return false
	})
	require.Zero(t, emptyAddrCtr)

	// upgrade signing infos
	v15.UpgradeSigningInfos(ctx, slashingKeeper)

	// check that all signing info have the address field correctly updated
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		require.NotEmpty(t, info.Address)
		require.Equal(t, address.String(), info.Address)

		return false
	})
}
