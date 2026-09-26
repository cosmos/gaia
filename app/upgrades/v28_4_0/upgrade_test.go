package v28_4_0_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	gaiaapp "github.com/cosmos/gaia/v29/app"
	v284 "github.com/cosmos/gaia/v29/app/upgrades/v28_4_0"
)

const denom = "uatom"

func TestRefundFromCommunityPool(t *testing.T) {
	recipient := sdk.MustAccAddressFromBech32(v284.RefundRecipient)
	enough := int64(v284.RefundAmount + 5)

	tests := []struct {
		name string
		// pool is the community pool accounting, coins is what the distribution module actually holds.
		pool, coins int64
		// existing is the recipient's pre-upgrade balance. Zero means no account.
		existing int64
		// setupAcc overrides the recipient's pre-upgrade account.
		setupAcc  func(app *gaiaapp.GaiaApp, ctx sdk.Context)
		blockTime time.Time
		wantErr   error
	}{
		{name: "existing account keeps spendable balance", pool: enough, coins: enough, existing: 7},
		{name: "no account", pool: enough, coins: enough, wantErr: v284.ErrRecipientNotFound},
		{name: "pool too small", pool: v284.RefundAmount - 1, coins: enough, existing: 7, wantErr: v284.ErrInsufficientPool},
		{name: "pool accounting not backed by coins", pool: enough, coins: v284.RefundAmount - 1, existing: 7, wantErr: v284.ErrInsufficientPool},
		{name: "unlock time already passed", pool: enough, coins: enough, existing: 7, blockTime: time.Unix(v284.RefundUnlockTime, 0), wantErr: v284.ErrUnlockInPast},
		{
			name: "recipient already vesting",
			pool: enough, coins: enough,
			setupAcc: func(app *gaiaapp.GaiaApp, ctx sdk.Context) {
				base := app.AccountKeeper.NewAccountWithAddress(ctx, recipient).(*authtypes.BaseAccount)
				acc, err := vestingtypes.NewDelayedVestingAccount(base, sdk.NewCoins(sdk.NewInt64Coin(denom, 1)), v284.RefundUnlockTime)
				require.NoError(t, err)
				app.AccountKeeper.SetAccount(ctx, acc)
			},
			wantErr: v284.ErrRecipientNotBase,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			appOpts := make(simtestutil.AppOptionsMap)
			appOpts[server.FlagInvCheckPeriod] = 5
			app := gaiaapp.NewGaiaApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, map[int64]bool{}, t.TempDir(), appOpts, []wasmkeeper.Option{})
			ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{Height: 1, Time: tc.blockTime}, false, log.NewNopLogger())

			stakingParams := stakingtypes.DefaultParams()
			stakingParams.BondDenom = denom
			require.NoError(t, app.StakingKeeper.SetParams(ctx, stakingParams))

			// Seed the community pool.
			coins := sdk.NewCoins(sdk.NewInt64Coin(denom, tc.coins))
			require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
			require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, distrtypes.ModuleName, coins))
			pool := sdk.NewDecCoinsFromCoins(sdk.NewInt64Coin(denom, tc.pool))
			require.NoError(t, app.DistrKeeper.FeePool.Set(ctx, distrtypes.FeePool{CommunityPool: pool}))

			if tc.existing > 0 {
				existing := sdk.NewCoins(sdk.NewInt64Coin(denom, tc.existing))
				require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, existing))
				require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, recipient, existing))
			}
			if tc.setupAcc != nil {
				tc.setupAcc(app, ctx)
			}

			mm := module.NewManager()
			configurator := module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
			handler := v284.CreateUpgradeHandler(mm, configurator, &app.AppKeepers)

			_, err := handler(ctx, upgradetypes.Plan{Name: v284.UpgradeName, Height: 1}, module.VersionMap{})
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			refund := sdk.NewInt64Coin(denom, v284.RefundAmount)
			require.Equal(t, refund.Amount.AddRaw(tc.existing), app.BankKeeper.GetBalance(ctx, recipient, denom).Amount)

			acc, ok := app.AccountKeeper.GetAccount(ctx, recipient).(*vestingtypes.DelayedVestingAccount)
			require.True(t, ok, "recipient should be a delayed vesting account")
			require.Equal(t, int64(v284.RefundUnlockTime), acc.EndTime)
			require.Equal(t, math.NewInt(tc.existing), app.BankKeeper.SpendableCoins(ctx, recipient).AmountOf(denom))

			feePool, err := app.DistrKeeper.FeePool.Get(ctx)
			require.NoError(t, err)
			require.Equal(t, math.LegacyNewDec(5), feePool.CommunityPool.AmountOf(denom))

			// Locked refund can't be transferred.
			other := sdk.AccAddress("other_address_______")
			require.Error(t, app.BankKeeper.SendCoins(ctx, recipient, other, sdk.NewCoins(refund)))

			// But it can be staked. This is the bank call x/staking makes on delegate.
			require.NoError(t, app.BankKeeper.DelegateCoinsFromAccountToModule(ctx, recipient, stakingtypes.NotBondedPoolName, sdk.NewCoins(refund)))
		})
	}
}
