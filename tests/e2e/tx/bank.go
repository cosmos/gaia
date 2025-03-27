package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
)

func (h *TestingSuite) ExecBankSend(
	c *common.Chain,
	valIdx int,
	from,
	to,
	amt,
	fees string,
	expectErr bool,
	opt ...common.FlagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, common.WithKeyValue(common.FlagFees, fees))
	opt = append(opt, common.WithKeyValue(common.FlagFrom, from))
	opts := common.ApplyOptions(c.ID, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"send",
		from,
		to,
		amt,
		"-y",
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.expectErrExecValidation(c, valIdx, expectErr))
}

func (h *TestingSuite) ExecBankMultiSend(
	c *common.Chain,
	valIdx int,
	from string,
	to []string,
	amt string,
	fees string,
	expectErr bool,
	opt ...common.FlagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, common.WithKeyValue(common.FlagFees, fees))
	opt = append(opt, common.WithKeyValue(common.FlagFrom, from))
	opts := common.ApplyOptions(c.ID, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		"multi-send",
		from,
	}

	gaiaCommand = append(gaiaCommand, to...)
	gaiaCommand = append(gaiaCommand, amt, "-y")

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.expectErrExecValidation(c, valIdx, expectErr))
}
