package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"

	"github.com/cosmos/gaia/v8/x/icamauth/keeper"
	"github.com/cosmos/gaia/v8/x/icamauth/types"
)

func (suite *KeeperTestSuite) TestRegisterInterchainAccount() {
	var (
		owner string
		path  *ibctesting.Path
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {}, true,
		},
		{
			"failure - port is already bound",
			func() {
				GetICAApp(suite.chainA).IBCKeeper.PortKeeper.BindPort(suite.chainA.GetContext(), TestPortID)
			},
			false,
		},
		{
			"faliure - owner is empty",
			func() {
				owner = ""
			},
			false,
		},
		{
			"failure - channel is already active",
			func() {
				portID, err := icatypes.NewControllerPortID(owner)
				suite.Require().NoError(err)

				channel := channeltypes.NewChannel(
					channeltypes.OPEN,
					channeltypes.ORDERED,
					channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID),
					[]string{path.EndpointA.ConnectionID},
					path.EndpointA.ChannelConfig.Version,
				)

				GetICAApp(suite.chainA).IBCKeeper.ChannelKeeper.SetChannel(suite.chainA.GetContext(), portID, ibctesting.FirstChannelID, channel)
				GetICAApp(suite.chainA).ICAControllerKeeper.SetActiveChannelID(suite.chainA.GetContext(), ibctesting.FirstConnectionID, portID, ibctesting.FirstChannelID)
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			owner = TestOwnerAddress // must be explicitly changed

			path = NewICAPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			tc.malleate() // malleate mutates test data

			msgSrv := keeper.NewMsgServerImpl(GetICAApp(suite.chainA).ICAMauthKeeper)
			msg := types.NewMsgRegisterAccount(owner, path.EndpointA.ConnectionID, path.EndpointA.ChannelConfig.Version)

			res, err := msgSrv.RegisterAccount(sdk.WrapSDKContext(suite.chainA.GetContext()), msg)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSubmitTx() {
	var (
		path                      *ibctesting.Path
		registerInterchainAccount bool
		owner                     string
		connectionId              string
		icaMsg                    sdk.Msg
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
			}, true,
		},
		{
			"failure - owner address is empty", func() {
				registerInterchainAccount = true
				owner = ""
				connectionId = path.EndpointA.ConnectionID
			}, false,
		},
		{
			"failure - active channel does not exist for connection ID", func() {
				registerInterchainAccount = true
				owner = TestOwnerAddress
				connectionId = "connection-100"
			}, false,
		},
		{
			"failure - active channel does not exist for port ID", func() {
				registerInterchainAccount = true
				owner = "cosmos153lf4zntqt33a4v0sm5cytrxyqn78q7kz8j8x5"
				connectionId = path.EndpointA.ConnectionID
			}, false,
		},
		{
			"failure - module does not own channel capability", func() {
				registerInterchainAccount = false
				owner = TestOwnerAddress
				connectionId = path.EndpointA.ConnectionID
				icaMsg = &banktypes.MsgSend{
					FromAddress: "source-address",
					ToAddress:   "destination-address",
					Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
				}
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			icaAppA := GetICAApp(suite.chainA)
			icaAppB := GetICAApp(suite.chainB)

			path = NewICAPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)

			tc.malleate() // malleate mutates test data

			if registerInterchainAccount {
				err := SetupICAPath(path, TestOwnerAddress)
				suite.Require().NoError(err)

				portID, err := icatypes.NewControllerPortID(TestOwnerAddress)
				suite.Require().NoError(err)

				// Get the address of the interchain account stored in state during handshake step
				interchainAccountAddr, found := GetICAApp(suite.chainA).ICAControllerKeeper.GetInterchainAccountAddress(suite.chainA.GetContext(), path.EndpointA.ConnectionID, portID)
				suite.Require().True(found)

				icaAddr, err := sdk.AccAddressFromBech32(interchainAccountAddr)
				suite.Require().NoError(err)

				// Check if account is created
				interchainAccount := icaAppB.AccountKeeper.GetAccount(suite.chainB.GetContext(), icaAddr)
				suite.Require().Equal(interchainAccount.GetAddress().String(), interchainAccountAddr)

				// Create bank transfer message to execute on the host
				icaMsg = &banktypes.MsgSend{
					FromAddress: interchainAccountAddr,
					ToAddress:   suite.chainB.SenderAccount.GetAddress().String(),
					Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
				}
			}

			msgSrv := keeper.NewMsgServerImpl(icaAppA.ICAMauthKeeper)
			msg, err := types.NewMsgSubmitTx(icaMsg, connectionId, owner)
			suite.Require().NoError(err)

			res, err := msgSrv.SubmitTx(sdk.WrapSDKContext(suite.chainA.GetContext()), msg)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}
