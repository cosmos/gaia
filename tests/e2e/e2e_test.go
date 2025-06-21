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

func (s *IntegrationTestSuite) TestRestInterfaces() {
	testName := "REST Interfaces"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runRestInterfacesTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testRestInterfaces()
}

func (s *IntegrationTestSuite) TestBank() {
	testName := "Bank"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runBankTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testBankTokenTransfer()
}

func (s *IntegrationTestSuite) TestEncode() {
	testName := "Encode"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runEncodeTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testEncode()
	s.testDecode()
}

func (s *IntegrationTestSuite) TestEvidence() {
	testName := "Evidence"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runEvidenceTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testEvidence()
}

func (s *IntegrationTestSuite) TestFeeGrant() {
	testName := "FeeGrant"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runFeeGrantTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testFeeGrant()
}

func (s *IntegrationTestSuite) TestGov() {
	testName := "Gov"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runGovTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.GovCancelSoftwareUpgrade()
	s.GovCommunityPoolSpend()

	s.testSetBlocksPerEpoch()
	s.GovSoftwareUpgradeExpedited()
}

func (s *IntegrationTestSuite) TestIBC() {
	testName := "IBC"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runIBCTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testIBCTokenTransfer()
	s.testMultihopIBCTokenTransfer()
	s.testFailedMultihopIBCTokenTransfer()
	s.testICARegisterAccountAndSendTx()
}

func (s *IntegrationTestSuite) TestSlashing() {
	testName := "Slashing"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runSlashingTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	chainAPI := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
}

// todo add fee test with wrong denom order
func (s *IntegrationTestSuite) TestStakingAndDistribution() {
	testName := "Staking and Distribution"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runStakingAndDistributionTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testStaking()
	s.testDistribution()
}

func (s *IntegrationTestSuite) TestVesting() {
	testName := "Vesting"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runVestingTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	chainAAPI := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
	// s.testPeriodicVestingAccount(chainAAPI) TODO: add back when v0.45 adds the missing CLI command.
}

func (s *IntegrationTestSuite) TestLiquid() {
	testName := "Liquid"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runLiquidTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testLiquid()
	s.testLiquidGlobalLimit()
	s.testLiquidValidatorLimit()
}

func (s *IntegrationTestSuite) TestRateLimit() {
	testName := "Rate Limit"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runRateLimitTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testAddRateLimits(false)
	s.testIBCTransfer(true, false)
	s.testUpdateRateLimit(false)
	s.testIBCTransfer(false, false)
	s.testResetRateLimit(false)
	s.testRemoveRateLimit(false)
}

func (s *IntegrationTestSuite) TestTxExtensions() {
	testName := "Tx Extensions"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runTxExtensionsTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.bankSendWithNonCriticalExtensionOptions()
	s.failedBankSendWithNonCriticalExtensionOptions()
}

func (s *IntegrationTestSuite) TestCW() {
	testName := "CosmWasm"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runCWTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testCWCounter()
}

func (s *IntegrationTestSuite) TestIbcV2() {
	testName := "IBC v2"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runIbcV2Test {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

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
	testName := "Callbacks"
	s.T().Logf("=== STARTING TEST: %s ===", testName)

	if !runCallbacksTest {
		s.T().Logf("=== SKIPPED TEST: %s ===", testName)
		s.T().Skip()
	}

	defer func() {
		if r := recover(); r != nil {
			s.T().Logf("=== FAILED TEST: %s - %v ===", testName, r)
			panic(r)
		} else {
			s.T().Logf("=== PASSED TEST: %s ===", testName)
		}
	}()

	s.testCallbacksCWSkipGo()
}
