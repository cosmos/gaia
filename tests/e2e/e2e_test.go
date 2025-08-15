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
	runLiquidTest                 = true
	runRateLimitTest              = true
	runTxExtensionsTest           = true
	runCWTest                     = true
	runCallbacksTest              = true
	runIbcV2Test                  = true
)

// logTestExecution handles consistent test logging across all test methods
func (s *IntegrationTestSuite) logTestExecution(testName string, runTest bool, testFunc func()) {
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		}
		s.T().Logf("=== PASSED TEST: %s ===", testName)
	}()

	testFunc()
}

func (s *IntegrationTestSuite) TestRestInterfaces() {
	s.logTestExecution("REST Interfaces", runRestInterfacesTest, func() {
		s.testRestInterfaces()
	})
}

func (s *IntegrationTestSuite) TestBank() {
	s.logTestExecution("Bank", runBankTest, func() {
		s.testBankTokenTransfer()
	})
}

func (s *IntegrationTestSuite) TestEncode() {
	s.logTestExecution("Encode", runEncodeTest, func() {
		s.testEncode()
		s.testDecode()
	})
}

func (s *IntegrationTestSuite) TestEvidence() {
	s.logTestExecution("Evidence", runEvidenceTest, func() {
		s.testEvidence()
	})
}

func (s *IntegrationTestSuite) TestFeeGrant() {
	s.logTestExecution("FeeGrant", runFeeGrantTest, func() {
		s.testFeeGrant()
	})
}

func (s *IntegrationTestSuite) TestGov() {
	s.logTestExecution("Gov", runGovTest, func() {
		s.GovCancelSoftwareUpgrade()
		s.GovCommunityPoolSpend()
		s.testSetBlocksPerEpoch()
		s.GovSoftwareUpgradeExpedited()
	})
}

func (s *IntegrationTestSuite) TestIBC() {
	s.logTestExecution("IBC", runIBCTest, func() {
		s.testIBCTokenTransfer()
		s.testMultihopIBCTokenTransfer()
		s.testFailedMultihopIBCTokenTransfer()
		s.testICARegisterAccountAndSendTx()
	})
}

func (s *IntegrationTestSuite) TestSlashing() {
	s.logTestExecution("Slashing", runSlashingTest, func() {
		chainAPI := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
		s.testSlashing(chainAPI)
	})
}

// todo add fee test with wrong denom order
func (s *IntegrationTestSuite) TestStakingAndDistribution() {
	s.logTestExecution("Staking and Distribution", runStakingAndDistributionTest, func() {
		s.testStaking()
		s.testDistribution()
	})
}

func (s *IntegrationTestSuite) TestVesting() {
	s.logTestExecution("Vesting", runVestingTest, func() {
		chainAAPI := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
		s.testDelayedVestingAccount(chainAAPI)
		s.testContinuousVestingAccount(chainAAPI)
		// s.testPeriodicVestingAccount(chainAAPI) TODO: add back when v0.45 adds the missing CLI command.
	})
}

func (s *IntegrationTestSuite) TestLiquid() {
	s.logTestExecution("Liquid", runLiquidTest, func() {
		s.testLiquid()
		s.testLiquidGlobalLimit()
		s.testLiquidValidatorLimit()
	})
}

func (s *IntegrationTestSuite) TestRateLimit() {
	s.logTestExecution("Rate Limit", runRateLimitTest, func() {
		s.testAddRateLimits(false)
		s.testIBCTransfer(true, false)
		s.testUpdateRateLimit(false)
		s.testIBCTransfer(false, false)
		s.testResetRateLimit(false)
		s.testRemoveRateLimit(false)
	})
}

func (s *IntegrationTestSuite) TestTxExtensions() {
	s.logTestExecution("Tx Extensions", runTxExtensionsTest, func() {
		s.bankSendWithNonCriticalExtensionOptions()
		s.failedBankSendWithNonCriticalExtensionOptions()
	})
}

func (s *IntegrationTestSuite) TestCW() {
	s.logTestExecution("CosmWasm", runCWTest, func() {
		s.testCWCounter()
	})
}

func (s *IntegrationTestSuite) TestIbcV2() {
	s.logTestExecution("IBC V2", runIbcV2Test, func() {
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
	})
}

func (s *IntegrationTestSuite) TestCallbacks() {
	s.logTestExecution("Callbacks", runCallbacksTest, func() {
		s.testCallbacksCWSkipGo()
	})
}
