package e2e

const jailedValidatorKey = "jailed"

func (s *IntegrationTestSuite) testSlashing(chainEndpoint string) {
	s.Run("test unjail validator", func() {
		validators, err := queryValidators(chainEndpoint)
		s.Require().NoError(err)

		for _, val := range validators.Validators {
			if val.Jailed {
				s.execUnjail(
					s.chainA,
					withKeyValue(flagFrom, jailedValidatorKey),
				)

				valQ, err := queryValidator(chainEndpoint, val.OperatorAddress)
				s.Require().NoError(err)
				s.Require().False(valQ.Jailed)
			}
		}
	})
}
