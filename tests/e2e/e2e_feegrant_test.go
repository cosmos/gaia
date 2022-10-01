package e2e

func (s *IntegrationTestSuite) TestFeeGrant() {
	s.Run("test fee grant module", func() {
		var (
			valIdx = 0
			chain  = s.chainA
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
