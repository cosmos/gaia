package e2e

import "fmt"

func (s *IntegrationTestSuite) TestFeeGrant() {
	s.Run("test fee grant module", func() {
		var (
			valIdx = 0
			chain  = s.chainA
			api    = fmt.Sprintf("http://%s", s.valResources[chain.id][0].GetHostPort("1317/tcp"))
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

		balance, err := getSpecificBalance(api, bob.String(), uatomDenom)
		s.Require().NoError(err)

		// withdrawal all balance + fee + fee granter flag should succeed
		s.execBankSend(
			chain,
			valIdx,
			bob.String(),
			Address(),
			balance.String(),
			fees.String(),
			false,
			withKeyValue(flagFeeGranter, alice.String()),
		)

		// withdrawal all balance again should fail after use the spend limit
		s.execBankSend(
			chain,
			valIdx,
			bob.String(),
			Address(),
			balance.String(),
			fees.String(),
			true,
			withKeyValue(flagFeeGranter, alice.String()),
		)

	})
}
