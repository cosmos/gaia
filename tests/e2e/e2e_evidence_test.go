package e2e

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/x/evidence/exported"
	evidencetypes "cosmossdk.io/x/evidence/types"
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

func (s *IntegrationTestSuite) execQueryEvidence(c *chain, valIdx int, hash string) (res evidencetypes.Equivocation) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("querying evidence %s on chain %s", hash, c.id)

	gaiaCommand := []string{
		gaiadBinary,
		queryCommand,
		evidencetypes.ModuleName,
		hash,
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		// TODO parse evidence after fix the SDK
		// https://github.com/cosmos/cosmos-sdk/issues/13444
		// s.Require().NoError(yaml.Unmarshal(stdOut, &res))
		return true
	})
	return res
}
