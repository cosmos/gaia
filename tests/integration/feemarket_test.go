package integration

import (
	"testing"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/gaia/v23/ante"
	gaiaApp "github.com/cosmos/gaia/v23/app"
)

const (
	LargeMsgNumber = 1000
	LargeFeeAmount = 1000000000
	LargeGasLimit  = simtestutil.DefaultGenTxGas * 10
)

type FeeMarketTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator
	chain       *ibctesting.TestChain
	app         *gaiaApp.GaiaApp
}

func TestFeeMarketTestSuite(t *testing.T) {
	feemarketSuite := &FeeMarketTestSuite{}
	suite.Run(t, feemarketSuite)
}

func (suite *FeeMarketTestSuite) SetupTest() {
	ante.UseFeeMarketDecorator = true
	ibctesting.DefaultTestingAppInit = GaiaAppIniter
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1)
	OverrideSendMsgs(suite.coordinator.Chains, sdk.NewInt64Coin(sdk.DefaultBondDenom, LargeFeeAmount), LargeGasLimit)

	chain, ok := suite.coordinator.Chains[ibctesting.GetChainID(1)]
	suite.Require().True(ok, "chain not found")
	suite.chain = chain
	suite.chain.ProposedHeader.ProposerAddress = sdk.ConsAddress(suite.chain.Vals.Validators[0].Address)

	app, ok := chain.App.(*gaiaApp.GaiaApp)
	suite.Require().True(ok, "expected App to be GaiaApp")
	suite.app = app
}

func (suite *FeeMarketTestSuite) TestBaseFeeAdjustment() {
	// BaseFee is initially set to DefaultMinBaseGasPrice
	ctx := suite.chain.GetContext()

	baseFee, err := suite.app.FeeMarketKeeper.GetBaseGasPrice(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(feemarkettypes.DefaultMinBaseGasPrice, baseFee)

	// BaseFee can not be lower than DefaultMinBaseGasPrice, even after N empty blocks
	suite.coordinator.CommitNBlocks(suite.chain, 10)

	ctx = suite.chain.GetContext()
	baseFee, err = suite.app.FeeMarketKeeper.GetBaseGasPrice(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(feemarkettypes.DefaultMinBaseGasPrice, baseFee)

	// Send a large transaction to consume a lot of gas
	sender := suite.chain.SenderAccounts[0].SenderAccount.GetAddress()
	receiver := suite.chain.SenderAccounts[1].SenderAccount.GetAddress()
	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10)))

	msgs := make([]sdk.Msg, LargeMsgNumber)
	for i := 0; i < LargeMsgNumber; i++ {
		bankSendMsg := banktypes.NewMsgSend(sender, receiver, amount)
		msgs[i] = bankSendMsg
	}

	_, err = suite.chain.SendMsgs(msgs...)
	suite.Require().NoError(err)

	// Check that BaseFee has increased due to the large gas usage
	ctx = suite.chain.GetContext()
	baseFee, err = suite.app.FeeMarketKeeper.GetBaseGasPrice(ctx)
	suite.Require().NoError(err)
	suite.Require().True(baseFee.GT(feemarkettypes.DefaultMinBaseGasPrice))

	// BaseFee should drop to DefaultMinBaseGasPrice after N empty blocks
	suite.coordinator.CommitNBlocks(suite.chain, 10)

	ctx = suite.chain.GetContext()
	baseFee, err = suite.app.FeeMarketKeeper.GetBaseGasPrice(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(feemarkettypes.DefaultMinBaseGasPrice, baseFee)
}
