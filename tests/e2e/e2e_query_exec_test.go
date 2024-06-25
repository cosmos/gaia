package e2e

import (
	"context"
	"fmt"
	"time"

	ccvtypes "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

func (s *IntegrationTestSuite) execQueryConsumerChains(
	c *chain,
	valIdx int,
	homePath string,
	queryValidation func(res ccvtypes.QueryConsumerChainsResponse, consumerChainId string) bool,
	consumerChainID string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Querying consumer chains for chain: %s", c.id)
	gaiaCommand := []string{
		gaiadBinary,
		"query",
		"provider",
		"list-consumer-chains",
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--output=json",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.validateQueryConsumers(queryValidation, consumerChainID))
	s.T().Logf("Successfully queried consumer chains for chain %s", c.id)
}

func (s *IntegrationTestSuite) validateQueryConsumers(queryValidation func(ccvtypes.QueryConsumerChainsResponse, string) bool, consumerChainID string) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var queryConsumersRes ccvtypes.QueryConsumerChainsResponse
		if err := cdc.UnmarshalJSON(stdOut, &queryConsumersRes); err != nil {
			s.T().Logf("Error unmarshalling query consumer chains: %s", err.Error())
			return false
		}
		return queryValidation(queryConsumersRes, consumerChainID)
	}
}
