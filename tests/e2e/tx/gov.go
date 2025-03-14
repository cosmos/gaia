package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
)

func (h *TestingSuite) RunGovExec(c *common.Chain, valIdx int, submitterAddr, govCommand string, proposalFlags []string, fees string, validationFunc func([]byte, []byte) bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	validateResponse := h.DefaultExecValidation(c, valIdx)
	if validationFunc != nil {
		validateResponse = validationFunc
	}

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		govCommand,
	}

	generalFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "50000000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	gaiaCommand = common.ConcatFlags(gaiaCommand, proposalFlags, generalFlags)
	h.Suite.T().Logf("Executing gaiad tx gov %s on chain %s", govCommand, c.ID)
	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, validateResponse)
	h.Suite.T().Logf("Successfully executed %s", govCommand)
}
