package e2e

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/gaia/v24/tests/e2e/common"
	"github.com/cosmos/gaia/v24/tests/e2e/msg"
	"github.com/cosmos/gaia/v24/tests/e2e/query"
)

func (s *IntegrationTestSuite) testStoreWasmLightClient() {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	err := msg.WriteStoreWasmLightClientProposal(s.Resources.ChainA)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalStoreWasmLightClientFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Store wasm light client code")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ibc.lightclients.wasm.v1.MsgStoreCode", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After StoreWasmLightClient proposal")

			res, err := query.IbcWasmChecksums(chainEndpoint)
			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(1, len(res))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) testCreateWasmLightClient() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	valIdx := 0
	val := s.Resources.ChainA.Validators[valIdx]
	address, _ := val.KeyInfo.GetAddress()
	sender := address.String()

	clientState := `{"@type":"/ibc.lightclients.wasm.v1.ClientState","data":"ZG9lc250IG1hdHRlcg==","checksum":"O45STPnbLLar4DtFwDx0dE6tuXQW5XTKPHpbjaugun4=","latest_height":{"revision_number":"0","revision_height":"7795583"}}`
	consensusState := `{"@type":"/ibc.lightclients.wasm.v1.ConsensusState","data":"ZG9lc250IG1hdHRlcg=="}`

	s.CreateClient(ctx, s.Resources.ChainA, clientState, consensusState, sender, valIdx)
	s.AddWasmClientCounterparty(ctx, s.Resources.ChainA, sender, valIdx)
}
