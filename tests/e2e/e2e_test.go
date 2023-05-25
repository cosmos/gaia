package e2e

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"time"
)

var (
	runBankTest                   = true
	runBypassMinFeeTest           = true
	runEncodeTest                 = true
	runEvidenceTest               = true
	runFeeGrantTest               = true
	runGlobalFeesTest             = true
	runGovTest                    = true
	runIBCTest                    = true
	runSlashingTest               = true
	runStakingAndDistributionTest = true
	runVestingTest                = true
	runRestInterfacesTest         = true
)

func (s *IntegrationTestSuite) TestRestInterfaces() {
	if !runRestInterfacesTest {
		s.T().Skip()
	}
	s.testRestInterfaces()
}

func (s *IntegrationTestSuite) TestBank() {
	if !runBankTest {
		s.T().Skip()
	}
	s.testBankTokenTransfer()
}

func (s *IntegrationTestSuite) TestByPassMinFee() {
	if !runBypassMinFeeTest {
		s.T().Skip()
	}
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testBypassMinFeeWithdrawReward(chainAPI)
}

func (s *IntegrationTestSuite) TestEncode() {
	if !runEncodeTest {
		s.T().Skip()
	}
	s.testEncode()
	s.testDecode()
}

func (s *IntegrationTestSuite) TestEvidence() {
	if !runEvidenceTest {
		s.T().Skip()
	}
	s.testEvidence()
}

func (s *IntegrationTestSuite) TestFeeGrant() {
	if !runFeeGrantTest {
		s.T().Skip()
	}
	s.testFeeGrant()
}

func (s *IntegrationTestSuite) TestGlobalFees() {
	if !runGlobalFeesTest {
		s.T().Skip()
	}
	s.testGlobalFees()
	s.testQueryGlobalFeesInGenesis()
}

func (s *IntegrationTestSuite) TestGov() {
	if !runGovTest {
		s.T().Skip()
	}
	s.GovSoftwareUpgrade()
	s.GovCancelSoftwareUpgrade()
	s.GovCommunityPoolSpend()
	s.AddRemoveConsumerChain()
}

func (s *IntegrationTestSuite) TestIBC() {
	if !runIBCTest {
		s.T().Skip()
	}
	s.testIBCTokenTransfer()
	//s.testMultihopIBCTokenTransfer()
	//s.testFailedMultihopIBCTokenTransfer()

	// stop hermes0 to prevent hermes0 relaying transactions
	s.Require().NoError(s.dkrPool.Purge(s.hermesResource0))

	s.testIBCBypassMsg()
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
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{})}, proposalCounter, submitter, standardFees.String())

	// use hermes1 to test default ibc bypass-msg
	//
	// test 1: transaction only contains bypass-msgs, pass
	s.T().Logf("testing transaction contains only ibc bypass messages")
	ok := s.hermesTransfer(hermesCofigWithGasPrices, s.chainA.id, s.chainB.id, transferChannel, uatomDenom, 100, 1000, 1)
	s.Require().True(ok)

	scrRelayerBalanceBefore, dstRelayerBalanceBefore := s.queryRelayerWalletsBalances()

	pass := s.hermesClearPacket(hermesCofigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pass)
	pendingPacketsExist := s.hermesPendingPackets(hermesCofigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().False(pendingPacketsExist)

	// confirm relayer wallets do not pay fees
	scrRelayerBalanceAfter, dstRelayerBalanceAfter := s.queryRelayerWalletsBalances()
	s.Require().Equal(scrRelayerBalanceBefore.String(), scrRelayerBalanceAfter.String())
	s.Require().Equal(dstRelayerBalanceBefore.String(), dstRelayerBalanceAfter.String())

	// test 2: test transactions contains both bypass and non-bypass msgs (ibc msg mix with time out msg)
	s.T().Logf("testing transaction contains both bypass and non-bypass messages")
	ok = s.hermesTransfer(hermesCofigWithGasPrices, s.chainA.id, s.chainB.id, transferChannel, uatomDenom, 100, 1, 1)
	s.Require().True(ok)
	// make sure that the transaction is timeout
	time.Sleep(3)
	pendingPacketsExist = s.hermesPendingPackets(hermesCofigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pendingPacketsExist)

	pass = s.hermesClearPacket(hermesCofigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().False(pass)
	// clear packets with paying fee, to not influence the next transaction
	pass = s.hermesClearPacket(hermesCofigWithGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pass)

	// test 3: test transactions contains both bypass and non-bypass msgs (ibc msg mix with time out msg)
	s.T().Logf("testing bypass messages exceed MaxBypassGasUsage")
	ok = s.hermesTransfer(hermesCofigWithGasPrices, s.chainA.id, s.chainB.id, transferChannel, uatomDenom, 100, 1000, 12)
	s.Require().True(ok)
	pass = s.hermesClearPacket(hermesCofigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().False(pass)

	pendingPacketsExist = s.hermesPendingPackets(hermesCofigNoGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pendingPacketsExist)

	pass = s.hermesClearPacket(hermesCofigWithGasPrices, s.chainA.id, transferChannel)
	s.Require().True(pass)

	// set the default bypass-msg back
	//proposalCounter++
	//s.govProposeNewBypassMsgs([]string{
	//	sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
	//	sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
	//	sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
	//	sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{}),
	//	sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeoutOnClose{})}, proposalCounter, submitter, standardFees.String())
}

func (s *IntegrationTestSuite) TestSlashing() {
	if !runSlashingTest {
		s.T().Skip()
	}
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
}

// todo add fee test with wrong denom order
func (s *IntegrationTestSuite) TestStakingAndDistribution() {
	if !runStakingAndDistributionTest {
		s.T().Skip()
	}
	s.testStaking()
	s.testDistribution()
}

func (s *IntegrationTestSuite) TestVesting() {
	if !runVestingTest {
		s.T().Skip()
	}
	chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
	// s.testPeriodicVestingAccount(chainAAPI) TODO: add back when v0.45 adds the missing CLI command.
}
