package v24_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v23/app/helpers"
	"github.com/cosmos/gaia/v23/app/upgrades/v24"
	"github.com/cosmos/gaia/v23/x/liquid/types"
)

var (
	addr1 = sdk.AccAddress("addr1_______________")
	addr2 = sdk.AccAddress("addr2_______________")
)

func TestMigrateLSMState(t *testing.T) {
	t.Run("single tokenize share record and lock", func(t *testing.T) {
		gaiaApp := helpers.Setup(t)
		ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{Time: time.Now()})

		// Params
		stakingParams, err := gaiaApp.StakingKeeper.GetParams(ctx)
		require.NoError(t, err)
		stakingParams.GlobalLiquidStakingCap = math.LegacyNewDec(1000)
		stakingParams.ValidatorLiquidStakingCap = math.LegacyNewDec(100)
		require.NoError(t, gaiaApp.StakingKeeper.SetParams(ctx, stakingParams))

		// Record
		record := types.TokenizeShareRecord{
			Id:            1,
			Owner:         sdk.MustBech32ifyAddressBytes("cosmos", addr2),
			ModuleAccount: "cosmos1modacct",
			Validator:     "cosmosvaloper1xyz",
		}
		require.NoError(t, gaiaApp.StakingKeeper.AddTokenizeShareRecord(ctx, stakingtypes.TokenizeShareRecord(record)))

		gaiaApp.StakingKeeper.SetLastTokenizeShareRecordID(ctx, 1)
		gaiaApp.StakingKeeper.SetTotalLiquidStakedTokens(ctx, math.NewInt(12345))

		// Lock
		unlockTime := time.Now()
		gaiaApp.StakingKeeper.AddTokenizeSharesLock(ctx, addr1)
		gaiaApp.StakingKeeper.SetTokenizeSharesUnlockTime(ctx, addr1, unlockTime)

		// Migrate
		require.NoError(t, v24.MigrateLSMState(ctx, &gaiaApp.AppKeepers))

		// Verify
		params, err := gaiaApp.LiquidKeeper.GetParams(ctx)
		require.NoError(t, err)
		require.Equal(t, stakingParams.GlobalLiquidStakingCap, params.GlobalLiquidStakingCap)

		records := gaiaApp.LiquidKeeper.GetAllTokenizeShareRecords(ctx)
		require.Len(t, records, 1)

		status, unlock := gaiaApp.LiquidKeeper.GetTokenizeSharesLock(ctx, addr1)
		require.Equal(t, types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING, status)
		require.Equal(t, unlockTime.Unix(), unlock.Unix())
	})

	t.Run("multiple records and locks", func(t *testing.T) {
		gaiaApp := helpers.Setup(t)
		ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{Time: time.Now()})

		// Add multiple records
		for i := uint64(1); i <= 3; i++ {
			owner := sdk.MustBech32ifyAddressBytes("cosmos", sdk.AccAddress([]byte(fmt.Sprintf("owner%d_____________", i))))
			record := stakingtypes.TokenizeShareRecord{
				Id:            i,
				Owner:         owner,
				ModuleAccount: fmt.Sprintf("cosmos1modacct%d", i),
				Validator:     fmt.Sprintf("cosmosvaloper1xyz%d", i),
			}
			require.NoError(t, gaiaApp.StakingKeeper.AddTokenizeShareRecord(ctx, record))
		}
		gaiaApp.StakingKeeper.SetLastTokenizeShareRecordID(ctx, 3)
		gaiaApp.StakingKeeper.SetTotalLiquidStakedTokens(ctx, math.NewInt(98765))

		// Locks
		for _, addr := range []sdk.AccAddress{addr1, addr2} {
			gaiaApp.StakingKeeper.AddTokenizeSharesLock(ctx, addr)
			gaiaApp.StakingKeeper.SetTokenizeSharesUnlockTime(ctx, addr, time.Now())
		}

		require.NoError(t, v24.MigrateLSMState(ctx, &gaiaApp.AppKeepers))

		records := gaiaApp.LiquidKeeper.GetAllTokenizeShareRecords(ctx)
		require.Len(t, records, 3)

		for _, addr := range []sdk.AccAddress{addr1, addr2} {
			status, _ := gaiaApp.LiquidKeeper.GetTokenizeSharesLock(ctx, addr)
			require.Equal(t, types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING, status)
		}
	})

	t.Run("empty state should not fail", func(t *testing.T) {
		gaiaApp := helpers.Setup(t)
		ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{Time: time.Now()})

		require.NoError(t, v24.MigrateLSMState(ctx, &gaiaApp.AppKeepers))

		records := gaiaApp.LiquidKeeper.GetAllTokenizeShareRecords(ctx)
		require.Empty(t, records)
	})

	t.Run("double migration", func(t *testing.T) {
		gaiaApp := helpers.Setup(t)
		ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{Time: time.Now()})

		// One record
		record := stakingtypes.TokenizeShareRecord{
			Id:            1,
			Owner:         sdk.MustBech32ifyAddressBytes("cosmos", addr2),
			ModuleAccount: "cosmos1modacct",
			Validator:     "cosmosvaloper1xyz",
		}
		require.NoError(t, gaiaApp.StakingKeeper.AddTokenizeShareRecord(ctx, record))

		// Run twice
		require.NoError(t, v24.MigrateLSMState(ctx, &gaiaApp.AppKeepers))
		require.Error(t, v24.MigrateLSMState(ctx, &gaiaApp.AppKeepers))

		records := gaiaApp.LiquidKeeper.GetAllTokenizeShareRecords(ctx)
		require.Len(t, records, 1)
	})
}
