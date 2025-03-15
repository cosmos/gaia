package e2e

import "fmt"

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
	runCallbacksTest              = true
	runIbcV2Test                  = true
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

func (s *IntegrationTestSuite) TestGov() {
	if !runGovTest {
		s.T().Skip()
	}

	s.GovCancelSoftwareUpgrade()
	s.GovCommunityPoolSpend()

	s.testSetBlocksPerEpoch()
	s.ExpeditedProposalRejected()
	s.GovSoftwareUpgradeExpedited()
}

func (s *IntegrationTestSuite) TestIBC() {
	if !runIBCTest {
		s.T().Skip()
	}

	s.testIBCTokenTransfer()
	s.testMultihopIBCTokenTransfer()
	s.testFailedMultihopIBCTokenTransfer()
	s.testICARegisterAccountAndSendTx()
}

func (s *IntegrationTestSuite) TestSlashing() {
	if !runSlashingTest {
		s.T().Skip()
	}
	chainAPI := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
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
	chainAAPI := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
	// s.testPeriodicVestingAccount(chainAAPI) TODO: add back when v0.45 adds the missing CLI command.
}

func (s *IntegrationTestSuite) TestLSM() {
	if !runLsmTest {
		s.T().Skip()
	}
	s.testLSM()
}

func (s *IntegrationTestSuite) TestRateLimit() {
	if !runRateLimitTest {
		s.T().Skip()
	}
	s.testAddRateLimits(false)
	s.testIBCTransfer(true, false)
	s.testUpdateRateLimit(false)
	s.testIBCTransfer(false, false)
	s.testResetRateLimit(false)
	s.testRemoveRateLimit(false)
}

func (s *IntegrationTestSuite) TestTxExtensions() {
	if !runTxExtensionsTest {
		s.T().Skip()
	}
	s.bankSendWithNonCriticalExtensionOptions()
	s.failedBankSendWithNonCriticalExtensionOptions()
}

func (s *IntegrationTestSuite) TestCW() {
	if !runCWTest {
		s.T().Skip()
	}
	s.testCWCounter()
}

func (s *IntegrationTestSuite) TestIbcV2() {
	if !runIbcV2Test {
		s.T().Skip()
	}

	// ibc v2 wasm light client tests
	s.testStoreWasmLightClient()
	s.testCreateWasmLightClient()
	s.TestV2RecvPacket()
	s.TestV2Callback()

	// ibc v2 rate limiting tests
	s.testAddRateLimits(true)
	s.testIBCTransfer(true, true)
	s.testUpdateRateLimit(true)
	s.testIBCTransfer(false, true)
	s.testResetRateLimit(true)
	s.testRemoveRateLimit(true)
}

func (s *IntegrationTestSuite) TestCallbacks() {
	if !runCallbacksTest {
		s.T().Skip()
	}

	s.testCallbacksCWSkipGo()
}
