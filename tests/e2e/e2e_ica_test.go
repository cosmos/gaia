package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
		s.sendIBC(s.chainA, 0, icaOwnerAccount, icaAccount, strconv.Itoa(tokenAmount)+uatomDenom, standardFees.String(), "", transferChannel, nil, false)

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
