package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	lsmkeeper "github.com/cosmos/gaia/v23/x/lsm/keeper"
	lsmtypes "github.com/cosmos/gaia/v23/x/lsm/types"
	"github.com/cosmos/gaia/v23/x/lsm/types/mocks"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	lsmKeeper     *lsmkeeper.Keeper
	stakingKeeper *mocks.StakingKeeper
	bankKeeper    *mocks.BankKeeper
	accountKeeper *mocks.AccountKeeper
	queryClient   lsmtypes.QueryClient
	msgServer     lsmtypes.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	require := s.Require()
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	accountKeeper := mocks.NewAccountKeeper(s.T())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos"))

	bankKeeper := mocks.NewBankKeeper(s.T())
	stakingKeeper := mocks.NewStakingKeeper(s.T())
	distributionKeeper := mocks.NewDistributionKeeper(s.T())

	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).Maybe()

	lsmKeeper := lsmkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distributionKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	require.NoError(lsmKeeper.SetParams(ctx, lsmtypes.DefaultParams()))

	s.ctx = ctx
	s.stakingKeeper = stakingKeeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.lsmKeeper = lsmKeeper

	lsmtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	lsmtypes.RegisterQueryServer(queryHelper, lsmkeeper.Querier{Keeper: lsmKeeper})
	s.queryClient = lsmtypes.NewQueryClient(queryHelper)
	s.msgServer = lsmkeeper.NewMsgServerImpl(lsmKeeper)
}

func (s *KeeperTestSuite) TestParams() {
	ctx, keeper := s.ctx, s.lsmKeeper
	require := s.Require()

	expParams := lsmtypes.DefaultParams()
	// check that the empty keeper loads the default
	resParams, err := keeper.GetParams(ctx)
	require.NoError(err)
	require.Equal(expParams, resParams)

	expParams.ValidatorBondFactor = sdkmath.LegacyNewDec(-1)
	expParams.GlobalLiquidStakingCap = sdkmath.LegacyNewDec(1)
	expParams.ValidatorLiquidStakingCap = sdkmath.LegacyNewDec(1)
	require.NoError(keeper.SetParams(ctx, expParams))
	resParams, err = keeper.GetParams(ctx)
	require.NoError(err)
	require.True(expParams.Equal(resParams))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
