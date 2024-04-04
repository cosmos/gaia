package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gogoproto/proto"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gaiaApp "github.com/cosmos/gaia/v16/app"
)

// These integration tests were modified to work with the GaiaApp
// Sources:
// * Transfer tests: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/transfer_test.go#L13
// * ICA tests: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/ica_test.go#L94
var (

	// transfer + IBC fee test variables
	defaultRecvFee    = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(100)}}
	defaultAckFee     = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(200)}}
	defaultTimeoutFee = sdk.Coins{sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(300)}}

	// ICA + IBC fee test variables
	defaultOwnerAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"                                               // defaultOwnerAddress defines a reusable bech32 address for testing purposes
	defaultPortID, _    = icatypes.NewControllerPortID(defaultOwnerAddress)                                             // defaultPortID defines a reusable port identifier for testing purposes
	defaultICAVersion   = icatypes.NewDefaultMetadataString(ibctesting.FirstConnectionID, ibctesting.FirstConnectionID) // defaultICAVersion defines a reusable interchainaccounts version string for testing purposes
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

func TestFeeMarketTestSuite(t *testing.T) {
	ibcfeeSuite := &IBCFeeTestSuite{}
	suite.Run(t, ibcfeeSuite)
}

func (suite *IBCFeeTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = GaiaAppIniter
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

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
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
		fee.AckFee.Add(fee.TimeoutFee...), // ack fee paid, timeout fee refunded
		sdk.NewCoins(getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), ibctesting.TestCoin.Denom)).Sub(originalChainASenderAccountBalance[0]))
}

// TestFeeInterchainAccounts Integration test to ensure ics29 works with ics27
// Source: https://github.com/cosmos/ibc-go/blob/v7.3.2/modules/apps/29-fee/ica_test.go#L94
func (suite *IBCFeeTestSuite) TestFeeInterchainAccounts() {
	path := NewIncentivizedICAPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupConnections(path)

	err := SetupPath(path, defaultOwnerAddress)
	suite.Require().NoError(err)

	// assert the newly established channel is fee enabled on both ends
	suite.Require().True(getApp(suite.chainA).IBCFeeKeeper.IsFeeEnabled(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID))
	suite.Require().True(getApp(suite.chainB).IBCFeeKeeper.IsFeeEnabled(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID))

	// register counterparty address on destination chainB as chainA.SenderAccounts[1] for recv fee distribution
	getApp(suite.chainB).IBCFeeKeeper.SetCounterpartyPayeeAddress(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccounts[1].SenderAccount.GetAddress().String(), path.EndpointB.ChannelID)

	// escrow a packet fee for the next send sequence
	expectedFee := ibcfeetypes.NewFee(defaultRecvFee, defaultAckFee, defaultTimeoutFee)
	msgPayPacketFee := ibcfeetypes.NewMsgPayPacketFee(expectedFee, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, suite.chainA.SenderAccount.GetAddress().String(), nil)

	// fetch the account balance before fees are escrowed and assert the difference below
	preEscrowBalance := getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)

	res, err := suite.chainA.SendMsgs(msgPayPacketFee)
	suite.Require().NotNil(res)
	suite.Require().NoError(err)

	postEscrowBalance := getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	suite.Require().Equal(postEscrowBalance.AddAmount(expectedFee.Total().AmountOf(sdk.DefaultBondDenom)), preEscrowBalance)

	packetID := channeltypes.NewPacketID(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, 1)
	packetFees, found := getApp(suite.chainA).IBCFeeKeeper.GetFeesInEscrow(suite.chainA.GetContext(), packetID)
	suite.Require().True(found)
	suite.Require().Equal(expectedFee, packetFees.PacketFees[0].Fee)

	interchainAccountAddr, found := getApp(suite.chainB).ICAHostKeeper.GetInterchainAccountAddress(suite.chainB.GetContext(), ibctesting.FirstConnectionID, path.EndpointA.ChannelConfig.PortID)
	suite.Require().True(found)

	// fund the interchain account on chainB
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000)))
	msgBankSend := &banktypes.MsgSend{
		FromAddress: suite.chainB.SenderAccount.GetAddress().String(),
		ToAddress:   interchainAccountAddr,
		Amount:      coins,
	}

	res, err = suite.chainB.SendMsgs(msgBankSend)
	suite.Require().NotEmpty(res)
	suite.Require().NoError(err)

	// prepare a simple stakingtypes.MsgDelegate to be used as the interchain account msg executed on chainB
	validatorAddr := (sdk.ValAddress)(suite.chainB.Vals.Validators[0].Address)
	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: interchainAccountAddr,
		ValidatorAddress: validatorAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(5000)),
	}

	data, err := icatypes.SerializeCosmosTx(getApp(suite.chainA).AppCodec(), []proto.Message{msgDelegate})
	suite.Require().NoError(err)

	icaPacketData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// ensure chainB is allowed to execute stakingtypes.MsgDelegate
	params := icahosttypes.NewParams(true, []string{sdk.MsgTypeURL(msgDelegate)})
	getApp(suite.chainB).ICAHostKeeper.SetParams(suite.chainB.GetContext(), params)

	// build the interchain accounts packet
	packet := buildInterchainAccountsPacket(path, icaPacketData.GetBytes())

	// write packet commitment to state on chainA and commit state
	commitment := channeltypes.CommitPacket(getApp(suite.chainA).AppCodec(), packet)
	getApp(suite.chainA).IBCKeeper.ChannelKeeper.SetPacketCommitment(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, 1, commitment)
	suite.chainA.NextBlock()

	err = path.RelayPacket(packet)
	suite.Require().NoError(err)

	// ensure escrowed fees are cleaned up
	packetFees, found = getApp(suite.chainA).IBCFeeKeeper.GetFeesInEscrow(suite.chainA.GetContext(), packetID)
	suite.Require().False(found)
	suite.Require().Empty(packetFees)

	// assert the value of the account balance after fee distribution
	// NOTE: the balance after fee distribution should be equal to the pre-escrow balance minus the recv fee
	// as chainA.SenderAccount is used as the msg signer and refund address for msgPayPacketFee above as well as the relyer account for acknowledgements in path.RelayPacket()
	postDistBalance := getApp(suite.chainA).BankKeeper.GetBalance(suite.chainA.GetContext(), suite.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	suite.Require().Equal(preEscrowBalance.SubAmount(defaultRecvFee.AmountOf(sdk.DefaultBondDenom)), postDistBalance)
}

func buildInterchainAccountsPacket(path *ibctesting.Path, data []byte) channeltypes.Packet {
	packet := channeltypes.NewPacket(
		data,
		1,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)

	return packet
}

// NewIncentivizedICAPath creates and returns a new ibctesting path configured for a fee enabled interchain accounts channel
func NewIncentivizedICAPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)

	feeMetadata := ibcfeetypes.Metadata{
		FeeVersion: ibcfeetypes.Version,
		AppVersion: defaultICAVersion,
	}

	feeICAVersion := string(ibcfeetypes.ModuleCdc.MustMarshalJSON(&feeMetadata))

	path.SetChannelOrdered()
	path.EndpointA.ChannelConfig.Version = feeICAVersion
	path.EndpointB.ChannelConfig.Version = feeICAVersion
	path.EndpointA.ChannelConfig.PortID = defaultPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID

	return path
}

// SetupPath performs the InterchainAccounts channel creation handshake using an ibctesting path
func SetupPath(path *ibctesting.Path, owner string) error {
	if err := RegisterInterchainAccount(path.EndpointA, owner); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	return path.EndpointB.ChanOpenConfirm()
}

// RegisterInterchainAccount invokes the the InterchainAccounts entrypoint, routes a new MsgChannelOpenInit to the appropriate handler,
// commits state changes and updates the testing endpoint accordingly
func RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) error {
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	if err := getApp(endpoint.Chain).ICAControllerKeeper.RegisterInterchainAccount(endpoint.Chain.GetContext(), endpoint.ConnectionID, owner, endpoint.ChannelConfig.Version); err != nil {
		return err
	}

	// commit state changes for proof verification
	endpoint.Chain.NextBlock()

	// update port/channel ids
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID

	return nil
}

func getApp(chain *ibctesting.TestChain) *gaiaApp.GaiaApp {
	app, ok := chain.App.(*gaiaApp.GaiaApp)
	if !ok {
		panic("expected App to be GaiaApp")
	}
	return app
}
