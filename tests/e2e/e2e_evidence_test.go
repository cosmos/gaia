package e2e

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cosmossdk.io/x/evidence/exported"
	evidencetypes "cosmossdk.io/x/evidence/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
	"github.com/cosmos/gaia/v26/tests/e2e/query"
)

func (s *IntegrationTestSuite) testEvidence() {
	s.Run("test evidence queries", func() {
		var (
			valIdx   = 0
			chain    = s.Resources.ChainA
			chainAPI = fmt.Sprintf("http://%s", s.Resources.ValResources[chain.ID][valIdx].GetHostPort("1317/tcp"))
		)
		res, err := query.AllEvidence(chainAPI)
		s.Require().NoError(err)
		s.Require().Equal(common.NumberOfEvidences, len(res.Evidence))
		for _, evidence := range res.Evidence {
			var exportedEvidence exported.Evidence
			err := common.Cdc.UnpackAny(evidence, &exportedEvidence)
			s.Require().NoError(err)
			eq, ok := exportedEvidence.(*evidencetypes.Equivocation)
			s.Require().True(ok)
			_, err = query.ExecQueryEvidence(chainAPI, strings.ToUpper(hex.EncodeToString(eq.Hash())))
			s.Require().NoError(err)
		}
	})
}
