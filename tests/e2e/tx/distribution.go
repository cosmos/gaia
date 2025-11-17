package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
)

func (h *TestingSuite) ExecDistributionFundCommunityPool(c *common.Chain, valIdx int, from, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx distribution fund-community-pool on chain %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"fund-community-pool",
		amt,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("Successfully funded community pool")
}

func (h *TestingSuite) ExecSetWithdrawAddress(
	c *common.Chain,
	valIdx int,
	fees,
	delegatorAddress,
	newWithdrawalAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Setting distribution withdrawal address on chain %s for %s to %s", c.ID, delegatorAddress, newWithdrawalAddress)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"set-withdraw-addr",
		newWithdrawalAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("Successfully set new distribution withdrawal address for %s to %s", delegatorAddress, newWithdrawalAddress)
}

func (h *TestingSuite) ExecWithdrawReward(
	c *common.Chain,
	valIdx int,
	delegatorAddress,
	validatorAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Withdrawing distribution rewards on chain %s for delegator %s from %s validator", c.ID, delegatorAddress, validatorAddress)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"withdraw-rewards",
		validatorAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300uatom"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagGasAdjustment, "1.5"),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("Successfully withdrew distribution rewards for delegator %s from validator %s", delegatorAddress, validatorAddress)
}
