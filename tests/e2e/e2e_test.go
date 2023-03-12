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
	// s.testPeriodicVestingAccount(chainAAPI) TODO: add back when v0.45 adds the missing CLI command.
}
