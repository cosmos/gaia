package e2e

import (
	"context"
	"time"

	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
)

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
