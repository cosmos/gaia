package e2e

import (
	"cosmossdk.io/x/evidence/exported"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
	"strings"
)

func (s *IntegrationTestSuite) testEvidence() {
	s.Run("test evidence queries", func() {
		var (
			valIdx   = 0
			chain    = s.commonHelper.Resources.ChainA
			chainAPI = fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[chain.Id][valIdx].GetHostPort("1317/tcp"))
		)
		res, err := query.QueryAllEvidence(chainAPI)
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
