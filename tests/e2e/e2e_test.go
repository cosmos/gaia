package e2e

import (
	"fmt"
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
)

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
	s.testByPassMinFeeWithdrawReward()
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

// TODO: Add back after antehandler is fixed
func (s *IntegrationTestSuite) TestGlobalFees() {
	if !runGlobalFeesTest {
		s.T().Skip()
	}
	s.testGlobalFees()
	s.testQueryGlobalFeesInGenesis()
}

// TODO: Add back gov tests using the legacy gov system
func (s *IntegrationTestSuite) TestGov() {
	if !runGovTest {
		s.T().Skip()
	}
	s.GovSoftwareUpgrade()
	s.GovCancelSoftwareUpgrade()
	s.GovCommunityPoolSpend()
}

func (s *IntegrationTestSuite) TestIBC() {
	if !runIBCTest {
		s.T().Skip()
	}
	s.testIBCTokenTransfer()
	s.testMultihopIBCTokenTransfer()
	s.testFailedMultihopIBCTokenTransfer()
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

	// TODO: Add back vesting account here
	// s.testPermanentLockedAccount(chainAAPI)
	// s.testPeriodicVestingAccount(chainAAPI)
}
