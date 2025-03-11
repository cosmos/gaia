package e2e

import (
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
)

const jailedValidatorKey = "jailed"

func (s *IntegrationTestSuite) testSlashing(chainEndpoint string) {
	s.Run("test unjail validator", func() {
		validators, err := query.QueryValidators(chainEndpoint)
		s.Require().NoError(err)

		for _, val := range validators.Validators {
			if val.Jailed {
				s.tx.ExecUnjail(
					s.commonHelper.Resources.ChainA,
					common.WithKeyValue(common.FlagFrom, jailedValidatorKey),
				)

				valQ, err := query.QueryValidator(chainEndpoint, val.OperatorAddress)
				s.Require().NoError(err)
				s.Require().False(valQ.Jailed)
			}
		}
	})
}
