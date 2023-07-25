package e2e

import (
	"time"

	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *IntegrationTestSuite) testBypassMinFeeWithdrawReward(endpoint string) {
	// submit gov prop to change bypass-msg param to MsgWithdrawDelegatorReward
	submitterAddr := s.chainA.validators[0].keyInfo.GetAddress()
	submitter := submitterAddr.String()
	proposalCounter++
	s.govProposeNewBypassMsgs([]string{sdk.MsgTypeURL(&distributiontypes.MsgWithdrawDelegatorReward{})}, proposalCounter, submitter, standardFees.String())

	paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
	payee := s.chainA.validators[0].keyInfo.GetAddress()

	testCases := []struct {
		name                    string
		fee                     string
		changeMaxBypassGasUsage bool
		expErr                  bool
	}{
		{
			"bypass-msg with fee in the denom of global fee, pass",
			paidFeeAmt + uatomDenom,
			false,
			false,
		},
		{
			"bypass-msg with zero coin in the denom of global fee, pass",
			"0" + uatomDenom,
			false,
			false,
		},
		{
			"bypass-msg with zero coin not in the denom of global fee, pass",
			"0" + photonDenom,
			false,
			false,
		},
		{
			"bypass-msg with non-zero coin not in the denom of global fee, fail",
			paidFeeAmt + photonDenom,
			false,
			true,
		},
		{
			"bypass-msg with zero coin in the denom of global fee and maxTotalBypassMinFeeMsgGasUsage set to 1, fail",
			"0" + uatomDenom,
			true,
			true,
		},
		{
			"bypass-msg with non zero coin in the denom of global fee and maxTotalBypassMinFeeMsgGasUsage set to 1, pass",
			paidFeeAmt + uatomDenom,
			false,
			false,
		},
	}

	for _, tc := range testCases {

		if tc.changeMaxBypassGasUsage {
			proposalCounter++
			// change MaxTotalBypassMinFeeMsgGasUsage through governance proposal from 1_000_0000 to 1
			s.govProposeNewMaxTotalBypassMinFeeMsgGasUsage(1, proposalCounter, submitter)
		}

		// get delegator rewards
		rewards, err := queryDelegatorTotalRewards(endpoint, payee.String())
		s.Require().NoError(err)

		// get delegator stake balance
		oldBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
		s.Require().NoError(err)

		// withdraw rewards
		s.Run(tc.name, func() {
			s.execWithdrawAllRewards(s.chainA, 0, payee.String(), tc.fee, tc.expErr)
		})

		if !tc.expErr {
			// get updated balance
			incrBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
			s.Require().NoError(err)

			// compute sum of old balance and stake token rewards
			oldBalancePlusReward := rewards.GetTotal().Add(sdk.NewDecCoinFromCoin(oldBalance))
			s.Require().Equal(oldBalancePlusReward[0].Denom, stakeDenom)

			// check updated balance got increased by at least oldBalancePlusReward
			s.Require().True(sdk.NewDecCoinFromCoin(incrBalance).IsGTE(oldBalancePlusReward[0]))
		}
	}
}

func (s *IntegrationTestSuite) testIBCBypassMsg() {
	// submit gov proposal to change bypass-msg param to
	// ["/ibc.core.channel.v1.MsgRecvPacket",
	//  "/ibc.core.channel.v1.MsgAcknowledgement",
	//  "/ibc.core.client.v1.MsgUpdateClient"]
	submitterAddr := s.chainA.validators[0].keyInfo.GetAddress()
	submitter := submitterAddr.String()
	proposalCounter++
	s.govProposeNewBypassMsgs([]string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
	}, proposalCounter, submitter, standardFees.String())

	// use hermes1 to test default ibc bypass-msg
	//
	// test 1: transaction only contains bypass-msgs, pass
	s.testTxContainsOnlyIBCBypassMsg()
	// test 2: test transactions contains both bypass and non-bypass msgs (sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{})
	s.testTxContainsMixBypassNonBypassMsg()
	// test 3: test bypass-msgs exceed the MaxBypassGasUsage
	s.testBypassMsgsExceedMaxBypassGasLimit()

	// set the default bypass-msg back
	proposalCounter++
	s.govProposeNewBypassMsgs([]string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeoutOnClose{}),
	}, proposalCounter, submitter, standardFees.String())
}

func (s *IntegrationTestSuite) testTxContainsOnlyIBCBypassMsg() {
	s.T().Logf("testing transaction contains only ibc bypass messages")
	ok := s.hermesTransfer(hermesConfigWithGasPrices, s.chainA.id, s.chainB.id, transferChannel, uatomDenom, 100, 1000, 1)
	s.Require().True(ok)

	scrRelayerBalanceBefore, dstRelayerBalanceBefore := s.queryRelayerWalletsBalances()

	pass := s.hermesClearPacket(hermesConfigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pass)
	pendingPacketsExist := s.hermesPendingPackets(hermesConfigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().False(pendingPacketsExist)

	// confirm relayer wallets do not pay fees
	scrRelayerBalanceAfter, dstRelayerBalanceAfter := s.queryRelayerWalletsBalances()
	s.Require().Equal(scrRelayerBalanceBefore.String(), scrRelayerBalanceAfter.String())
	s.Require().Equal(dstRelayerBalanceBefore.String(), dstRelayerBalanceAfter.String())
}

func (s *IntegrationTestSuite) testTxContainsMixBypassNonBypassMsg() {
	s.T().Logf("testing transaction contains both bypass and non-bypass messages")
	// hermesTransfer with --timeout-height-offset=1
	ok := s.hermesTransfer(hermesConfigWithGasPrices, s.chainA.id, s.chainB.id, transferChannel, uatomDenom, 100, 1, 1)
	s.Require().True(ok)
	// make sure that the transaction is timeout
	time.Sleep(3 * time.Second)
	pendingPacketsExist := s.hermesPendingPackets(hermesConfigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pendingPacketsExist)

	pass := s.hermesClearPacket(hermesConfigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().False(pass)
	// clear packets with paying fee, to not influence the next transaction
	pass = s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pass)
}

func (s *IntegrationTestSuite) testBypassMsgsExceedMaxBypassGasLimit() {
	s.T().Logf("testing bypass messages exceed MaxBypassGasUsage")
	ok := s.hermesTransfer(hermesConfigWithGasPrices, s.chainA.id, s.chainB.id, transferChannel, uatomDenom, 100, 1000, 12)
	s.Require().True(ok)
	pass := s.hermesClearPacket(hermesConfigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().False(pass)

	pendingPacketsExist := s.hermesPendingPackets(hermesConfigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pendingPacketsExist)

	pass = s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pass)
}
