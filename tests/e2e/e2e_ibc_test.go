package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ForwardMetadata struct {
	Receiver string `json:"receiver"`
	Port     string `json:"port"`
	Channel  string `json:"channel"`
	// Timeout        time.Duration `json:"timeout"`
	// Retries        *uint8        `json:"retries,omitempty"`
	// Next           *string       `json:"next,omitempty"`
	// RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

//nolint:unparam

func (s *IntegrationTestSuite) testIBCTokenTransfer() {
	s.Run("send_uatom_to_chainB", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances      sdk.Coins
			err           error
			beforeBalance int64
			ibcStakeDenom string
		)

		address, _ := s.commonHelper.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
		sender := address.String()

		address, _ = s.commonHelper.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
		recipient := address.String()

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainB.Id][0].GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				balances, err = query.QueryGaiaAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				beforeBalance = c.Amount.Int64()
				break
			}
		}

		tokenAmt := 3300000000
		s.tx.SendIBC(s.commonHelper.Resources.ChainA, 0, sender, recipient, strconv.Itoa(tokenAmt)+common.UatomDenom, common.StandardFees.String(), "", common.TransferChannel, nil, false)

		pass := s.commonHelper.HermesClearPacket(common.HermesConfigWithGasPrices, s.commonHelper.Resources.ChainA.Id, common.TransferPort, common.TransferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				balances, err = query.QueryGaiaAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal((int64(tokenAmt) + beforeBalance), c.Amount.Int64())
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

/*
TestMultihopIBCTokenTransfer tests that sending an IBC transfer using the IBC Packet Forward Middleware accepts a port, channel and account address

Steps:
1. Check balance of Account 1 on Chain 1
2. Check balance of Account 2 on Chain 1
3. Account 1 on Chain 1 sends x tokens to Account 2 on Chain 1 via Account 1 on Chain 2
4. Check Balance of Account 1 on Chain 1, confirm it is original minus x tokens
5. Check Balance of Account 2 on Chain 1, confirm it is original plus x tokens

*/
// TODO: Add back only if packet forward middleware has a working version compatible with IBC v3.0.x
func (s *IntegrationTestSuite) testMultihopIBCTokenTransfer() {
	time.Sleep(30 * time.Second)

	s.Run("send_successful_multihop_uatom_to_chainA_from_chainA", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			err error
		)

		address, _ := s.commonHelper.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
		sender := address.String()

		address, _ = s.commonHelper.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
		middlehop := address.String()

		address, _ = s.commonHelper.Resources.ChainA.Validators[1].KeyInfo.GetAddress()
		recipient := address.String()

		forwardPort := "transfer"
		forwardChannel := "channel-0"

		tokenAmt := 3300000000

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainA.Id][0].GetHostPort("1317/tcp"))

		var (
			beforeSenderUAtomBalance    sdk.Coin
			beforeRecipientUAtomBalance sdk.Coin
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderUAtomBalance, err = query.GetSpecificBalance(chainAAPIEndpoint, sender, common.UatomDenom)
				s.Require().NoError(err)

				beforeRecipientUAtomBalance, err = query.GetSpecificBalance(chainAAPIEndpoint, recipient, common.UatomDenom)
				s.Require().NoError(err)

				return beforeSenderUAtomBalance.IsValid() && beforeRecipientUAtomBalance.IsValid()
			},
			1*time.Minute,
			5*time.Second,
		)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: recipient,
				Channel:  forwardChannel,
				Port:     forwardPort,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		s.Require().NoError(err)

		s.tx.SendIBC(s.commonHelper.Resources.ChainA, 0, sender, middlehop, strconv.Itoa(tokenAmt)+common.UatomDenom, common.StandardFees.String(), string(memo), common.TransferChannel, nil, false)

		pass := s.commonHelper.HermesClearPacket(common.HermesConfigWithGasPrices, s.commonHelper.Resources.ChainA.Id, common.TransferPort, common.TransferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				afterSenderUAtomBalance, err := query.GetSpecificBalance(chainAAPIEndpoint, sender, common.UatomDenom)
				s.Require().NoError(err)

				afterRecipientUAtomBalance, err := query.GetSpecificBalance(chainAAPIEndpoint, recipient, common.UatomDenom)
				s.Require().NoError(err)

				decremented := beforeSenderUAtomBalance.Sub(common.TokenAmount).Sub(common.StandardFees).IsEqual(afterSenderUAtomBalance)
				incremented := beforeRecipientUAtomBalance.Add(common.TokenAmount).IsEqual(afterRecipientUAtomBalance)

				return decremented && incremented
			},
			1*time.Minute,
			5*time.Second,
		)
	})
}

/*
TestFailedMultihopIBCTokenTransfer tests that sending a failing IBC transfer using the IBC Packet Forward
Middleware will send the tokens back to the original account after failing.
*/
func (s *IntegrationTestSuite) testFailedMultihopIBCTokenTransfer() {
	time.Sleep(30 * time.Second)

	s.Run("send_failed_multihop_uatom_to_chainA_from_chainA", func() {
		address, _ := s.commonHelper.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
		sender := address.String()

		address, _ = s.commonHelper.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
		middlehop := address.String()

		address, _ = s.commonHelper.Resources.ChainA.Validators[1].KeyInfo.GetAddress()
		recipient := strings.Replace(address.String(), "cosmos", "foobar", 1) // this should be an invalid recipient to force the tx to fail

		forwardPort := "transfer"
		forwardChannel := "channel-0"

		tokenAmt := 3300000000

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainA.Id][0].GetHostPort("1317/tcp"))

		var (
			beforeSenderUAtomBalance sdk.Coin
			err                      error
		)

		s.Require().Eventually(
			func() bool {
				beforeSenderUAtomBalance, err = query.GetSpecificBalance(chainAAPIEndpoint, sender, common.UatomDenom)
				s.Require().NoError(err)

				return beforeSenderUAtomBalance.IsValid()
			},
			1*time.Minute,
			5*time.Second,
		)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: recipient,
				Channel:  forwardChannel,
				Port:     forwardPort,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		s.Require().NoError(err)

		s.tx.SendIBC(s.commonHelper.Resources.ChainA, 0, sender, middlehop, strconv.Itoa(tokenAmt)+common.UatomDenom, common.StandardFees.String(), string(memo), common.TransferChannel, nil, false)

		// Sender account should be initially decremented the full amount
		s.Require().Eventually(
			func() bool {
				afterSenderUAtomBalance, err := query.GetSpecificBalance(chainAAPIEndpoint, sender, common.UatomDenom)
				s.Require().NoError(err)

				returned := beforeSenderUAtomBalance.Sub(common.TokenAmount).Sub(common.StandardFees).IsEqual(afterSenderUAtomBalance)

				return returned
			},
			1*time.Minute,
			5*time.Second,
		)

		// since the forward receiving account is invalid, it should be refunded to the original sender (minus the original fee)
		s.Require().Eventually(
			func() bool {
				pass := s.commonHelper.HermesClearPacket(common.HermesConfigWithGasPrices, s.commonHelper.Resources.ChainA.Id, common.TransferPort, common.TransferChannel)
				s.Require().True(pass)

				afterSenderUAtomBalance, err := query.GetSpecificBalance(chainAAPIEndpoint, sender, common.UatomDenom)
				s.Require().NoError(err)
				returned := beforeSenderUAtomBalance.Sub(common.StandardFees).IsEqual(afterSenderUAtomBalance)
				return returned
			},
			5*time.Minute,
			10*time.Second,
		)
	})
}
