package tx

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"time"
)

func (h *Helper) execVestingTx( //nolint:unused

	c *common.Chain,
	method string,
	args []string,
	opt ...common.FlagOption,
) {
	opts := common.ApplyOptions(c.Id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("%s - Executing gaiad %s with %v", c.Id, method, args)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		method,
		"-y",
	}
	gaiaCommand = append(gaiaCommand, args...)

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, h.CommonHelper.DefaultExecValidation(c, 0))
	h.Suite.T().Logf("successfully %s with %v", method, args)
}

func (h *Helper) ExecCreatePeriodicVestingAccount( //nolint:unused

	c *common.Chain,
	address,
	jsonPath string,
	opt ...common.FlagOption,
) {
	h.Suite.T().Logf("Executing gaiad create periodic vesting account %s", c.Id)
	h.execVestingTx(c, "create-periodic-vesting-account", []string{address, jsonPath}, opt...)
	h.Suite.T().Logf("successfully created periodic vesting account %s with %s", address, jsonPath)
}
