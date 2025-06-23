package amiavalidator

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v25/telemetry"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	storetypes "cosmossdk.io/store/types"
)

func TestPreBlock(t *testing.T) {
	moniker := "test"
	pk := simtestutil.CreateTestPubKeys(1)[0]

	// Create validator
	v, err := stakingtypes.NewValidator(
		sdk.ValAddress(pk.Address()).String(),
		pk,
		stakingtypes.Description{Moniker: moniker},
	)
	require.NoError(t, err)

	addr, err := v.GetConsAddr()
	require.NoError(t, err)

	// Setup telemetry client
	valInfo := telemetry.ValidatorInfo{
		ChainID:     "",
		IsValidator: false,
		Moniker:     moniker,
		Address:     addr,
	}

	tests := []struct {
		name                string
		validatorStatus     stakingtypes.BondStatus
		IsValidatorInitial  bool
		expectedIsValidator bool
		blockHeight         int64
	}{
		{
			name:                "bonded validator sets IsValidator to true",
			validatorStatus:     stakingtypes.Bonded,
			IsValidatorInitial:  false,
			expectedIsValidator: true,
			blockHeight:         20,
		},
		{
			name:                "unbonded validator sets IsValidator to false",
			validatorStatus:     stakingtypes.Unbonded,
			IsValidatorInitial:  true,
			expectedIsValidator: false,
			blockHeight:         20,
		},
		{
			name:                "not updated if %20 != 0",
			validatorStatus:     stakingtypes.Unbonded,
			IsValidatorInitial:  true,
			expectedIsValidator: true,
			blockHeight:         15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valInfo.IsValidator = tt.IsValidatorInitial
			oc := telemetry.NewOtelClient(telemetry.OtelConfig{}, valInfo)

			// Setup staking keeper and module
			sk, ctx := setupStakingKeeper(t)
			q := stakingkeeper.Querier{Keeper: sk}
			mod := NewAppModule(&q, oc)
			ctx = ctx.WithBlockHeight(tt.blockHeight)
			v.Status = tt.validatorStatus
			require.NoError(t, sk.SetValidatorByConsAddr(ctx, v))
			require.NoError(t, sk.SetValidator(ctx, v))

			_, err := mod.PreBlock(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.expectedIsValidator, oc.IsValidator())
		})
	}
}

func setupStakingKeeper(t *testing.T) (*stakingkeeper.Keeper, sdk.Context) {
	t.Helper()

	var (
		bondedAcc    = authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName)
		notBondedAcc = authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName)
	)
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(t)
	accountKeeper := stakingtestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := stakingtestutil.NewMockBankKeeper(ctrl)

	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
	)
	require.NoError(t, keeper.SetParams(ctx, stakingtypes.DefaultParams()))
	return keeper, ctx
}
