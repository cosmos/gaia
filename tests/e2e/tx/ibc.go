package tx

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/gaia/v27/tests/e2e/common"
)

func (h *TestingSuite) AddWasmClientCounterparty(ctx context.Context, c *common.Chain, sender string, valIdx int) {
	cmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"ibc",
		"client",
		"add-counterparty",
		common.V2TransferClient,
		common.CounterpartyID,
		"aWJj",
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("Adding wasm light client counterparty on chain %s", c.ID)
	h.ExecuteGaiaTxCommand(ctx, h.Resources.ChainA, cmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully added wasm light client counterparty")
}

func (h *TestingSuite) CreateClient(ctx context.Context, c *common.Chain, clientState string, consensusState string, sender string, valIdx int) {
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
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("Creating wasm light client on chain %s", c.ID)
	h.ExecuteGaiaTxCommand(ctx, c, cmd, valIdx, h.DefaultExecValidation(h.Resources.ChainA, valIdx))
	h.Suite.T().Log("successfully created wasm light client")
}
