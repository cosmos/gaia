package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"

	gaiaApp "github.com/cosmos/gaia/v25/app"
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
	ibctesting.DefaultTestingAppInit = GaiaAppIniter
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1)
	OverrideSendMsgs(suite.coordinator.Chains, sdk.NewInt64Coin(sdk.DefaultBondDenom, LargeFeeAmount), LargeGasLimit)

	chain, ok := suite.coordinator.Chains[ibctesting.GetChainID(1)]
	suite.Require().True(ok, "chain not found")
	suite.chain = chain
	suite.chain.ProposedHeader.ProposerAddress = suite.chain.Vals.Validators[0].Address

	app, ok := chain.App.(*gaiaApp.GaiaApp)
	suite.Require().True(ok, "expected App to be GaiaApp")
	suite.app = app
}

func (suite *FeeMarketTestSuite) TestBaseFeeAdjustment() {
	// BaseFee is initially set to DefaultMinBaseGasPrice
	ctx := suite.chain.GetContext()

	baseFee := suite.app.FeeMarketKeeper.GetBaseFee(ctx)
	suite.Require().True(baseFee.LT(feemarkettypes.DefaultBaseFee))

	// BaseFee can not be lower than DefaultMinBaseGasPrice, even after N empty blocks
	suite.coordinator.CommitNBlocks(suite.chain, 10)

	ctx = suite.chain.GetContext()
	newBaseFee := suite.app.FeeMarketKeeper.GetBaseFee(ctx)
	suite.Require().Truef(newBaseFee.LT(baseFee), "newBaseFee %d prevBaseFee %d", newBaseFee, baseFee)
}
