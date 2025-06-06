package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v25/tests/e2e/common"
)

func (h *TestingSuite) ExecDelegate(c *common.Chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking delegate %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"delegate",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"), // default 200_000 is not enough
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully delegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (h *TestingSuite) ExecUnbondDelegation(c *common.Chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking unbond %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"unbond",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--gas=300000", // default 200_000 is not enough; gas fees are higher when unbonding is done after LSM operations
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully undelegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (h *TestingSuite) ExecCancelUnbondingDelegation(c *common.Chain, valIdx int, amount, valOperAddress, creationHeight, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking cancel-unbond %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"cancel-unbond",
		valOperAddress,
		amount,
		creationHeight,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully canceled unbonding %s to %s", delegatorAddr, amount, valOperAddress)
}

func (h *TestingSuite) ExecRedelegate(c *common.Chain, valIdx int, amount, originalValOperAddress,
	newValOperAddress, delegatorAddr, home, delegateFees string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking redelegate %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"redelegate",
		originalValOperAddress,
		newValOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "350000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully redelegated %s from %s to %s", delegatorAddr, amount, originalValOperAddress, newValOperAddress)
}

func (h *TestingSuite) ExecuteValidatorBond(c *common.Chain, valIdx int, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking validator-bond %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"validator-bond",
		valOperAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully executed validator bond tx to %s", delegatorAddr, valOperAddress)
}
