package e2e

import (
	"cosmossdk.io/x/evidence/exported"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"encoding/hex"
	"fmt"
	"strings"
)

func (s *IntegrationTestSuite) testEvidence() {
	s.Run("test evidence queries", func() {
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
			s.execQueryEvidence(chain, valIdx, strings.ToUpper(hex.EncodeToString(eq.Hash())))
		}
	})
}
