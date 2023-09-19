package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testBankTokenTransfer() {
	s.Run("send_photon_between_accounts", func() {
		var (
			err           error
			c             = s.chainA
			chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
		)

		// define one sender and two recipient accounts
		alice, _ := c.genesisAccounts[1].keyInfo.GetAddress()
		bob, _ := c.genesisAccounts[2].keyInfo.GetAddress()
		charlie, _ := c.genesisAccounts[3].keyInfo.GetAddress()

		var beforeAliceUAtomBalance,
			beforeBobUAtomBalance,
			beforeCharlieUAtomBalance,
			afterAliceUAtomBalance,
			afterBobUAtomBalance,
			afterCharlieUAtomBalance sdk.Coin

		s.Require().Eventually(
			func() bool {
				beforeAliceUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				beforeBobUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				beforeCharlieUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				return beforeAliceUAtomBalance.IsValid() && beforeBobUAtomBalance.IsValid() && beforeCharlieUAtomBalance.IsValid()
			},
			10*time.Second,
			5*time.Second,
		)

		// alice sends tokens to bob
		s.execBankSend(s.chainA, 0, alice.String(), bob.String(), tokenAmount.String(), standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterAliceUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				afterBobUAtomBalance, err = getSpecificBalance(chainEndpoint, bob.String(), uatomDenom)
				s.Require().NoError(err)

				decremented := beforeAliceUAtomBalance.Sub(tokenAmount).Sub(standardFees).IsEqual(afterAliceUAtomBalance)
				incremented := beforeBobUAtomBalance.Add(tokenAmount).IsEqual(afterBobUAtomBalance)

				return decremented && incremented
			},
			time.Minute,
			5*time.Second,
		)

		// save the updated account balances of alice and bob
		beforeAliceUAtomBalance, beforeBobUAtomBalance = afterAliceUAtomBalance, afterBobUAtomBalance

		// alice sends tokens to bob and charlie, at once
		s.execBankMultiSend(s.chainA, 0, alice.String(), []string{bob.String(), charlie.String()}, tokenAmount.String(), standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterAliceUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				afterBobUAtomBalance, err = getSpecificBalance(chainEndpoint, bob.String(), uatomDenom)
				s.Require().NoError(err)

				afterCharlieUAtomBalance, err = getSpecificBalance(chainEndpoint, charlie.String(), uatomDenom)
				s.Require().NoError(err)

				// assert alice's account gets decremented the amount of tokens twice
				decremented := beforeAliceUAtomBalance.Sub(tokenAmount).Sub(tokenAmount).Sub(standardFees).IsEqual(afterAliceUAtomBalance)
				incremented := beforeBobUAtomBalance.Add(tokenAmount).IsEqual(afterBobUAtomBalance) &&
					beforeCharlieUAtomBalance.Add(tokenAmount).IsEqual(afterCharlieUAtomBalance)

				return decremented && incremented
			},
			time.Minute,
			5*time.Second,
		)
	})
}
