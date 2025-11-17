package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
)

func (h *TestingSuite) ExecUnjail(
	c *common.Chain,
	opt ...common.FlagOption,
) {
	opts := common.ApplyOptions(c.ID, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad slashing unjail %s with options: %v", c.ID, opt)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"unjail",
		"-y",
	}

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, h.DefaultExecValidation(c, 0))
	h.Suite.T().Logf("successfully unjail with options %v", opt)
}
