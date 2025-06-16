package tx

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/gaia/v25/tests/e2e/common"
)

func (h *TestingSuite) SendIBC(c *common.Chain, valIdx int, sender, recipient, token, fees, note, channel string, absoluteTimeout *int64, expErr bool) {
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
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		// fmt.Sprintf("--%s=%s", flags.FlagNote, note),
		fmt.Sprintf("--memo=%s", note),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}...)

	h.Suite.T().Logf("sending %s from %s (%s) to %s (%s) with memo %s", token, h.Resources.ChainA.ID, sender, h.Resources.ChainB.ID, recipient, note)
	if expErr {
		h.ExecuteGaiaTxCommand(ctx, c, ibcCmd, valIdx, h.ExpectErrExecValidation(c, valIdx, true))
		h.Suite.T().Log("unsuccessfully sent IBC tokens")
	} else {
		h.ExecuteGaiaTxCommand(ctx, c, ibcCmd, valIdx, h.DefaultExecValidation(c, valIdx))
		h.Suite.T().Log("successfully sent IBC tokens")
	}
}
