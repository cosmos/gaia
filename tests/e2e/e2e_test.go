package e2e

import (
	"fmt"
)

func (s *IntegrationTestSuite) TestBank() {
	s.testBankTokenTransfer()
}

func (s *IntegrationTestSuite) TestByPassMinFee() {
	// TODO: Add back after antehandler is fixed
	s.T().Skip("Skipping TestByPassMinFee")
	s.testByPassMinFeeWithdrawReward()
}

func (s *IntegrationTestSuite) TestDistribution() {
	s.testDistribution()
}

func (s *IntegrationTestSuite) TestEncode() {
	s.testEncode()
	s.testDecode()
}

func (s *IntegrationTestSuite) TestEvidence() {
	s.testEvidence()
}

// TODO: Fix and add back
func (s *IntegrationTestSuite) TestFeeGrant() {
	s.T().Skip("Skipping TestFeeGrant")
	s.testFeeGrant()
}

// TODO: Add back after antehandler is fixed
func (s *IntegrationTestSuite) TestGlobalFees() {
	s.T().Skip("Skipping TestGlobalFees")
	s.testGlobalFees()
	s.testQueryGlobalFeesInGenesis()
}

// TODO: Add back gov tests using the legacy gov system
func (s *IntegrationTestSuite) TestGov() {
	s.T().Skip("Skipping TestGov")
	s.SendTokensFromNewGovAccount()
	s.GovSoftwareUpgrade()
	s.GovCancelSoftwareUpgrade()
}

func (s *IntegrationTestSuite) TestIBC() {
	s.testIBCTokenTransfer()
	s.testBankTokenTransfer()
	s.T().Skip("Skipping multihop IBC Tests")
	s.testMultihopIBCTokenTransfer()
	s.testFailedMultihopIBCTokenTransfer()
}

func (s *IntegrationTestSuite) TestSlashing() {
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
}

// todo add fee test with wrong denom order
func (s *IntegrationTestSuite) TestStaking() {
	s.testStaking()
}

func (s *IntegrationTestSuite) TestVesting() {
	chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)

	// TODO: Add back vesting account here
	s.T().Skip("Skipping some TestVesting")
	s.testPermanentLockedAccount(chainAAPI)
	s.testPeriodicVestingAccount(chainAAPI)
}
