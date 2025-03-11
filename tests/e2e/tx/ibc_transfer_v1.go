package tx

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"time"
)

//nolint:unparam
func (h *Helper) SendIBC(c *common.Chain, valIdx int, sender, recipient, token, fees, note, channel string, absoluteTimeout *int64, expErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"ibc-transfer",
		"transfer",
		"transfer",
		channel,
		recipient,
		token,
	}

	if absoluteTimeout != nil {
		ibcCmd = append(ibcCmd, "--absolute-timeouts")
		ibcCmd = append(ibcCmd, fmt.Sprintf("--packet-timeout-timestamp=%d", *absoluteTimeout))
	}

	ibcCmd = append(ibcCmd, []string{
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		// fmt.Sprintf("--%s=%s", flags.FlagNote, note),
		fmt.Sprintf("--memo=%s", note),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}...)

	h.Suite.T().Logf("sending %s from %s (%s) to %s (%s) with memo %s", token, h.CommonHelper.Resources.ChainA.Id, sender, h.CommonHelper.Resources.ChainB.Id, recipient, note)
	if expErr {
		h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, ibcCmd, valIdx, h.expectErrExecValidation(c, valIdx, true))
		h.Suite.T().Log("unsuccessfully sent IBC tokens")
	} else {
		h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, ibcCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
		h.Suite.T().Log("successfully sent IBC tokens")
	}
}
