package e2e

import (
	"github.com/cosmos/gaia/v27/tests/e2e/common"
	"github.com/cosmos/gaia/v27/tests/e2e/query"
)

const jailedValidatorKey = "jailed"

func (s *IntegrationTestSuite) testSlashing(chainEndpoint string) {
	s.Run("test unjail validator", func() {
		validators, err := query.Validators(chainEndpoint)
		s.Require().NoError(err)

		for _, val := range validators.Validators {
			if val.Jailed {
				s.ExecUnjail(
					s.Resources.ChainA,
					common.WithKeyValue(common.FlagFrom, jailedValidatorKey),
				)

				valQ, err := query.Validator(chainEndpoint, val.OperatorAddress)
				s.Require().NoError(err)
				s.Require().False(valQ.Jailed)
			}
		}
	})
}
