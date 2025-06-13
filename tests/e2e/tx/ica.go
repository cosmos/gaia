package tx

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/gaia/v25/tests/e2e/common"
)

func (h *TestingSuite) RegisterICAAccount(c *common.Chain, valIdx int, sender, connectionID, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	version := string(types.ModuleCdc.MustMarshalJSON(&types.Metadata{
		Version:                types.Version,
		ControllerConnectionId: connectionID,
		HostConnectionId:       connectionID,
		Encoding:               types.EncodingProtobuf,
		TxType:                 types.TxTypeSDKMultiMsg,
	}))

	icaCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"interchain-accounts",
		"controller",
		"register",
		connectionID,
		fmt.Sprintf("--version=%s", version),
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--gas=250000", // default 200_000 is not enough; gas fees increased after adding IBC fee middleware
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	h.Suite.T().Logf("%s registering ICA account on host chain %s", sender, h.Resources.ChainB.ID)
	h.ExecuteGaiaTxCommand(ctx, c, icaCmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent register ICA account tx")
}

func (h *TestingSuite) SendICATransaction(c *common.Chain, valIdx int, sender, connectionID, packetMsgPath, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	icaCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"interchain-accounts",
		"controller",
		"send-tx",
		connectionID,
		packetMsgPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	h.Suite.T().Logf("%s sending ICA transaction to the host chain %s", sender, h.Resources.ChainB.ID)
	h.ExecuteGaiaTxCommand(ctx, c, icaCmd, valIdx, h.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent ICA transaction")
}

func (h *TestingSuite) BuildICASendTransactionFile(cdc codec.Codec, msgs []proto.Message, outputBaseDir string) {
	data, err := types.SerializeCosmosTx(cdc, msgs, types.EncodingProtobuf)
	h.Suite.Require().NoError(err)

	sendICATransaction := types.InterchainAccountPacketData{
		Type: types.EXECUTE_TX,
		Data: data,
	}

	sendICATransactionBody, err := json.MarshalIndent(sendICATransaction, "", " ")
	h.Suite.Require().NoError(err)

	outputPath := filepath.Join(outputBaseDir, "config", common.ICASendTransactionFileName)
	err = common.WriteFile(outputPath, sendICATransactionBody)
	h.Suite.Require().NoError(err)
}
