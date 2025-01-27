package apptesting

import (
	"fmt"
	"time"

	"github.com/strangelove-ventures/tokenfactory/app"
	appparams "github.com/strangelove-ventures/tokenfactory/app/params"
	"github.com/stretchr/testify/suite"

	"github.com/cometbft/cometbft/crypto/ed25519"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakinghelper "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type KeeperTestHelper struct {
	suite.Suite

	App         *app.TokenFactoryApp
	Ctx         sdk.Context
	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress

	StakingHelper *stakinghelper.Helper
}

// Setup sets up basic environment for suite (App, Ctx, and test accounts)
func (s *KeeperTestHelper) Setup() {
	t := s.T()
	s.Ctx, s.App = app.Setup(t)

	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)

	s.StakingHelper = stakinghelper.NewHelper(s.Suite.T(), s.Ctx, s.App.StakingKeeper)
	s.StakingHelper.Denom = "stake"
}

func (s *KeeperTestHelper) SetupTestForInitGenesis() {
	t := s.T()
	// Setting to True, leads to init genesis not running
	s.Ctx, s.App = app.Setup(t)
	s.Ctx = s.App.BaseApp.NewContext(true)
	s.Ctx = s.Ctx.WithBlockHeader(tmtypes.Header{
		ChainID: "testing",
	})
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) CreateTestContext() sdk.Context {
	ctx, _ := s.CreateTestContextWithMultiStore()
	return ctx
}

// CreateTestContextWithMultiStore creates a test context and returns it together with multi store.
func (s *KeeperTestHelper) CreateTestContextWithMultiStore() (sdk.Context, storetypes.CommitMultiStore) {
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	ms := rootmulti.NewStore(db, logger, metrics.NewNoOpMetrics())

	return sdk.NewContext(ms, tmtypes.Header{}, false, logger), ms
}

// CreateTestContext creates a test context.
func (s *KeeperTestHelper) Commit() {
	oldHeight := s.Ctx.BlockHeight()
	oldHeader := s.Ctx.BlockHeader()
	_, err := s.App.Commit()
	s.Require().NoError(err)
	newHeader := tmtypes.Header{Height: oldHeight + 1, ChainID: "testing", Time: oldHeader.Time.Add(time.Second)}
	s.Ctx = s.App.NewContext(false)
	s.Ctx = s.Ctx.WithBlockHeader(newHeader)
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := banktestutil.FundAccount(s.Ctx, s.App.BankKeeper, acc, amounts)
	s.Require().NoError(err)
}

// FundModuleAcc funds target modules with specified amount.
func (s *KeeperTestHelper) FundModuleAcc(moduleName string, amounts sdk.Coins) {
	err := banktestutil.FundModuleAccount(s.Ctx, s.App.BankKeeper, moduleName, amounts)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) MintCoins(coins sdk.Coins) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, coins)
	s.Require().NoError(err)
}

// SetupValidator sets up a validator and returns the ValAddress.
func (s *KeeperTestHelper) SetupValidator(bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	valPriv := secp256k1.GenPrivKey()
	valPub := valPriv.PubKey()
	valAddr := sdk.ValAddress(valPub.Address())
	p, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	bondDenom := p.BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: math.NewInt(100), Denom: bondDenom})

	s.FundAcc(sdk.AccAddress(valAddr), selfBond)

	msg := s.StakingHelper.CreateValidatorMsg(valAddr, valPub, selfBond[0].Amount)
	res, err := s.StakingHelper.CreateValidatorWithMsg(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(res)

	val, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	val = val.UpdateStatus(bondStatus)
	s.Require().NoError(s.App.StakingKeeper.SetValidator(s.Ctx, val))

	consAddr, err := val.GetConsAddr()
	s.Suite.Require().NoError(err)

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		s.Ctx.BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	s.Require().NoError(s.App.SlashingKeeper.SetValidatorSigningInfo(s.Ctx, consAddr, signingInfo))

	return valAddr
}

// BeginNewBlock starts a new block.
func (s *KeeperTestHelper) BeginNewBlock() {
	var valAddr []byte

	validators, err := s.App.StakingKeeper.GetAllValidators(s.Ctx)
	s.Require().NoError(err)

	if len(validators) >= 1 {
		valAddrFancy, err := validators[0].GetConsAddr()
		s.Require().NoError(err)
		valAddr = valAddrFancy
	} else {
		valAddrFancy := s.SetupValidator(stakingtypes.Bonded)
		validator, _ := s.App.StakingKeeper.GetValidator(s.Ctx, valAddrFancy)
		valAddr2, _ := validator.GetConsAddr()
		valAddr = valAddr2
	}

	s.BeginNewBlockWithProposer(valAddr)
}

// BeginNewBlockWithProposer begins a new block with a proposer.
func (s *KeeperTestHelper) BeginNewBlockWithProposer(proposer sdk.ValAddress) {
	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, proposer)
	s.Assert().NoError(err)

	valConsAddr, err := validator.GetConsAddr()
	s.Require().NoError(err)

	valAddr := valConsAddr

	newBlockTime := s.Ctx.BlockTime().Add(5 * time.Second)

	s.Ctx = s.Ctx.WithBlockTime(newBlockTime).WithBlockHeight(s.Ctx.BlockHeight() + 1).WithProposer(valAddr)

	fmt.Println("beginning block ", s.Ctx.BlockHeight())
	_, err = s.App.BeginBlocker(s.Ctx)
	s.Require().NoError(err)
}

// EndBlock ends the block.
func (s *KeeperTestHelper) EndBlock() {
	_, err := s.App.EndBlocker(s.Ctx)
	s.Require().NoError(err)
}

// AllocateRewardsToValidator allocates reward tokens to a distribution module then allocates rewards to the validator address.
func (s *KeeperTestHelper) AllocateRewardsToValidator(valAddr sdk.ValAddress, rewardAmt math.Int) {
	validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
	s.Require().NoError(err)

	// allocate reward tokens to distribution module
	coins := sdk.Coins{sdk.NewCoin(appparams.BondDenom, rewardAmt)}
	err = banktestutil.FundModuleAccount(s.Ctx, s.App.BankKeeper, distrtypes.ModuleName, coins)
	s.Require().NoError(err)

	// allocate rewards to validator
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	decTokens := sdk.DecCoins{{Denom: appparams.BondDenom, Amount: math.LegacyNewDec(20000)}}
	s.Require().NoError(s.App.DistrKeeper.AllocateTokensToValidator(s.Ctx, validator, decTokens))
}

// BuildTx builds a transaction.
func (s *KeeperTestHelper) BuildTx(
	txBuilder client.TxBuilder,
	msgs []sdk.Msg,
	sigV2 signing.SignatureV2,
	memo string, txFee sdk.Coins,
	gasLimit uint64,
) authsigning.Tx {
	err := txBuilder.SetMsgs(msgs[0])
	s.Require().NoError(err)

	err = txBuilder.SetSignatures(sigV2)
	s.Require().NoError(err)

	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(txFee)
	txBuilder.SetGasLimit(gasLimit)

	return txBuilder.GetTx()
}

func (s *KeeperTestHelper) ConfirmUpgradeSucceeded(upgradeName string, upgradeHeight int64) {
	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight - 1)
	plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, err = s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight)
	s.Require().NotPanics(func() {
		_, err := s.App.BeginBlocker(s.Ctx)
		s.Require().NoError(err)
	})
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

func GenerateTestAddrs() (string, string) {
	pk1 := ed25519.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid").String()
	return validAddr, invalidAddr
}
