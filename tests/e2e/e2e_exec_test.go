package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ory/dockertest/v3/docker"

	icamauth "github.com/cosmos/gaia/v8/x/icamauth/types"
)

const (
	flagFrom           = "from"
	flagHome           = "home"
	flagFees           = "fees"
	flagGas            = "gas"
	flagOutput         = "output"
	flagChainID        = "chain-id"
	flagBroadcastMode  = "broadcast-mode"
	flagKeyringBackend = "keyring-backend"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command
func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyOptions(chainID string, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagChainID:        chainID,
		flagHome:           gaiaHomePath,
		flagFees:           fees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) execEncode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing gaiad encoding with %v", c.id, txPath)
	gaiaCommand := []string{
		gaiadBinary,
		txCommand,
		"encode",
		txPath,
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var encoded string
	s.executeGaiaTxCommand(ctx, c, gaiaCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		encoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.T().Logf("successfully encode with %v", txPath)
	return encoded
}

func (s *IntegrationTestSuite) execDecode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing gaiad decoding with %v", c.id, txPath)
	gaiaCommand := []string{
		gaiadBinary,
		txCommand,
		"decode",
		txPath,
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var decoded string
	s.executeGaiaTxCommand(ctx, c, gaiaCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		decoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.T().Logf("successfully decode %v", txPath)
	return decoded
}

func (s *IntegrationTestSuite) execVestingTx(
	c *chain,
	method string,
	args []string,
	opt ...flagOption,
) {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing gaiad %s with %v", c.id, method, args)
	gaiaCommand := []string{
		gaiadBinary,
		txCommand,
		vestingtypes.ModuleName,
		method,
		"-y",
	}
	gaiaCommand = append(gaiaCommand, args...)

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, 0, s.defaultExecValidation(c, 0))
	s.T().Logf("successfully %s with %v", method, args)
}

func (s *IntegrationTestSuite) execCreatePermanentLockedAccount(
	c *chain,
	address,
	amount string,
	opt ...flagOption,
) {
	s.T().Logf("Executing gaiad create a permanent locked vesting account %s", c.id)
	s.execVestingTx(c, "create-permanent-locked-account", []string{address, amount}, opt...)
	s.T().Logf("successfully created permanent locked vesting account %s with %s", address, amount)
}

func (s *IntegrationTestSuite) execCreatePeriodicVestingAccount(
	c *chain,
	address,
	jsonPath string,
	opt ...flagOption,
) {
	s.T().Logf("Executing gaiad create periodic vesting account %s", c.id)
	s.execVestingTx(c, "create-periodic-vesting-account", []string{address, jsonPath}, opt...)
	s.T().Logf("successfully created periodic vesting account %s with %s", address, jsonPath)
}

func (s *IntegrationTestSuite) execBankSend(c *chain, valIdx int, from, to, amt, fees string, expectErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.id)

	gaiaCommand := []string{
		gaiadBinary,
		txCommand,
		banktypes.ModuleName,
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
		txCommand,
		distributiontypes.ModuleName,
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
		txCommand,
		distributiontypes.ModuleName,
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
		txCommand,
		govtypes.ModuleName,
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
		txCommand,
		govtypes.ModuleName,
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
		txCommand,
		govtypes.ModuleName,
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
		txCommand,
		govtypes.ModuleName,
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
		txCommand,
		govtypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		txCommand,
		grouptypes.ModuleName,
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
		keysCommand,
		"add",
		name,
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--keyring-backend=test",
		"--output=json",
	}

	var addrRecord AddressResponse
	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		// Gaiad keys add by default returns payload to stdErr
		if err := json.Unmarshal(stdErr, &addrRecord); err != nil {
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
		keysCommand,
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
		txCommand,
		stakingtypes.ModuleName,
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
		txCommand,
		stakingtypes.ModuleName,
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

func (s *IntegrationTestSuite) getLatestBlockHeight(c *chain, valIdx int) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	type syncInfo struct {
		SyncInfo struct {
			LatestHeight string `json:"latest_block_height"`
		} `json:"SyncInfo"`
	}

	var currentHeight int
	gaiaCommand := []string{gaiadBinary, "status"}
	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		var (
			err   error
			block syncInfo
		)
		s.Require().NoError(json.Unmarshal(stdErr, &block))
		currentHeight, err = strconv.Atoi(block.SyncInfo.LatestHeight)
		s.Require().NoError(err)
		return currentHeight > 0
	})
	return currentHeight
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

func (s *IntegrationTestSuite) execSetWithdrawAddress(
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
		txCommand,
		distributiontypes.ModuleName,
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
		txCommand,
		distributiontypes.ModuleName,
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
		txCommand,
		icamauth.ModuleName,
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
		txCommand,
		icamauth.ModuleName,
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

	s.T().Logf("%s registered an interchain account on chain %s from chain %s", owner, s.chainB.id, s.chainA.id)
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
