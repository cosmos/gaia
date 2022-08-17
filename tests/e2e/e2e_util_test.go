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
	group "github.com/cosmos/cosmos-sdk/x/group"
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

func (s *IntegrationTestSuite) sendMsgSend(c *chain, valIdx int, from, to, amt, fees string, expectErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.id][valIdx].Container.ID,
		User:         "nonroot",
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
			gotErr := queryGaiaTx(endpoint, txResp.TxHash) != nil
			return gotErr == expectErr
		},
		time.Minute,
		5*time.Second,
		"stdout: %s, stderr: %s",
		outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) withdrawReward(c *chain, valIdx int, endpoint, payee, fees string, expectErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s withdraw-all-rewards on chain %s", payee, c.id)

	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.id][valIdx].Container.ID,
		User:         "nonroot",
		Cmd: []string{
			"gaiad",
			"tx",
			"distribution",
			"withdraw-all-rewards",
			fmt.Sprintf("--%s=%s", flags.FlagFrom, payee),
			fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
			fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
			"--keyring-backend=test",
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

	// wait for the tx to be committed on chain
	s.Require().Eventuallyf(
		func() bool {
			gotErr := queryGaiaTx(endpoint, txResp.TxHash) != nil
			return gotErr == expectErr
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

func (s *IntegrationTestSuite) execDistributionFundCommunityPool(c *chain, valIdx int, endpoint, from, amt, fees string) {
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

func (s *IntegrationTestSuite) execGovSubmitLegacyGovProposal(c *chain, valIdx int, endpoint, submitterAddr, govProposalPath, fees, govProposalSubType string) {
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

func (s *IntegrationTestSuite) execGovDepositProposal(c *chain, valIdx int, endpoint, submitterAddr string, proposalId int, amount, fees string) {
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

func (s *IntegrationTestSuite) execGovVoteProposal(c *chain, valIdx int, endpoint, submitterAddr string, proposalId int, vote, fees string) {
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

func (s *IntegrationTestSuite) execGovWeightedVoteProposal(c *chain, valIdx int, endpoint, submitterAddr string, proposalId int, vote, fees string) {
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

func (s *IntegrationTestSuite) execGovSubmitProposal(c *chain, valIdx int, endpoint, submitterAddr, govProposalPath, fees string) {
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
			time.Sleep(3 * time.Second)
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.id][valIdx].Container.ID,
				User:         "nonroot",
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

			return strings.Contains(txResp.String(), "code: 0") || txResp.Code == uint32(0)
		},
		10*time.Second,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
	endpoint = fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
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
				User:         "nonroot",
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

func (s *IntegrationTestSuite) execCreateGroup(c *chain, valIdx int, endpoint string, adminAddr string, metadata string, groupMembersPath string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group create-group on chain %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"create-group",
		adminAddr,
		metadata,
		groupMembersPath,
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully created group: %s", adminAddr, groupMembersPath)
}

func (s *IntegrationTestSuite) execUpdateGroupMembers(c *chain, valIdx int, endpoint string, adminAddr string, groupId string, groupMembersPath string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group update-group-members %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"update-group-members",
		adminAddr,
		groupId,
		groupMembersPath,
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully updated group members: %s", adminAddr, groupMembersPath)
}

func (s *IntegrationTestSuite) executeCreateGroupPolicy(c *chain, valIdx int, endpoint string, adminAddr string, groupId string, metadata string, policyFile string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group create-group-policy %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"create-group-policy",
		adminAddr,
		groupId,
		metadata,
		policyFile,
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully created group policy: %s", adminAddr, policyFile)
}

func (s *IntegrationTestSuite) executeSubmitGroupProposal(c *chain, valIdx int, endpoint string, fromAddress string, proposalPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group submit-proposal %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"submit-proposal",
		proposalPath,
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, fromAddress),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully submited group proposal: %s", fromAddress, proposalPath)
}

func (s *IntegrationTestSuite) executeVoteGroupProposal(c *chain, valIdx int, endpoint string, proposalId string, voterAddress string, voteOption string, metadata string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group vote %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"vote",
		proposalId,
		voterAddress,
		voteOption,
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		metadata,
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully voted %s on proposal: %s", voterAddress, voteOption, proposalId)
}

func (s *IntegrationTestSuite) executeExecGroupProposal(c *chain, valIdx int, endpoint string, proposalId string, proposerAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group exec %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"exec",
		proposalId,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, proposerAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully executed proposal: %s", proposerAddress, proposalId)
}

func (s *IntegrationTestSuite) executeUpdateGroupAdmin(c *chain, valIdx int, endpoint string, admin string, groupId string, newAdmin string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group update-group-admin %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"group",
		"update-group-admin",
		admin,
		groupId,
		newAdmin,
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("Successfully updated group admin from %s to %s", admin, newAdmin)
}

func (s *IntegrationTestSuite) queryGroupMembers(endpoint string, groupId int) (group.QueryGroupMembersResponse, error) {
	var res group.QueryGroupMembersResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/group_members/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupInfo(endpoint string, groupId int) (group.QueryGroupInfoResponse, error) {
	var res group.QueryGroupInfoResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/group_info/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupsbyAdmin(endpoint string, adminAddress string) (group.QueryGroupsByAdminResponse, error) {
	var res group.QueryGroupsByAdminResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/groups_by_admin/%s", endpoint, adminAddress)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupPolicies(endpoint string, groupId int) (group.QueryGroupPoliciesByGroupResponse, error) {
	var res group.QueryGroupPoliciesByGroupResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/group_policies_by_group/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupProposal(endpoint string, groupId int) (group.QueryProposalResponse, error) {
	var res group.QueryProposalResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/proposal/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupProposalByGroupPolicy(endpoint string, policyAddress string) (group.QueryProposalsByGroupPolicyResponse, error) {
	var res group.QueryProposalsByGroupPolicyResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/proposals_by_group_policy/%s", endpoint, policyAddress)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) verifyBalanceChange(endpoint string, expectedAmount sdk.Coin, recipientAddress string) {
	s.Require().Eventually(
		func() bool {
			afterAtomBalance, err := getSpecificBalance(endpoint, recipientAddress, uatomDenom)
			s.Require().NoError(err)

			return afterAtomBalance.IsEqual(expectedAmount)
		},
		20*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) executeGKeysAddCommand(c *chain, valIdx int, name string, home string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var (
		outBuf     bytes.Buffer
		errBuf     bytes.Buffer
		addrRecord AddressResponse
	)

	gaiaCommand := []string{
		"gaiad",
		"keys",
		"add",
		name,
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--keyring-backend=test",
		"--output=json",
	}

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.id][valIdx].Container.ID,
				User:         "nonroot",
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
			s.Require().NoError(json.Unmarshal(errBuf.Bytes(), &addrRecord))

			return strings.Contains(addrRecord.Address, "cosmos")
		},
		10*time.Second,
		time.Second,
		"Returned an error; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	return addrRecord.Address
}

func (s *IntegrationTestSuite) executeKeysList(c *chain, valIdx int, home string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	gaiaCommand := []string{
		"gaiad",
		"keys",
		"list",
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
	}

	s.Require().Eventually(
		func() bool {
			exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    s.valResources[c.id][valIdx].Container.ID,
				User:         "nonroot",
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
			return true
		},
		10*time.Second,
		time.Second,
		"Returned an error; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)
}

func (s *IntegrationTestSuite) executeDelegate(c *chain, valIdx int, endpoint string, amount string, valOperAddress string, delegatorAddr string, home string, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx staking delegate %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"staking",
		"delegate",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagFees, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully delegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (s *IntegrationTestSuite) executeRedelegate(c *chain, valIdx int, endpoint string, amount string, originalValOperAddress string, newValOperAddress string, delegatorAddr string, home string, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx staking redelegate %s", c.id)

	gaiaCommand := []string{
		"gaiad",
		"tx",
		"staking",
		"redelegate",
		originalValOperAddress,
		newValOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagFees, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, endpoint)
	s.T().Logf("%s successfully redelegated %s from %s to %s", delegatorAddr, amount, originalValOperAddress, newValOperAddress)
}
