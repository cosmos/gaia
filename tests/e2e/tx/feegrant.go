package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/feegrant"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/gaia/v25/tests/e2e/common"
)

func (h *TestingSuite) ExecFeeGrant(c *common.Chain, valIdx int, granter, grantee, spendLimit string, opt ...common.FlagOption) {
	opt = append(opt, common.WithKeyValue(common.FlagFrom, granter))
	opt = append(opt, common.WithKeyValue(common.FlagSpendLimit, spendLimit))
	opts := common.ApplyOptions(c.ID, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("granting %s fee from %s on chain %s", grantee, granter, c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		feegrant.ModuleName,
		"grant",
		granter,
		grantee,
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "300000"), // default 200000 isn't enough
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%s", flag, value))
	}
	h.Suite.T().Logf("running feegrant on chain: %s - Tx %v", c.ID, gaiaCommand)

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
}

func (h *TestingSuite) ExecFeeGrantRevoke(c *common.Chain, valIdx int, granter, grantee string, opt ...common.FlagOption) {
	opt = append(opt, common.WithKeyValue(common.FlagFrom, granter))
	opts := common.ApplyOptions(c.ID, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("revoking %s fee grant from %s on chain %s", grantee, granter, c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		feegrant.ModuleName,
		"revoke",
		granter,
		grantee,
		"-y",
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.DefaultExecValidation(c, valIdx))
}
