package e2e

import (
	"fmt"
	"net/http"
)

/*
RestRegression tests the continuity of critical endpoints that node operators, block explorers, and ecosystem participants depend on.
Test Node REST Endpoints:
1. http://host:1317/validatorsets/latest
2. http://host:1317/validatorsets/{height}
3. http://host:1317/blocks/latest
4. http://host:1317/blocks/{height}
5. http://host:1317/syncing
6. http://host:1317/node_info
7. http://host:1317/txs
Test Module REST Endpoints
1. Bank total
2. Auth params
3. Distribution for Community Pool
4. Evidence
5. Gov proposals
6. Mint params
7. Slashing params
8. Staking params
*/

const (
	valSetLatestPath                    = "/validatorsets/latest"
	valSetHeightPath                    = "/validatorsets/1"
	blocksLatestPath                    = "/blocks/latest"
	blocksHeightPath                    = "/blocks/1"
	syncingPath                         = "/syncing"
	nodeInfoPath                        = "/node_info"
	transactionsPath                    = "/txs"
	bankTotalModuleQueryPath            = "/bank/total"
	authParamsModuleQueryPath           = "/auth/params"
	distributionCommPoolModuleQueryPath = "/distribution/community_pool"
	evidenceModuleQueryPath             = "/evidence"
	govPropsModuleQueryPath             = "/gov/proposals"
	mintingParamsModuleQueryPath        = "/minting/parameters"
	slashingParamsModuleQueryPath       = "/slashing/parameters"
	stakingParamsModuleQueryPath        = "/staking/parameters"
	missingPath                         = "/missing_endpoint"
)

func (s *IntegrationTestSuite) testRestInterfaces() {
	s.Run("test rest interfaces", func() {
		var (
			valIdx        = 0
			c             = s.chainA
			endpointURL   = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
			testEndpoints = []struct {
				Path           string
				ExpectedStatus int
			}{
				// Client Endpoints
				{nodeInfoPath, 200},
				{syncingPath, 200},
				{valSetLatestPath, 200},
				{valSetHeightPath, 200},
				{blocksLatestPath, 200},
				{blocksHeightPath, 200},
				{transactionsPath, 200},
				// Module Endpoints
				{bankTotalModuleQueryPath, 200},
				{authParamsModuleQueryPath, 200},
				{distributionCommPoolModuleQueryPath, 200},
				{evidenceModuleQueryPath, 200},
				{govPropsModuleQueryPath, 200},
				{mintingParamsModuleQueryPath, 200},
				{slashingParamsModuleQueryPath, 200},
				{stakingParamsModuleQueryPath, 200},
				{missingPath, 501},
			}
		)

		for _, endpoint := range testEndpoints {
			resp, err := http.Get(fmt.Sprintf("%s%s", endpointURL, endpoint.Path))
			s.NoError(err, fmt.Sprintf("failed to get endpoint: %s%s", endpointURL, endpoint.Path))

			_, err = readJSON(resp)
			s.NoError(err, fmt.Sprintf("failed to read body of endpoint: %s%s", endpointURL, endpoint.Path))

			s.EqualValues(resp.StatusCode, endpoint.ExpectedStatus)
		}
	})
}
