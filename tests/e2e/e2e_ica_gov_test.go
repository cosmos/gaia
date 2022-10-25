package e2e

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestICA_3_Gov() {
	s.Run("test ica transactions", func() {
		var (
			portID    = "1317/tcp"
			resourceA = s.valResources[s.chainA.id][0]
			resourceB = s.valResources[s.chainB.id][0]
			chainAAPI = fmt.Sprintf("http://%s", resourceA.GetHostPort(portID))
			chainBAPI = fmt.Sprintf("http://%s", resourceB.GetHostPort(portID))
		)

		// step 1: get ica addr
		icaOwnerAddr, err := s.chainA.genesisAccounts[icaOwnerAccountIndex].keyInfo.GetAddress()
		s.Require().NoError(err)
		icaOwner := icaOwnerAddr.String()

		var ica string
		s.Require().Eventually(
			func() bool {
				ica, err = queryICAaddr(chainAAPI, icaOwner, icaConnectionID)
				s.Require().NoError(err)

				return err == nil && ica != ""
			},
			time.Minute,
			5*time.Second,
		)

		// step 2: fund ica, send tokens from chain b val to ica on chain b
		senderAddr, err := s.chainB.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddr.String()

		s.execBankSend(s.chainB, 0, sender, ica, tokenAmount.String(), fees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterSenderICAbalance, err := getSpecificBalance(chainBAPI, ica, uatomDenom)
				s.Require().NoError(err)
				return afterSenderICAbalance.IsEqual(tokenAmount)
			},
			time.Minute,
			5*time.Second,
		)

		receiver := sender
		var beforeICASendReceiverBalance sdk.Coin
		s.Require().Eventually(
			func() bool {
				beforeICASendReceiverBalance, err = getSpecificBalance(chainBAPI, receiver, uatomDenom)
				s.Require().NoError(err)

				return !beforeICASendReceiverBalance.IsNil()
			},
			time.Minute,
			5*time.Second,
		)

		// step 3: prepare ica bank send json
		sendamt := sdk.NewCoin(uatomDenom, math.NewInt(100000))
		txCmd := []string{
			gaiadBinary,
			txCommand,
			banktypes.ModuleName,
			"send",
			ica,
			receiver,
			sendamt.String(),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, ica),
			fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainA.id),
			"--keyring-backend=test",
		}
		path := filepath.Join(s.chainA.validators[0].configDir(), "config", "ica_bank_send.json")
		s.writeICAtx(txCmd, path)

		// step 4: ica sends some tokens from ica to val on chain b
		s.submitICAtx(icaOwner, icaConnectionID, configFile("ica_bank_send.json"))

		s.Require().Eventually(
			func() bool {
				afterICASendReceiverBalance, err := getSpecificBalance(chainBAPI, receiver, uatomDenom)
				s.Require().NoError(err)

				return afterICASendReceiverBalance.Sub(beforeICASendReceiverBalance).IsEqual(sendamt)
			},
			time.Minute,
			5*time.Second,
		)

		// repeat step3: prepare ica ibc send
		sendIBCamt := math.NewInt(10)
		icaIBCsendCmd := []string{
			gaiadBinary,
			txCommand,
			"ibc-transfer",
			"transfer",
			"transfer",
			icaChannelID,
			icaOwner,
			sendIBCamt.String() + uatomDenom,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, ica),
			fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainB.id),
			"--keyring-backend=test",
		}

		path = filepath.Join(s.chainA.validators[0].configDir(), "config", "ica_ibc_send.json")
		s.writeICAtx(icaIBCsendCmd, path)

		s.submitICAtx(icaOwner, icaConnectionID, configFile("ica_ibc_send.json"))

		var balances sdk.Coins
		s.Require().Eventually(
			func() bool {
				balances, err = queryGaiaAllBalances(chainAAPI, icaOwner)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)

		var ibcAmt math.Int
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcAmt = c.Amount
				break
			}
		}

		s.Require().Equal(sendIBCamt, ibcAmt)

		// todo add ica delegation after delegation e2e merged
	})
}
