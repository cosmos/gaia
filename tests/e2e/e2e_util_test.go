package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/ory/dockertest/v3/docker"
)

func (s *IntegrationTestSuite) connectIBCChains() {
	s.T().Logf("connecting %s and %s chains via IBC", s.chainA.id, s.chainB.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"create",
			"channel",
			s.chainA.id,
			s.chainB.id,
			"--port-a=transfer",
			"--port-b=transfer",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed connect chains; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.Require().Containsf(
		errBuf.String(),
		"successfully opened init channel",
		"failed to connect chains via IBC: %s", errBuf.String(),
	)

	s.T().Logf("connected %s and %s chains via IBC", s.chainA.id, s.chainB.id)
}

func (s *IntegrationTestSuite) sendMsgSend(c *chain, valIdx int, from, to, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.id][valIdx].Container.ID,
		User:         "root",
		Cmd: []string{
			"gaiad",
			"tx",
			"bank",
			"send",
			from,
			to,
			amt,
			fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
			fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
			"--keyring-backend=test",
			"--broadcast-mode=sync",
			"--output=json",
			"-y",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(err, "stdout: %s, stderr: %s", outBuf.String(), errBuf.String())

	var txResp sdk.TxResponse
	s.Require().NoError(cdc.UnmarshalJSON(outBuf.Bytes(), &txResp))

	endpoint := fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))

	// wait for the tx to be committed on chain
	s.Require().Eventuallyf(
		func() bool {
			return queryGaiaTx(endpoint, txResp.TxHash) == nil
		},
		time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) sendIBC(srcChainID, dstChainID, recipient string, token sdk.Coin) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s from %s to %s (%s)", token, srcChainID, dstChainID, recipient)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.hermesResource.Container.ID,
		User:         "root",
		Cmd: []string{
			"hermes",
			"tx",
			"raw",
			"ft-transfer",
			dstChainID,
			srcChainID,
			"transfer",  // source chain port ID
			"channel-0", // since only one connection/channel exists, assume 0
			token.Amount.String(),
			fmt.Sprintf("--denom=%s", token.Denom),
			fmt.Sprintf("--receiver=%s", recipient),
			"--timeout-height-offset=1000",
		},
	})
	s.Require().NoError(err)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoErrorf(
		err,
		"failed to send IBC tokens; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	s.T().Log("successfully sent IBC tokens")
}

func (s *IntegrationTestSuite) execDistributionFundCommunityPool(c *chain, valIdx int, endpoint string, from, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx distribution fund-community-pool on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"distribution",
		"fund-community-pool",
		amt,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully funded community pool")
}

func (s *IntegrationTestSuite) execGovSubmitLegacyGovProposal(c *chain, valIdx int, endpoint string, submitterAddr string, govProposalPath string, fees string, govProposalSubType string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov submit-legacy-proposal on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"gov",
		"submit-legacy-proposal",
		govProposalSubType,
		govProposalPath,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully submitted legacy proposal")
}

func (s *IntegrationTestSuite) execGovDepositProposal(c *chain, valIdx int, endpoint string, submitterAddr string, proposalId int, amount string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov deposit on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"gov",
		"deposit",
		fmt.Sprintf("%d", proposalId),
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully deposited proposal %d", proposalId)
}

func (s *IntegrationTestSuite) execGovVoteProposal(c *chain, valIdx int, endpoint string, submitterAddr string, proposalId int, vote string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov vote on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"gov",
		"vote",
		fmt.Sprintf("%d", proposalId),
		vote,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully voted on proposal %d", proposalId)
}

func (s *IntegrationTestSuite) execGovWeightedVoteProposal(c *chain, valIdx int, endpoint string, submitterAddr string, proposalId int, vote string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov vote on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"gov",
		"weighted-vote",
		fmt.Sprintf("%d", proposalId),
		vote,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully voted on proposal %d", proposalId)
}

func (s *IntegrationTestSuite) execGovSubmitProposal(c *chain, valIdx int, endpoint string, submitterAddr string, govProposalPath string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov submit-proposal on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"gov",
		"submit-proposal",
		govProposalPath,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully submitted proposal %s", govProposalPath)
}

func (s *IntegrationTestSuite) executeGaiaTxCommand(ctx context.Context, c *chain, gaiaCommand []string, valIdx int, endpoint string) {
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
		txResp sdk.TxResponse
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.id][valIdx].Container.ID,
				User:         "root",
				Cmd:          gaiaCommand,
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			s.Require().NoError(err)

			s.Require().NoError(cdc.UnmarshalJSON(outBuf.Bytes(), &txResp))
			return strings.Contains(txResp.String(), "code: 0")
		},
		5*time.Second,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
	s.Require().Eventually(
		func() bool {
			return queryGaiaTx(endpoint, txResp.TxHash) == nil
		},
		time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) queryGovProposal(endpoint string, proposalId uint64) (govv1beta1.QueryProposalResponse, error) {
	var emptyProp govv1beta1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%d", endpoint, proposalId)
	resp, err := http.Get(path)

	if err != nil {
		s.T().Logf("This is the err: %s", err.Error())
	}

	if err != nil {
		return emptyProp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		s.T().Logf("This is the err: %s", err.Error())
	}
	if err != nil {

		return emptyProp, err
	}
	s.T().Logf("This is the body: %s", body)
	var govProposalResp govv1beta1.QueryProposalResponse

	if err := cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return emptyProp, err
	}
	s.T().Logf("This is the gov response: %s", govProposalResp)

	return govProposalResp, nil
}

func (s *IntegrationTestSuite) getLatestBlockHeight(c *chain, valIdx int) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	type status struct {
		LatestHeight string `json:"latest_block_height"`
	}

	type syncInfo struct {
		SyncInfo status `json:"SyncInfo"`
	}

	var (
		outBuf        bytes.Buffer
		errBuf        bytes.Buffer
		block         syncInfo
		currentHeight int
	)

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.id][valIdx].Container.ID,
				User:         "root",
				Cmd:          []string{"gaiad", "status"},
			})
			s.Require().NoError(err)

			err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			s.Require().NoError(err)
			s.Require().NoError(json.Unmarshal(errBuf.Bytes(), &block))

			currentHeight, err = strconv.Atoi(block.SyncInfo.LatestHeight)
			s.Require().NoError(err)
			return currentHeight > 0
		},
		5*time.Second,
		time.Second,
		"Get node status returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
	return currentHeight
}
