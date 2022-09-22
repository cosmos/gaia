package e2e

import "time"

func (s *IntegrationTestSuite) testSlashing(chainEndpoint string) {
	time.Sleep(30 * time.Second)
	validators, err := queryValidators(chainEndpoint)
	s.Require().NoError(err)

	for _, val := range validators {
		if val.Jailed {

			valQ, err := queryValidator(chainEndpoint, val.OperatorAddress)
			s.Require().NoError(err)
			s.T().Logf("validator: %s", valQ)
		}
	}
}
