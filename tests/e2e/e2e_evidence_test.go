package e2e

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func (s *IntegrationTestSuite) TestEvidence() {
	s.Run("teste query evidences", func() {
		var (
			valIdx   = 0
			chain    = s.chainA
			chainAPI = fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
		)
		res, err := queryAllEvidence(chainAPI)
		s.Require().NoError(err)
		s.Require().Equal(numberOfEvidences, len(res.Evidence))
		for _, evidence := range res.Evidence {
			var exportedEvidence exported.Evidence
			err := cdc.UnpackAny(evidence, &exportedEvidence)
			s.Require().NoError(err)
			eq, ok := exportedEvidence.(*evidencetypes.Equivocation)
			s.Require().True(ok)
			s.execQueryEvidence(chain, valIdx, eq.Hash().String())
		}
	})
}
