package e2e

import (
	"fmt"
	"testing"
)

var (
	runBankTest                   = true
	runEncodeTest                 = true
	runEvidenceTest               = true
	runFeeGrantTest               = true
	runGovTest                    = true
	runIBCTest                    = true
	runSlashingTest               = true
	runStakingAndDistributionTest = true
	runVestingTest                = true
	runRestInterfacesTest         = true
	runLsmTest                    = true
	runRateLimitTest              = true
	runTxExtensionsTest           = true
	runCWTest                     = true
	runWasmLightClientTest        = true
)

// logTestStart logs when a test starts
func logTestStart(testName string) {
	fmt.Printf("▶️ Starting test: %s...\n", testName)
}

// logTestResult logs the result of a test
func logTestResult(t *testing.T, testName string) {
	if t.Failed() {
		fmt.Printf("❌ Test FAILED: %s\n", testName)
	} else {
		fmt.Printf("✅ Test PASSED: %s\n", testName)
	}
}

func (s *IntegrationTestSuite) TestRestInterfaces() {
	if !runRestInterfacesTest {
		s.T().Skip()
	}
	logTestStart("Rest Interfaces")
	s.testRestInterfaces()
	logTestResult("Rest Interfaces", s.T())
}

func (s *IntegrationTestSuite) TestBank() {
	if !runBankTest {
		s.T().Skip()
	}
	logTestStart("Bank Token Transfer")
	s.testBankTokenTransfer()
	logTestResult("Bank Token Transfer", s.T())
}

func (s *IntegrationTestSuite) TestEncode() {
	if !runEncodeTest {
		s.T().Skip()
	}
	logTestStart("Encoding")
	s.testEncode()
	s.testDecode()
	logTestResult("Encoding", s.T())
}

func (s *IntegrationTestSuite) TestEvidence() {
	if !runEvidenceTest {
		s.T().Skip()
	}
	logTestStart("Evidence Handling")
	s.testEvidence()
	logTestResult("Evidence Handling", s.T())
}

func (s *IntegrationTestSuite) TestFeeGrant() {
	if !runFeeGrantTest {
		s.T().Skip()
	}
	logTestStart("Fee Grant")
	s.testFeeGrant()
	logTestResult("Fee Grant", s.T())
}

func (s *IntegrationTestSuite) TestGov() {
	if !runGovTest {
		s.T().Skip()
	}
	logTestStart("Governance")
	s.GovCancelSoftwareUpgrade()
	s.GovCommunityPoolSpend()
	s.testSetBlocksPerEpoch()
	s.ExpeditedProposalRejected()
	s.GovSoftwareUpgradeExpedited()
	logTestResult("Governance", s.T())
}

func (s *IntegrationTestSuite) TestIBC() {
	if !runIBCTest {
		s.T().Skip()
	}
	logTestStart("IBC Transfer")
	s.testIBCTokenTransfer()
	s.testMultihopIBCTokenTransfer()
	s.testFailedMultihopIBCTokenTransfer()
	s.testICARegisterAccountAndSendTx()
	logTestResult("IBC Transfer", s.T())
}

func (s *IntegrationTestSuite) TestSlashing() {
	if !runSlashingTest {
		s.T().Skip()
	}
	logTestStart("Slashing")
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
	logTestResult("Slashing", s.T())
}

func (s *IntegrationTestSuite) TestStakingAndDistribution() {
	if !runStakingAndDistributionTest {
		s.T().Skip()
	}
	logTestStart("Staking & Distribution")
	s.testStaking()
	s.testDistribution()
	logTestResult("Staking & Distribution", s.T())
}

func (s *IntegrationTestSuite) TestVesting() {
	if !runVestingTest {
		s.T().Skip()
	}
	logTestStart("Vesting Accounts")
	chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
	logTestResult("Vesting Accounts", s.T())
}

func (s *IntegrationTestSuite) TestLSM() {
	if !runLsmTest {
		s.T().Skip()
	}
	logTestStart("Liquid Staking Module (LSM)")
	s.testLSM()
	logTestResult("Liquid Staking Module (LSM)", s.T())
}

func (s *IntegrationTestSuite) TestRateLimit() {
	if !runRateLimitTest {
		s.T().Skip()
	}
	logTestStart("Rate Limit")
	s.testAddRateLimits()
	s.testIBCTransfer(true)
	s.testUpdateRateLimit()
	s.testIBCTransfer(false)
	s.testResetRateLimit()
	s.testRemoveRateLimit()
	logTestResult("Rate Limit", s.T())
}

func (s *IntegrationTestSuite) TestTxExtensions() {
	if !runTxExtensionsTest {
		s.T().Skip()
	}
	logTestStart("Transaction Extensions")
	s.bankSendWithNonCriticalExtensionOptions()
	s.failedBankSendWithNonCriticalExtensionOptions()
	logTestResult("Transaction Extensions", s.T())
}

func (s *IntegrationTestSuite) TestCW() {
	if !runCWTest {
		s.T().Skip()
	}
	logTestStart("CosmWasm Tests")
	s.testCWCounter()
	logTestResult("CosmWasm Tests", s.T())
}

func (s *IntegrationTestSuite) TestWasmLightClient() {
	if !runWasmLightClientTest {
		s.T().Skip()
	}
	s.testStoreWasmLightClient()
	s.testCreateWasmLightClient()
}
