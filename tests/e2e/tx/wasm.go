package tx

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
)

// todo: change this to a query instead of a command when https://github.com/CosmWasm/wasmd/issues/2147 is fixed
func (h *TestingSuite) QueryBuildAddress(ctx context.Context, c *common.Chain, valIdx int, codeHash, creatorAddress, saltHexEncoded string,
) (res string) {
	cmd := []string{
		common.GaiadBinary,
		common.QueryCommand,
		"wasm",
		"build-address",
		codeHash,
		creatorAddress,
		saltHexEncoded,
	}

	h.ExecuteGaiaTxCommand(ctx, c, cmd, valIdx, func(stdOut []byte, stdErr []byte) bool {
		h.Suite.Require().NoError(yaml.Unmarshal(stdOut, &res))
		return true
	})
	return res
}

func (h *TestingSuite) StoreWasm(ctx context.Context, c *common.Chain, valIdx int, sender, wasmPath string) string {
	storeCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"store",
		wasmPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--gas=5000000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("%s storing wasm on host chain %s", sender, h.Resources.ChainB.ID)
	h.ExecuteGaiaTxCommand(ctx, c, storeCmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent store wasm tx")
	h.TestCounters.ContractsCounter++
	return strconv.Itoa(h.TestCounters.ContractsCounter)
}

func (h *TestingSuite) InstantiateWasm(ctx context.Context, c *common.Chain, valIdx int, sender, codeID,
	msg, label string,
) string {
	storeCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"instantiate",
		codeID,
		msg,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--label=%s", label),
		"--no-admin",
		"--gas=500000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("%s instantiating wasm on host chain %s", sender, h.Resources.ChainB.ID)
	h.ExecuteGaiaTxCommand(ctx, c, storeCmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent instantiate wasm tx")
	chainEndpoint := fmt.Sprintf("http://%s", h.Resources.ValResources[c.ID][0].GetHostPort("1317/tcp"))
	address, err := query.WasmContractAddress(chainEndpoint, sender, h.TestCounters.ContractsCounterPerSender[sender])
	h.Suite.Require().NoError(err)
	h.TestCounters.ContractsCounterPerSender[sender]++
	return address
}

func (h *TestingSuite) Instantiate2Wasm(ctx context.Context, c *common.Chain, valIdx int, sender, codeID,
	msg, salt, label string,
) string {
	storeCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"instantiate2",
		codeID,
		msg,
		salt,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--label=%s", label),
		"--no-admin",
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("%s instantiating wasm on host chain %s", sender, h.Resources.ChainB.ID)

	h.ExecuteGaiaTxCommand(ctx, c, storeCmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent instantiate2 wasm tx")
	chainEndpoint := fmt.Sprintf("http://%s", h.Resources.ValResources[c.ID][0].GetHostPort("1317/tcp"))
	address, err := query.WasmContractAddress(chainEndpoint, sender, h.TestCounters.ContractsCounterPerSender[sender])
	h.Suite.Require().NoError(err)
	h.TestCounters.ContractsCounterPerSender[sender]++
	return address
}

func (h *TestingSuite) ExecuteWasm(ctx context.Context, c *common.Chain, valIdx int, sender, addr, msg string) {
	execCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"execute",
		addr,
		msg,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	h.Suite.T().Logf("%s executing wasm on host chain %s", sender, h.Resources.ChainB.ID)
	h.ExecuteGaiaTxCommand(ctx, c, execCmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent execute wasm tx")
}
