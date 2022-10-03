package e2e

import "fmt"

func (s *IntegrationTestSuite) TestFeeGrant() {
	s.Run("test fee grant module", func() {
		var (
			valIdx = 0
			chain  = s.chainA
			api    = fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
		)

		alice, err := chain.genesisAccounts[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		bob, err := chain.genesisAccounts[1].keyInfo.GetAddress()
		s.Require().NoError(err)

		// add fee grant from alice to bob
		s.execFeeGrant(
			chain,
			valIdx,
			alice.String(),
			bob.String(),
			fees.String(),
		)

		bobBalance, err := getSpecificBalance(api, bob.String(), uatomDenom)
		s.Require().NoError(err)

		// withdrawal all balance + fee + fee granter flag should succeed
		s.execBankSend(
			chain,
			valIdx,
			bob.String(),
			Address(),
			tokenAmount.String(),
			fees.String(),
			false,
			withKeyValue(flagFeeGranter, alice.String()),
		)

		// check if the bob balance was subtracted without the fees
		expectedBobBalance := bobBalance.Sub(tokenAmount)
		bobBalance, err = getSpecificBalance(api, bob.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(expectedBobBalance, bobBalance)

		// tx should fail after spend limit reach
		s.execBankSend(
			chain,
			valIdx,
			bob.String(),
			Address(),
			tokenAmount.String(),
			fees.String(),
			true,
			withKeyValue(flagFeeGranter, alice.String()),
		)
	})
}
