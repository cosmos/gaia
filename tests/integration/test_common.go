package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtprototypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/tokenfactory/x/tokenfactory"
	tokenfactorykeeper "github.com/cosmos/tokenfactory/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/cosmos/tokenfactory/x/tokenfactory/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v26/x/liquid"
	liquidkeeper "github.com/cosmos/gaia/v26/x/liquid/keeper"
	liquidtypes "github.com/cosmos/gaia/v26/x/liquid/types"
)

type fixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	accountKeeper      authkeeper.AccountKeeper
	bankKeeper         bankkeeper.Keeper
	distributionKeeper distributionkeeper.Keeper
	stakingKeeper      *stakingkeeper.Keeper
	liquidKeeper       *liquidkeeper.Keeper
	tokenFactoryKeeper tokenfactorykeeper.Keeper
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distributiontypes.StoreKey, stakingtypes.StoreKey, liquidtypes.StoreKey, tokenfactorytypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, staking.AppModuleBasic{}, vesting.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, cmtprototypes.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		distributiontypes.ModuleName:   {authtypes.Minter},
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.ModuleName:        {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		tokenfactorytypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()))
	distributionKeeper := distributionkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[distributiontypes.
		StoreKey]), accountKeeper, bankKeeper, stakingKeeper, distributiontypes.ModuleName, authority.String())
	liquidKeeper := liquidkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[liquidtypes.StoreKey]), accountKeeper,
		bankKeeper, stakingKeeper, distributionKeeper, authority.String())

	tokenFactoryKeeper := tokenfactorykeeper.NewKeeper(
		cdc,
		keys[tokenfactorytypes.StoreKey],
		maccPerms,
		accountKeeper,
		bankKeeper,
		distributionKeeper,
		[]string{tokenfactorytypes.EnableSetMetadata, tokenfactorytypes.EnableCommunityPoolFeeFunding},
		authority.String(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	distributionModule := distribution.NewAppModule(cdc, distributionKeeper, accountKeeper, bankKeeper,
		stakingKeeper, nil)
	liquidModule := liquid.NewAppModule(cdc, liquidKeeper, accountKeeper, bankKeeper, stakingKeeper)
	tokenFactoryModule := tokenfactory.NewAppModule(tokenFactoryKeeper, accountKeeper, bankKeeper, nil)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:         authModule,
		banktypes.ModuleName:         bankModule,
		distributiontypes.ModuleName: distributionModule,
		stakingtypes.ModuleName:      stakingModule,
		liquidtypes.ModuleName:       liquidModule,
		tokenfactorytypes.ModuleName: tokenFactoryModule,
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	stakingKeeper.SetHooks(stakingtypes.NewMultiStakingHooks(
		liquidKeeper.Hooks(),
	))

	// Register staking MsgServer and QueryServer
	stakingtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), stakingkeeper.NewMsgServerImpl(stakingKeeper))
	stakingtypes.RegisterQueryServer(integrationApp.QueryHelper(), stakingkeeper.NewQuerier(stakingKeeper))

	// set default staking params
	require.NoError(tb, stakingKeeper.SetParams(sdkCtx, stakingtypes.DefaultParams()))

	// Register liquid MsgServer and QueryServer
	liquidtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), liquidkeeper.NewMsgServerImpl(liquidKeeper))
	liquidtypes.RegisterQueryServer(integrationApp.QueryHelper(), liquidkeeper.NewQuerier(liquidKeeper))

	// set default liquid params
	require.NoError(tb, liquidKeeper.SetParams(sdkCtx, liquidtypes.DefaultParams()))

	// Register tokenfactory MsgServer and QueryServer
	tokenfactorytypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), tokenfactorykeeper.NewMsgServerImpl(tokenFactoryKeeper))
	tokenfactorytypes.RegisterQueryServer(integrationApp.QueryHelper(), tokenFactoryKeeper)

	// Set default tokenfactory params
	err := tokenFactoryKeeper.SetParams(sdkCtx, tokenfactorytypes.DefaultParams())
	require.NoError(tb, err)

	f := fixture{
		app:                integrationApp,
		sdkCtx:             sdkCtx,
		cdc:                cdc,
		keys:               keys,
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		distributionKeeper: distributionKeeper,
		stakingKeeper:      stakingKeeper,
		liquidKeeper:       liquidKeeper,
		tokenFactoryKeeper: tokenFactoryKeeper,
	}

	return &f
}

func delegateCoinsFromAccount(ctx sdk.Context, sk stakingkeeper.Keeper, addr sdk.AccAddress, amount math.Int,
	val stakingtypes.ValidatorI,
) error {
	_, err := sk.Delegate(ctx, addr, amount, stakingtypes.Unbonded, val.(stakingtypes.Validator), true)

	return err
}

func applyValidatorSetUpdates(t *testing.T, ctx sdk.Context, k *stakingkeeper.Keeper,
	expectedUpdatesLen int,
) []abci.ValidatorUpdate {
	t.Helper()
	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	if expectedUpdatesLen >= 0 {
		require.Equal(t, expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}
