package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// func (s *IntegrationTestSuite) TestIBCTokenTransfer() {
// 	var ibcStakeDenom string

// 	s.Run("send_photon_to_chainB", func() {

// 		// require the recipient account receives the IBC tokens (IBC packets ACKd)
// 		var (
// 			balances sdk.Coins
// 			err      error
// 		)

// 		address, err := s.chainB.validators[0].keyInfo.GetAddress()
// 		s.Require().NoError(err)
// 		recipient := address.String()
// 		token := sdk.NewInt64Coin(photonDenom, 3300000000) // 3,300photon
// 		s.sendIBC(s.chainA.id, s.chainB.id, recipient, token)

// 		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

// 		s.Require().Eventually(
// 			func() bool {
// 				balances, err = queryGaiaAllBalances(chainBAPIEndpoint, recipient)
// 				s.Require().NoError(err)

// 				return balances.Len() == 3
// 			},
// 			time.Minute,
// 			5*time.Second,
// 		)

// 		for _, c := range balances {
// 			if strings.Contains(c.Denom, "ibc/") {
// 				ibcStakeDenom = c.Denom
// 				s.Require().Equal(token.Amount.Int64(), c.Amount.Int64())
// 				break
// 			}
// 		}

// 		s.Require().NotEmpty(ibcStakeDenom)
// 	})
// }

func (s *IntegrationTestSuite) TestBankTokenTransfer() {
	s.Run("send_photon_between_accounts", func() {

		var (
			err error
		)

		senderAddress, err := s.chainA.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		recipientAddress, err := s.chainA.validators[1].keyInfo.GetAddress()
		s.Require().NoError(err)
		recipient := recipientAddress.String()

		token := sdk.NewInt64Coin(photonDenom, 3300000000) // 3,300photon
		fees := sdk.NewInt64Coin(photonDenom, 330000)      // 0.33photon

		chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

		beforeSenderPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, "photon")
		s.Require().NoError(err)

		beforeRecipientPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, "photon")
		s.Require().NoError(err)

		s.sendMsgSend(s.chainA, 0, sender, recipient, token.String(), fees.String())

		s.Require().Eventually(
			func() bool {
				afterSenderPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, sender, "photon")
				s.Require().NoError(err)

				afterRecipientPhotonBalance, err := getSpecificBalance(chainAAPIEndpoint, recipient, "photon")
				s.Require().NoError(err)

				decremented := beforeSenderPhotonBalance.Sub(token).Sub(fees).IsEqual(afterSenderPhotonBalance)
				incremented := beforeRecipientPhotonBalance.Add(token).IsEqual(afterRecipientPhotonBalance)

				return decremented && incremented
			},
			time.Minute,
			5*time.Second,
		)
	})
}
