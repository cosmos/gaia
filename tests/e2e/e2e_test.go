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
	logTestResult(s.T(),"Rest Interfaces")
}

func (s *IntegrationTestSuite) TestBank() {
	if !runBankTest {
		s.T().Skip()
	}
	logTestStart("Bank Token Transfer")
	s.testBankTokenTransfer()
	logTestResult(s.T(),"Bank Token Transfer")
}

func (s *IntegrationTestSuite) TestEncode() {
	if !runEncodeTest {
		s.T().Skip()
	}
	logTestStart("Encoding")
	s.testEncode()
	s.testDecode()
	logTestResult(s.T(),"Encoding")
}

func (s *IntegrationTestSuite) TestEvidence() {
	if !runEvidenceTest {
		s.T().Skip()
	}
	logTestStart("Evidence Handling")
	s.testEvidence()
	logTestResult(s.T(),"Evidence Handling")
}

func (s *IntegrationTestSuite) TestFeeGrant() {
	if !runFeeGrantTest {
		s.T().Skip()
	}
	logTestStart("Fee Grant")
	s.testFeeGrant()
	logTestResult(s.T(),"Fee Grant")
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
	logTestResult(s.T(),"Governance")
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
	logTestResult(s.T(),"IBC Transfer")
}

func (s *IntegrationTestSuite) TestSlashing() {
	if !runSlashingTest {
		s.T().Skip()
	}
	logTestStart("Slashing")
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
	logTestResult(s.T(),"Slashing")
}

func (s *IntegrationTestSuite) TestStakingAndDistribution() {
	if !runStakingAndDistributionTest {
		s.T().Skip()
	}
	logTestStart("Staking & Distribution")
	s.testStaking()
	s.testDistribution()
	logTestResult(s.T(),"Staking & Distribution")
}

func (s *IntegrationTestSuite) TestVesting() {
	if !runVestingTest {
		s.T().Skip()
	}
	logTestStart("Vesting Accounts")
	chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
	logTestResult(s.T(),"Vesting Accounts")
}

func (s *IntegrationTestSuite) TestLSM() {
	if !runLsmTest {
		s.T().Skip()
	}
	logTestStart("Liquid Staking Module (LSM)")
	s.testLSM()
	logTestResult(s.T(),"Liquid Staking Module (LSM)")
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
	logTestResult(s.T(),"Rate Limit")
}

func (s *IntegrationTestSuite) TestTxExtensions() {
	if !runTxExtensionsTest {
		s.T().Skip()
	}
	logTestStart("Transaction Extensions")
	s.bankSendWithNonCriticalExtensionOptions()
	s.failedBankSendWithNonCriticalExtensionOptions()
	logTestResult(s.T(),"Transaction Extensions")
}

func (s *IntegrationTestSuite) TestCW() {
	if !runCWTest {
		s.T().Skip()
	}
	logTestStart("CosmWasm Tests")
	s.testCWCounter()
	logTestResult(s.T(),"CosmWasm Tests")
}


func (s *IntegrationTestSuite) TestWasmLightClient() {
	if !runWasmLightClientTest {
		s.T().Skip()
	}
	logTestStart("Wasm Light Client") 
	s.testStoreWasmLightClient()
	s.testCreateWasmLightClient()
	logTestResult(s.T(), "Wasm Light Client")
}
