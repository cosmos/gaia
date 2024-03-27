package ics

import (
	"math/rand"
	"testing"
	"time"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	gaiaApp "github.com/cosmos/gaia/v16/app"
)

const (
	LargeMsgNumber = 1000
	LargeFeeAmount = 1000000000000000
	LargeGasLimit  = simtestutil.DefaultGenTxGas * 100
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
	suite.overrideSendMsgs()

	chain, ok := suite.coordinator.Chains[ibctesting.GetChainID(1)]
	suite.Require().True(ok, "chain not found")
	suite.chain = chain

	app, ok := chain.App.(*gaiaApp.GaiaApp)
	suite.Require().True(ok, "expected App to be GaiaApp")
	suite.app = app
}

func (suite *FeeMarketTestSuite) TestBaseFeeAdjustment() {
	// BaseFee is initially set to DefaultMinBaseFee
	ctx := suite.chain.GetContext()
	baseFee, err := suite.app.FeeMarketKeeper.GetBaseFee(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(feemarkettypes.DefaultMinBaseFee, baseFee)

	// BaseFee can not be lower than DefaultMinBaseFee, even after N empty blocks
	suite.coordinator.CommitNBlocks(suite.chain, 10)

	ctx = suite.chain.GetContext()
	baseFee, err = suite.app.FeeMarketKeeper.GetBaseFee(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(feemarkettypes.DefaultMinBaseFee, baseFee)

	// Send a large transaction to consume a lot of gas
	sender := suite.chain.SenderAccounts[0].SenderAccount.GetAddress()
	receiver := suite.chain.SenderAccounts[1].SenderAccount.GetAddress()
	amount := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10)))

	msgs := make([]sdk.Msg, LargeMsgNumber)
	for i := 0; i < LargeMsgNumber; i++ {
		bankSendMsg := banktypes.NewMsgSend(sender, receiver, amount)
		msgs[i] = bankSendMsg
	}

	_, err = suite.chain.SendMsgs(msgs...)
	suite.Require().NoError(err)

	// Check that BaseFee has increased due to the large gas usage
	ctx = suite.chain.GetContext()
	baseFee, err = suite.app.FeeMarketKeeper.GetBaseFee(ctx)
	suite.Require().NoError(err)
	suite.Require().Greater(baseFee.Uint64(), feemarkettypes.DefaultMinBaseFee.Uint64())

	// BaseFee should drop to DefaultMinBaseFee after N empty blocks
	suite.coordinator.CommitNBlocks(suite.chain, 10)

	ctx = suite.chain.GetContext()
	baseFee, err = suite.app.FeeMarketKeeper.GetBaseFee(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(feemarkettypes.DefaultMinBaseFee, baseFee)
}

// SendMsgs() behavior must be changed since the default one uses zero fees
func (suite *FeeMarketTestSuite) overrideSendMsgs() {
	for _, chain := range suite.coordinator.Chains {
		chain.SendMsgsOverride = func(msgs ...sdk.Msg) (*sdk.Result, error) {
			return SendMsgsOverride(chain, msgs...)
		}
	}
}

func SendMsgsOverride(chain *ibctesting.TestChain, msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	_, r, err := SignAndDeliver(
		chain,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true, true,
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// NextBlock calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

func SignAndDeliver(
	chain *ibctesting.TestChain, txCfg client.TxConfig, app *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, expSimPass, expPass bool, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, LargeFeeAmount)},
		LargeGasLimit,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(chain.T, err)

	// Simulate a sending a transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	if expPass {
		require.NoError(chain.T, err)
		require.NotNil(chain.T, res)
	} else {
		require.Error(chain.T, err)
		require.Nil(chain.T, res)
	}

	return gInfo, res, err
}
