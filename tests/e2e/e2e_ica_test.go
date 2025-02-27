package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	ICASendTransactionFileName = "execute_ica_transaction.json"
	connectionID               = "connection-0"
	icaChannel                 = "channel-1"
)

func (s *IntegrationTestSuite) testICARegisterAccountAndSendTx() {
	s.Run("register_ICA_account_and_send_tx_to_chainB", func() {
		var (
			icaAccount             string
			icaAccountBalances     sdk.Coins
			recipientBalances      sdk.Coins
			recipientBalanceBefore int64
			err                    error
			ibcStakeDenom          string
		)

		address, _ := s.chainA.validators[0].keyInfo.GetAddress()
		icaOwnerAccount := address.String()
		icaOwnerPortID, _ := icatypes.NewControllerPortID(icaOwnerAccount)

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		s.registerICAAccount(s.chainA, 0, icaOwnerAccount, connectionID, standardFees.String())
		s.completeChannelHandshakeFromTry(
			s.chainA.id, s.chainB.id,
			connectionID, connectionID,
			icaOwnerPortID, icatypes.HostPortID,
			icaChannel, icaChannel)

		s.Require().Eventually(
			func() bool {
				icaAccount, _ = queryICAAccountAddress(chainAAPIEndpoint, icaOwnerAccount, connectionID)
				return icaAccount != ""
			},
			time.Minute,
			5*time.Second,
		)

		tokenAmount := 3300000000
		s.sendIBC(s.chainA, 0, icaOwnerAccount, icaAccount, strconv.Itoa(tokenAmount)+uatomDenom, standardFees.String(), "", false)

		pass := s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferPort, transferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				icaAccountBalances, err = queryGaiaAllBalances(chainBAPIEndpoint, icaAccount)
				s.Require().NoError(err)
				return icaAccountBalances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range icaAccountBalances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal((int64(tokenAmount)), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)

		address, _ = s.chainB.validators[0].keyInfo.GetAddress()
		recipientB := address.String()

		s.Require().Eventually(
			func() bool {
				recipientBalances, err = queryGaiaAllBalances(chainBAPIEndpoint, recipientB)
				s.Require().NoError(err)
				return recipientBalances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range recipientBalances {
			if c.Denom == ibcStakeDenom {
				recipientBalanceBefore = c.Amount.Int64()
				break
			}
		}

		amountToICASend := int64(tokenAmount / 3)
		bankSendMsg := banktypes.NewMsgSend(
			sdk.MustAccAddressFromBech32(icaAccount),
			sdk.MustAccAddressFromBech32(recipientB),
			sdk.NewCoins(sdk.NewCoin(ibcStakeDenom, math.NewInt(amountToICASend))))

		s.buildICASendTransactionFile(cdc, []proto.Message{bankSendMsg}, s.chainA.validators[0].configDir())
		s.sendICATransaction(s.chainA, 0, icaOwnerAccount, connectionID, configFile(ICASendTransactionFileName), standardFees.String())
		s.Require().True(s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, icaOwnerPortID, icaChannel))

		s.Require().Eventually(
			func() bool {
				recipientBalances, err = queryGaiaAllBalances(chainBAPIEndpoint, recipientB)
				s.Require().NoError(err)
				return recipientBalances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)

		for _, c := range recipientBalances {
			if c.Denom == ibcStakeDenom {
				s.Require().Equal(recipientBalanceBefore+amountToICASend, c.Amount.Int64())
				break
			}
		}
	})
}

func (s *IntegrationTestSuite) registerICAAccount(c *chain, valIdx int, sender, connectionID, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	version := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: connectionID,
		HostConnectionId:       connectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	icaCmd := []string{
		gaiadBinary,
		txCommand,
		"interchain-accounts",
		"controller",
		"register",
		connectionID,
		fmt.Sprintf("--version=%s", version),
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--gas=250000", // default 200_000 is not enough; gas fees increased after adding IBC fee middleware
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("%s registering ICA account on host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, icaCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent register ICA account tx")
}

func (s *IntegrationTestSuite) sendICATransaction(c *chain, valIdx int, sender, connectionID, packetMsgPath, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	icaCmd := []string{
		gaiadBinary,
		txCommand,
		"interchain-accounts",
		"controller",
		"send-tx",
		connectionID,
		packetMsgPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("%s sending ICA transaction to the host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, icaCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent ICA transaction")
}

func (s *IntegrationTestSuite) buildICASendTransactionFile(cdc codec.Codec, msgs []proto.Message, outputBaseDir string) {
	data, err := icatypes.SerializeCosmosTx(cdc, msgs, icatypes.EncodingProtobuf)
	s.Require().NoError(err)

	sendICATransaction := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	sendICATransactionBody, err := json.MarshalIndent(sendICATransaction, "", " ")
	s.Require().NoError(err)

	outputPath := filepath.Join(outputBaseDir, "config", ICASendTransactionFileName)
	err = writeFile(outputPath, sendICATransactionBody)
	s.Require().NoError(err)
}
