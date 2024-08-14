package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v20/ante"
	gaiaApp "github.com/cosmos/gaia/v20/app"
)

// These integration tests were modified to work with the GaiaApp
// Sources:
// * Transfer tests: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/transfer_test.go#L13
// * ICA tests: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/ica_test.go#L94
var (

	// transfer + IBC fee test variables
	defaultRecvFee    = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(100)}}
	defaultAckFee     = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(200)}}
	defaultTimeoutFee = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(300)}}
)

type IBCFeeTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	chainC *ibctesting.TestChain

	path     *ibctesting.Path
	pathAToC *ibctesting.Path
}

func TestIBCFeeTestSuite(t *testing.T) {
	ibcfeeSuite := &IBCFeeTestSuite{}
	suite.Run(t, ibcfeeSuite)
}

func (suite *IBCFeeTestSuite) SetupTest() {
	ante.UseFeeMarketDecorator = false
	ibctesting.DefaultTestingAppInit = GaiaAppIniterTempDir
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	suite.chainC = suite.coordinator.GetChain(ibctesting.GetChainID(3))

	chain, ok := suite.coordinator.Chains[ibctesting.GetChainID(1)]
	suite.Require().True(ok, "chain not found")

	_, ok = chain.App.(*gaiaApp.GaiaApp)
	suite.Require().True(ok, "expected App to be GaiaApp")

	path := ibctesting.NewPath(suite.chainA, suite.chainB)
	mockFeeVersion := string(ibcfeetypes.ModuleCdc.MustMarshalJSON(
		&ibcfeetypes.Metadata{
			FeeVersion: ibcfeetypes.Version,
			AppVersion: "test-version",
		},
	))
	path.EndpointA.ChannelConfig.Version = mockFeeVersion
	path.EndpointB.ChannelConfig.Version = mockFeeVersion
	path.EndpointA.ChannelConfig.PortID = ibctesting.MockFeePort
	path.EndpointB.ChannelConfig.PortID = ibctesting.MockFeePort
	suite.path = path

	path = ibctesting.NewPath(suite.chainA, suite.chainC)
	path.EndpointA.ChannelConfig.Version = mockFeeVersion
	path.EndpointB.ChannelConfig.Version = mockFeeVersion
	path.EndpointA.ChannelConfig.PortID = ibctesting.MockFeePort
	path.EndpointB.ChannelConfig.PortID = ibctesting.MockFeePort
	suite.pathAToC = path
}

// Integration test to ensure ics29 works with ics20
// Source: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/transfer_test.go#L13
func (suite *IBCFeeTestSuite) TestFeeTransfer() {
	path := ibctesting.NewPath(suite.chainA, suite.chainB)
	feeTransferVersion := string(ibcfeetypes.ModuleCdc.MustMarshalJSON(&ibcfeetypes.Metadata{FeeVersion: ibcfeetypes.Version, AppVersion: transfertypes.Version}))
	path.EndpointA.ChannelConfig.Version = feeTransferVersion
	path.EndpointB.ChannelConfig.Version = feeTransferVersion
	path.EndpointA.ChannelConfig.PortID = transfertypes.PortID
	path.EndpointB.ChannelConfig.PortID = transfertypes.PortID

	suite.coordinator.Setup(path)

	// set up coin & ics20 packet
	coin := ibctesting.TestCoin
	fee := ibcfeetypes.Fee{
		RecvFee:    defaultRecvFee,
		AckFee:     defaultAckFee,
		TimeoutFee: defaultTimeoutFee,
	}

	msgs := []sdk.Msg{
		ibcfeetypes.NewMsgPayPacketFee(fee, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, suite.chainA.SenderAccount.GetAddress().String(), nil),
		transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coin, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), clienttypes.NewHeight(1, 100), 0, ""),
	}
	res, err := suite.chainA.SendMsgs(msgs...)
	suite.Require().NoError(err) // message committed

	// after incentivizing the packets
	originalChainASenderAccountBalance := sdk.NewCoins(getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), ibctesting.TestCoin.Denom))

	packet, err := ibctesting.ParsePacketFromEvents(res.Events)
	suite.Require().NoError(err)

	// register counterparty address on chainB
	// relayerAddress is address of sender account on chainB, but we will use it on chainA
	// to differentiate from the chainA.SenderAccount for checking successful relay payouts
	relayerAddress := suite.chainB.SenderAccount.GetAddress()

	msgRegister := ibcfeetypes.NewMsgRegisterCounterpartyPayee(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, suite.chainB.SenderAccount.GetAddress().String(), relayerAddress.String())
	_, err = suite.chainB.SendMsgs(msgRegister)
	suite.Require().NoError(err) // message committed

	// relay packet
	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	// ensure relayers got paid
	// relayer for forward relay: chainB.SenderAccount
	// relayer for reverse relay: chainA.SenderAccount

	// check forward relay balance
	suite.Require().Equal(
		fee.RecvFee,
		sdk.NewCoins(getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainB.SenderAccount.GetAddress(), ibctesting.TestCoin.Denom)),
	)

	suite.Require().Equal(
		fee.AckFee, // ack fee paid, no refund needed since timeout_fee = recv_fee + ack_fee
		sdk.NewCoins(getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), ibctesting.TestCoin.Denom)).Sub(originalChainASenderAccountBalance[0]))
}

func getApp(chain *ibctesting.TestChain) *gaiaApp.GaiaApp {
	app, ok := chain.App.(*gaiaApp.GaiaApp)
	if !ok {
		panic("expected App to be GaiaApp")
	}
	return app
}
