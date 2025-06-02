package v24_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gaia "github.com/cosmos/gaia/v25/app"
	"github.com/cosmos/gaia/v25/app/helpers"
	"github.com/cosmos/gaia/v25/app/upgrades/v24"
	"github.com/cosmos/gaia/v25/x/liquid/types"
)

func TestMigrateLSMState(t *testing.T) {
	type testCase struct {
		name        string
		setup       func(sdk.Context, *gaia.GaiaApp)
		expectError bool
		validate    func(*testing.T, sdk.Context, *gaia.GaiaApp)
	}

	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")

	newContext := func(t *testing.T) (*gaia.GaiaApp, sdk.Context) {
		t.Helper()

		app := helpers.Setup(t)
		ctx := app.NewUncachedContext(true, tmproto.Header{Time: time.Now()})
		return app, ctx
	}

	setupParams := func(app *gaia.GaiaApp, ctx sdk.Context) {
		params, err := app.StakingKeeper.GetParams(ctx)
		require.NoError(t, err)
		params.GlobalLiquidStakingCap = math.LegacyNewDec(1000)
		params.ValidatorLiquidStakingCap = math.LegacyNewDec(100)
		require.NoError(t, app.StakingKeeper.SetParams(ctx, params))
	}

	addValidatorAndDelegation := func(t *testing.T, app *gaia.GaiaApp, ctx sdk.Context, valAddr sdk.ValAddress, moduleAddr sdk.AccAddress) {
		t.Helper()

		val := stakingtypes.Validator{
			OperatorAddress: valAddr.String(),
			Tokens:          math.NewInt(1_000_000),
			DelegatorShares: math.LegacyNewDec(1_000_000),
		}
		require.NoError(t, app.StakingKeeper.SetValidator(ctx, val))

		del := stakingtypes.Delegation{
			DelegatorAddress: moduleAddr.String(),
			ValidatorAddress: valAddr.String(),
			Shares:           math.LegacyNewDec(1_000_000),
		}
		require.NoError(t, app.StakingKeeper.SetDelegation(ctx, del))
	}

	cases := []testCase{
		{
			name: "single record + lock",
			setup: func(ctx sdk.Context, app *gaia.GaiaApp) {
				setupParams(app, ctx)

				valAddr := sdk.ValAddress(addr1)
				record := types.TokenizeShareRecord{
					Id:            1,
					Owner:         sdk.MustBech32ifyAddressBytes("cosmos", addr2),
					ModuleAccount: "cosmos1modacct",
					Validator:     valAddr.String(),
				}
				require.NoError(t, app.StakingKeeper.AddTokenizeShareRecord(ctx, stakingtypes.TokenizeShareRecord(record)))
				addValidatorAndDelegation(t, app, ctx, valAddr, record.GetModuleAddress())

				app.StakingKeeper.SetLastTokenizeShareRecordID(ctx, 1)
				app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, math.NewInt(12345))

				unlock := time.Now()
				app.StakingKeeper.AddTokenizeSharesLock(ctx, addr1)
				app.StakingKeeper.SetTokenizeSharesUnlockTime(ctx, addr1, unlock)
			},
			expectError: false,
			validate: func(t *testing.T, ctx sdk.Context, app *gaia.GaiaApp) {
				t.Helper()

				records := app.LiquidKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 1)

				status, _ := app.LiquidKeeper.GetTokenizeSharesLock(ctx, addr1)
				require.Equal(t, types.TOKENIZE_SHARE_LOCK_STATUS_LOCK_EXPIRING, status)
			},
		},
		{
			name: "missing delegation",
			setup: func(ctx sdk.Context, app *gaia.GaiaApp) {
				valAddr := sdk.ValAddress(addr1)
				val := stakingtypes.Validator{
					OperatorAddress: valAddr.String(),
					Tokens:          math.NewInt(1_000_000),
				}
				require.NoError(t, app.StakingKeeper.SetValidator(ctx, val))

				record := stakingtypes.TokenizeShareRecord{
					Id:            1,
					Owner:         sdk.MustBech32ifyAddressBytes("cosmos", addr2),
					ModuleAccount: "cosmos1modacct",
					Validator:     valAddr.String(),
				}
				require.NoError(t, app.StakingKeeper.AddTokenizeShareRecord(ctx, record))
			},
			expectError: true,
		},
		{
			name: "invalid validator address",
			setup: func(ctx sdk.Context, app *gaia.GaiaApp) {
				record := stakingtypes.TokenizeShareRecord{
					Id:            1,
					Owner:         sdk.MustBech32ifyAddressBytes("cosmos", addr1),
					ModuleAccount: "cosmos1modacct",
					Validator:     "not-a-bech32-address",
				}
				require.NoError(t, app.StakingKeeper.AddTokenizeShareRecord(ctx, record))
			},
			expectError: true,
		},
		{
			name: "0 liquid shares migrated",
			setup: func(ctx sdk.Context, app *gaia.GaiaApp) {
				setupParams(app, ctx)

				valAddr1 := sdk.ValAddress(addr1)
				val1 := stakingtypes.Validator{
					OperatorAddress: valAddr1.String(),
					Tokens:          math.NewInt(1_000_000),
				}
				require.NoError(t, app.StakingKeeper.SetValidator(ctx, val1))
				valAddr2 := sdk.ValAddress(addr2)
				val2 := stakingtypes.Validator{
					OperatorAddress: valAddr2.String(),
					Tokens:          math.NewInt(1_000_000),
				}
				require.NoError(t, app.StakingKeeper.SetValidator(ctx, val2))

				app.StakingKeeper.SetLastTokenizeShareRecordID(ctx, 1)
				app.StakingKeeper.SetTotalLiquidStakedTokens(ctx, math.NewInt(12345))
			},
			validate: func(t *testing.T, ctx sdk.Context, app *gaia.GaiaApp) {
				t.Helper()

				lv1, err := app.LiquidKeeper.GetLiquidValidator(ctx, sdk.ValAddress(addr1))
				require.NoError(t, err)
				require.True(t, math.LegacyZeroDec().Equal(lv1.LiquidShares))

				lv2, err := app.LiquidKeeper.GetLiquidValidator(ctx, sdk.ValAddress(addr2))
				require.NoError(t, err)
				require.True(t, math.LegacyZeroDec().Equal(lv2.LiquidShares))
			},
			expectError: false,
		},
		{
			name: "empty state",
			setup: func(ctx sdk.Context, app *gaia.GaiaApp) {
				// No-op
			},
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, ctx := newContext(t)
			tc.setup(ctx, app)

			err := v24.MigrateLSMState(ctx, &app.AppKeepers)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, ctx, app)
				}
			}
		})
	}
}
