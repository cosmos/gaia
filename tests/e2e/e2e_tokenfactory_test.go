package e2e

import "fmt"

func (s *IntegrationTestSuite) testTokenfactory() {
	s.Run("create, mint, burn, change admin", func() {
		var (
			err           error
			valIdx        = 0
			c             = s.chainA
			chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
		)

		denom := "testdenom"
		alice, _ := c.genesisAccounts[1].keyInfo.GetAddress()

		denoms, err := queryDenomsFromAdmin(chainEndpoint, alice.String())
		s.Require().NoError(err)
		s.Require().Equal(0, len(denoms.Denoms))

		// Create denom
		s.executeCreateDenom(c, valIdx, denom, gaiaHomePath, alice.String(), standardFees.String())

		// Check denoms from admin
		denoms, err = queryDenomsFromAdmin(chainEndpoint, alice.String())
		s.Require().NoError(err)
		s.Require().Equal(1, len(denoms.Denoms))
		s.Require().Equal(denom, denoms.Denoms[0])

		// Mint
		amount := "1000000"
		s.executeMint(c, valIdx, denom, amount, gaiaHomePath, alice.String(), standardFees.String())

		// Check balance
		balance, err := getSpecificBalance(chainEndpoint, alice.String(), denom)
		s.Require().NoError(err)
		s.Require().Equal(amount, balance.String())

		// Burn
		amount = "500000"
		s.executeBurn(c, valIdx, denom, amount, gaiaHomePath, alice.String(), standardFees.String())

		// Check balance
		balance, err = getSpecificBalance(chainEndpoint, alice.String(), denom)
		s.Require().NoError(err)
		s.Require().Equal(amount, balance.String())

		// Change admin
		bob, _ := c.genesisAccounts[2].keyInfo.GetAddress()
		s.executeChangeAdmin(c, valIdx, denom, bob.String(), gaiaHomePath, alice.String(), standardFees.String())

		// Check denoms from new admin
		denoms, err = queryDenomsFromAdmin(chainEndpoint, bob.String())
		s.Require().NoError(err)
		s.Require().Equal(1, len(denoms.Denoms))
		s.Require().Equal(denom, denoms.Denoms[0])
	})
}
