package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/gaia/v23/tests/e2e/data"
)

const (
	proposalStoreWasmLightClientFilename = "proposal_store_wasm_light_client.json"
)

func (s *IntegrationTestSuite) testStoreWasmLightClient() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeStoreWasmLightClientProposal(s.chainA)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalStoreWasmLightClientFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Store wasm light client code")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ibc.lightclients.wasm.v1.MsgStoreCode", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After StoreWasmLightClient proposal")

			res, err := queryIbcWasmChecksums(chainEndpoint)
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
	val := s.chainA.validators[valIdx]
	address, _ := val.keyInfo.GetAddress()
	sender := address.String()

	clientState := `{"@type":"/ibc.lightclients.wasm.v1.ClientState","data":"ZG9lc250IG1hdHRlcg==","checksum":"O45STPnbLLar4DtFwDx0dE6tuXQW5XTKPHpbjaugun4=","latest_height":{"revision_number":"0","revision_height":"7795583"}}`
	consensusState := `{"@type":"/ibc.lightclients.wasm.v1.ConsensusState","data":"ZG9lc250IG1hdHRlcg=="}`

	cmd := []string{
		gaiadBinary,
		txCommand,
		"ibc",
		"client",
		"create",
		clientState,
		consensusState,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainA.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Creating wasm light client on chain %s", s.chainA.id)
	s.executeGaiaTxCommand(ctx, s.chainA, cmd, valIdx, s.defaultExecValidation(s.chainA, valIdx))
	s.T().Log("successfully created wasm light client")
}

func (s *IntegrationTestSuite) writeStoreWasmLightClientProposal(c *chain) {
	template := `
	{
		"messages": [
			{
			"@type": "/ibc.lightclients.wasm.v1.MsgStoreCode",
			"signer": "%s",
			"wasm_byte_code": "%s"
			}
		],
		"metadata": "AQ==",
		"deposit": "100uatom",
		"title": "Store wasm light client code",
		"summary": "e2e-test storing wasm light client code"
	   }`
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		data.WasmDummyLightClient,
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalStoreWasmLightClientFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}
