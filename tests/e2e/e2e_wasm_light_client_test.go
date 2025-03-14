package e2e

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/msg"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
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

	cmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"ibc",
		"client",
		"create",
		clientState,
		consensusState,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.Resources.ChainA.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Creating wasm light client on chain %s", s.Resources.ChainA.ID)
	s.ExecuteGaiaTxCommand(ctx, s.Resources.ChainA, cmd, valIdx, s.DefaultExecValidation(s.Resources.ChainA, valIdx))
	s.T().Log("successfully created wasm light client")

	cmd2 := []string{
		common.GaiadBinary,
		common.TxCommand,
		"ibc",
		"client",
		"add-counterparty",
		common.V2TransferClient,
		"client-0",
		"aWJj",
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.Resources.ChainA.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Adding wasm light client counterparty on chain %s", s.Resources.ChainA.ID)
	s.ExecuteGaiaTxCommand(ctx, s.Resources.ChainA, cmd2, valIdx, s.DefaultExecValidation(s.Resources.ChainA, valIdx))
	s.T().Log("successfully added wasm light client counterparty")
}
