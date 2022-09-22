package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ory/dockertest/v3/docker"
)

func (s *IntegrationTestSuite) execBankSend(c *chain, valIdx int, from, to, amt, fees string, expectErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

	gaiaCommand := []string{
		gaiadBinary,
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
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

func (s *IntegrationTestSuite) execWithdrawAllRewards(c *chain, valIdx int, payee, fees string, expectErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	gaiaCommand := []string{
		gaiadBinary,
		"tx",
		"distribution",
		"withdraw-all-rewards",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, payee),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}

func (s *IntegrationTestSuite) execDistributionFundCommunityPool(c *chain, valIdx int, from, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx distribution fund-community-pool on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully funded community pool")
}

func (s *IntegrationTestSuite) execGovSubmitLegacyGovProposal(c *chain, valIdx int, submitterAddr, govProposalPath, fees, govProposalSubType string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov submit-legacy-proposal on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully submitted legacy proposal")
}

func (s *IntegrationTestSuite) execGovDepositProposal(c *chain, valIdx int, submitterAddr string, proposalId int, amount, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov deposit on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully deposited proposal %d", proposalId)
}

func (s *IntegrationTestSuite) execGovVoteProposal(c *chain, valIdx int, submitterAddr string, proposalId int, vote, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov vote on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully voted on proposal %d", proposalId)
}

func (s *IntegrationTestSuite) execGovWeightedVoteProposal(c *chain, valIdx int, submitterAddr string, proposalId int, vote, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov vote on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully voted on proposal %d", proposalId)
}

func (s *IntegrationTestSuite) execGovSubmitProposal(c *chain, valIdx int, submitterAddr, govProposalPath, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx gov submit-proposal on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully submitted proposal %s", govProposalPath)
}

func (s *IntegrationTestSuite) execCreateGroup(c *chain, valIdx int, adminAddr, metadata, groupMembersPath, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group create-group on chain %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully created group: %s", adminAddr, groupMembersPath)
}

func (s *IntegrationTestSuite) execUpdateGroupMembers(c *chain, valIdx int, adminAddr, groupId, groupMembersPath, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group update-group-members %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully updated group members: %s", adminAddr, groupMembersPath)
}

func (s *IntegrationTestSuite) executeCreateGroupPolicy(c *chain, valIdx int, adminAddr, groupId, metadata, policyFile, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group create-group-policy %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully created group policy: %s", adminAddr, policyFile)
}

func (s *IntegrationTestSuite) executeSubmitGroupProposal(c *chain, valIdx int, fromAddress, proposalPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group submit-proposal %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully submited group proposal: %s", fromAddress, proposalPath)
}

func (s *IntegrationTestSuite) executeVoteGroupProposal(c *chain, valIdx int, proposalId, voterAddress, voteOption, metadata string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group vote %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully voted %s on proposal: %s", voterAddress, voteOption, proposalId)
}

func (s *IntegrationTestSuite) executeExecGroupProposal(c *chain, valIdx int, proposalId, proposerAddress string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group exec %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed proposal: %s", proposerAddress, proposalId)
}

func (s *IntegrationTestSuite) executeUpdateGroupAdmin(c *chain, valIdx int, admin, groupId, newAdmin string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx group update-group-admin %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully updated group admin from %s to %s", admin, newAdmin)
}

func (s *IntegrationTestSuite) executeGKeysAddCommand(c *chain, valIdx int, name string, home string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	gaiaCommand := []string{
		gaiadBinary,
		"keys",
		"add",
		name,
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--keyring-backend=test",
		"--output=json",
	}

	var addrRecord AddressResponse
	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		if err := json.Unmarshal(stdOut, &addrRecord); err != nil {
			return false
		}
		return strings.Contains(addrRecord.Address, "cosmos")
	})
	return addrRecord.Address
}

func (s *IntegrationTestSuite) executeKeysList(c *chain, valIdx int, home string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	gaiaCommand := []string{
		gaiadBinary,
		"keys",
		"list",
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, func([]byte, []byte) bool {
		return true
	})
}

func (s *IntegrationTestSuite) executeDelegate(c *chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx staking delegate %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully delegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (s *IntegrationTestSuite) executeRedelegate(c *chain, valIdx int, amount, originalValOperAddress,
	newValOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx staking redelegate %s", c.id)

	gaiaCommand := []string{
		gaiadBinary,
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

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully redelegated %s from %s to %s", delegatorAddr, amount, originalValOperAddress, newValOperAddress)
}

func (s *IntegrationTestSuite) execSetWithrawAddress(
	c *chain,
	valIdx int,
	fees,
	delegatorAddress,
	newWithdrawalAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Setting distribution withdrawal address on chain %s for %s to %s", c.id, delegatorAddress, newWithdrawalAddress)
	gaiaCommand := []string{
		gaiadBinary,
		"tx",
		"distribution",
		"set-withdraw-addr",
		newWithdrawalAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully set new distribution withdrawal address for %s to %s", delegatorAddress, newWithdrawalAddress)
}

func (s *IntegrationTestSuite) execWithdrawReward(
	c *chain,
	valIdx int,
	delegatorAddress,
	validatorAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Withdrawing distribution rewards on chain %s for delegator %s from %s validator", c.id, delegatorAddress, validatorAddress)
	gaiaCommand := []string{
		gaiadBinary,
		"tx",
		"distribution",
		"withdraw-rewards",
		validatorAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300uatom"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagGasAdjustment, "1.5"),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully withdrew distribution rewards for delegator %s from validator %s", delegatorAddress, validatorAddress)
}

// register a ica on chainB from registrant on chainA
func (s *IntegrationTestSuite) submitICAtx(owner, connectionID, txJsonPath string) {
	fee := sdk.NewCoin(uatomDenom, math.NewInt(930000))
	s.T().Logf("register an interchain account on chain %s for %s from chain %s", s.chainB.id, owner, s.chainA.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	submitTX := []string{
		gaiadBinary,
		"tx",
		"icamauth",
		"submit",
		txJsonPath,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner),
		fmt.Sprintf("--%s=%s", "connection-id", connectionID),
		fmt.Sprintf("--%s=%s", flags.FlagGas, flags.GasFlagAuto),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fee),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainA.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, s.chainA, submitTX, 0, s.defaultExecValidation(s.chainA, 0))

	s.T().Logf("%s submit a transaction on chain %s", owner, s.chainB.id)
}

func (s *IntegrationTestSuite) registerICA(owner, connectionID string) {
	fee := sdk.NewCoin(uatomDenom, math.NewInt(930000))
	s.T().Logf("register an interchain account on chain %s for %s from chain %s", s.chainB.id, owner, s.chainA.id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	registerICAcmd := []string{
		gaiadBinary,
		"tx",
		"icamauth",
		"register",
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner),
		fmt.Sprintf("--%s=%s", "connection-id", connectionID),
		fmt.Sprintf("--%s=%s", flags.FlagGas, flags.GasFlagAuto),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fee),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainA.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.executeGaiaTxCommand(ctx, s.chainA, registerICAcmd, 0, s.defaultExecValidation(s.chainA, 0))

	s.T().Logf("%s reigstered an interchain account on chain %s from chain %s", owner, s.chainB.id, s.chainA.id)
}

func (s *IntegrationTestSuite) executeGaiaTxCommand(ctx context.Context, c *chain, gaiaCommand []string, valIdx int, validation func([]byte, []byte) bool) {
	if validation == nil {
		validation = s.defaultExecValidation(s.chainA, 0)
	}
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
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

	stdOut := outBuf.Bytes()
	stdErr := errBuf.Bytes()
	if !validation(stdOut, stdErr) {
		s.Require().FailNowf("tx validation failed", "stdout: %s, stderr: %s",
			string(stdOut), string(stdErr))
	}
}

func (s *IntegrationTestSuite) expectErrExecValidation(chain *chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		s.Require().NoError(cdc.UnmarshalJSON(stdOut, &txResp))
		endpoint := fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
		// wait for the tx to be committed on chain
		var err error
		s.Require().Eventuallyf(
			func() bool {
				gotErr := queryGaiaTx(endpoint, txResp.TxHash) != nil
				return gotErr == expectErr
			},
			time.Minute,
			5*time.Second,
			"stdOut: %s, stdErr: %s, err: %v",
			string(stdOut), string(stdErr), err,
		)
		return true
	}
}

func (s *IntegrationTestSuite) defaultExecValidation(chain *chain, valIdx int) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		if err := cdc.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.String(), "code: 0") || txResp.Code == 0 {
			endpoint := fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
			var err error
			s.Require().Eventually(
				func() bool {
					return queryGaiaTx(endpoint, txResp.TxHash) == nil
				},
				time.Minute,
				5*time.Second,
				"stdOut: %s, stdErr: %s, err: %v",
				string(stdOut), string(stdErr), err,
			)
			return true
		}
		return false
	}
}
