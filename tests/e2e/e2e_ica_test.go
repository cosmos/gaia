package e2e

import (
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
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

		address, _ := s.commonHelper.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
		icaOwnerAccount := address.String()
		icaOwnerPortID, _ := icatypes.NewControllerPortID(icaOwnerAccount)

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainA.Id][0].GetHostPort("1317/tcp"))
		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainB.Id][0].GetHostPort("1317/tcp"))

		s.tx.RegisterICAAccount(s.commonHelper.Resources.ChainA, 0, icaOwnerAccount, common.ConnectionID, common.StandardFees.String())
		s.commonHelper.CompleteChannelHandshakeFromTry(
			s.commonHelper.Resources.ChainA.Id, s.commonHelper.Resources.ChainB.Id,
			common.ConnectionID, common.ConnectionID,
			icaOwnerPortID, icatypes.HostPortID,
			common.IcaChannel, common.IcaChannel)

		s.Require().Eventually(
			func() bool {
				icaAccount, _ = query.QueryICAAccountAddress(chainAAPIEndpoint, icaOwnerAccount, common.ConnectionID)
				return icaAccount != ""
			},
			time.Minute,
			5*time.Second,
		)

		TokenAmount := 3300000000
		s.tx.SendIBC(s.commonHelper.Resources.ChainA, 0, icaOwnerAccount, icaAccount, strconv.Itoa(TokenAmount)+common.UAtomDenom, common.StandardFees.String(), "", common.TransferChannel, nil, false)

		pass := s.commonHelper.HermesClearPacket(common.HermesConfigWithGasPrices, s.commonHelper.Resources.ChainA.Id, common.TransferPort, common.TransferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				icaAccountBalances, err = query.QueryGaiaAllBalances(chainBAPIEndpoint, icaAccount)
				s.Require().NoError(err)
				return icaAccountBalances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range icaAccountBalances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal((int64(TokenAmount)), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)

		address, _ = s.commonHelper.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
		recipientB := address.String()

		s.Require().Eventually(
			func() bool {
				recipientBalances, err = query.QueryGaiaAllBalances(chainBAPIEndpoint, recipientB)
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

		amountToICASend := int64(TokenAmount / 3)
		bankSendMsg := banktypes.NewMsgSend(
			sdk.MustAccAddressFromBech32(icaAccount),
			sdk.MustAccAddressFromBech32(recipientB),
			sdk.NewCoins(sdk.NewCoin(ibcStakeDenom, math.NewInt(amountToICASend))))

		s.tx.BuildICASendTransactionFile(common.Cdc, []proto.Message{bankSendMsg}, s.commonHelper.Resources.ChainA.Validators[0].ConfigDir())
		s.tx.SendICATransaction(s.commonHelper.Resources.ChainA, 0, icaOwnerAccount, common.ConnectionID, configFile(common.ICASendTransactionFileName), common.StandardFees.String())
		s.Require().True(s.commonHelper.HermesClearPacket(common.HermesConfigWithGasPrices, s.commonHelper.Resources.ChainA.Id, icaOwnerPortID, common.IcaChannel))

		s.Require().Eventually(
			func() bool {
				recipientBalances, err = query.QueryGaiaAllBalances(chainBAPIEndpoint, recipientB)
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
