package integrator_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/semver"
)

type EndpointsSuite struct {
	*chainsuite.Suite
}

func (s *EndpointsSuite) TestAPIEndpoints() {
	wallets := s.Chain.ValidatorWallets
	const proposalID = "1"

	tests := []struct {
		name string
		path string
		key  string
	}{
		{name: "auth", path: "/cosmos/auth/v1beta1/accounts", key: "accounts"},
		{name: "bank", path: "/cosmos/bank/v1beta1/balances/" + wallets[0].Address, key: "balances"},
		{name: "bank_denoms_metadata", path: "/cosmos/bank/v1beta1/denoms_metadata", key: "metadatas"},
		{name: "supply", path: "/cosmos/bank/v1beta1/supply", key: "supply"},
		{name: "dist_slashes", path: "/cosmos/distribution/v1beta1/validators/" + wallets[0].ValoperAddress + "/slashes", key: "slashes"},
		{name: "evidence", path: "/cosmos/evidence/v1beta1/evidence", key: "evidence"},
		{name: "gov_proposals", path: "/cosmos/gov/v1beta1/proposals", key: "proposals"},
		{name: "gov_deposits", path: "/cosmos/gov/v1beta1/proposals/" + proposalID + "/deposits", key: "deposits"},
		{name: "gov_votes", path: "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes", key: "votes"},
		{name: "slash_signing_infos", path: "/cosmos/slashing/v1beta1/signing_infos", key: "info"},
		{name: "staking_delegations", path: "/cosmos/staking/v1beta1/delegations/" + wallets[0].Address, key: "delegation_responses"},
		{name: "staking_redelegations", path: "/cosmos/staking/v1beta1/delegators/" + wallets[0].Address + "/redelegations", key: "redelegation_responses"},
		{name: "staking_unbonding", path: "/cosmos/staking/v1beta1/delegators/" + wallets[0].Address + "/unbonding_delegations", key: "unbonding_responses"},
		{name: "staking_del_validators", path: "/cosmos/staking/v1beta1/delegators/" + wallets[0].Address + "/validators", key: "validators"},
		{name: "staking_validators", path: "/cosmos/staking/v1beta1/validators", key: "validators"},
		{name: "staking_val_delegations", path: "/cosmos/staking/v1beta1/validators/" + wallets[0].ValoperAddress + "/delegations", key: "delegation_responses"},
		{name: "staking_val_unbonding", path: "/cosmos/staking/v1beta1/validators/" + wallets[0].ValoperAddress + "/unbonding_delegations", key: "unbonding_responses"},
		{name: "tm_validatorsets", path: "/cosmos/base/tendermint/v1beta1/validatorsets/latest", key: "validators"},
	}

	for _, tt := range tests {
		tt := tt
		s.Run("API "+tt.name, func() {
			endpoint := s.Chain.GetHostAPIAddress() + tt.path
			resp, err := http.Get(endpoint)
			s.Require().NoError(err)
			defer resp.Body.Close()
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			body := map[string]interface{}{}
			err = json.NewDecoder(resp.Body).Decode(&body)
			s.Require().NoError(err)
			s.Require().Contains(body, tt.key)
		})
	}
}

func (s *EndpointsSuite) TestRPCEndpoints() {
	blockEventsKey := "begin_block_events"
	if semver.Compare(s.Chain.GetNode().GetBuildInformation(s.GetContext()).CosmosSdkVersion, "v0.50.0") >= 0 {
		blockEventsKey = "finalize_block_events"
	}
	tests := []struct {
		name string
		path string
		key  string
	}{
		{name: "abci_info", path: "/abci_info", key: "response"},
		{name: "block", path: "/block", key: "block"},
		{name: "block_results", path: "/block_results", key: blockEventsKey},
		{name: "blockchain", path: "/blockchain", key: "block_metas"},
		{name: "commit", path: "/commit", key: "signed_header"},
		{name: "consensus_params", path: "/consensus_params", key: "consensus_params"},
		{name: "consensus_state", path: "/consensus_state", key: "round_state"},
		{name: "dump_consensus_state", path: "/dump_consensus_state", key: "round_state"},
		{name: "genesis_chunked", path: "/genesis_chunked", key: "chunk"},
		{name: "net_info", path: "/net_info", key: "peers"},
		{name: "num_unconfirmed_txs", path: "/num_unconfirmed_txs", key: "n_txs"},
		{name: "unconfirmed_txs", path: "/unconfirmed_txs", key: "n_txs"},
		{name: "status", path: "/status", key: "node_info"},
		{name: "validators", path: "/validators", key: "validators"},
	}
	for _, tt := range tests {
		tt := tt
		s.Run("RPC "+tt.name, func() {
			endpoint := s.Chain.GetHostRPCAddress() + tt.path
			resp, err := http.Get(endpoint)
			s.Require().NoError(err)
			defer resp.Body.Close()
			s.Require().Equal(http.StatusOK, resp.StatusCode)
			body := map[string]interface{}{}
			err = json.NewDecoder(resp.Body).Decode(&body)
			s.Require().NoError(err)
			s.Require().Contains(body, "result")
			s.Require().Contains(body["result"], tt.key)
		})
	}
}

func TestEndpoints(t *testing.T) {
	s := &EndpointsSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{UpgradeOnSetup: true}),
	}
	suite.Run(t, s)
}
