package e2e

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testByPassMinFeeWithdrawReward(endpoint string) {
	paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
	payee := s.chainA.validators[0].keyInfo.GetAddress()

	testCases := []struct {
		name   string
		fee    string
		expErr bool
	}{
		{
			"bypass-msg with fee in the denom of global fee, pass",
			paidFeeAmt + uatomDenom,
			false,
		},
		{
			"bypass-msg with zero coin in the denom of global fee, pass",
			"0" + uatomDenom,
			false,
		},
		{
			"bypass-msg with zero coin not in the denom of global fee, pass",
			"0" + photonDenom,
			false,
		},
		{
			"bypass-msg with non-zero coin not in the denom of global fee, fail",
			paidFeeAmt + photonDenom,
			true,
		},
	}

	for _, tc := range testCases {

		// get delegator rewards
		rewards, err := queryDelegatorTotalRewards(endpoint, payee.String())
		s.Require().NoError(err)

		// get delegator stake balance
		stakeBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
		s.Require().NoError(err)
		fmt.Println(stakeBalance)

		// withdraw rewards
		s.Run(tc.name, func() {
			s.execWithdrawAllRewards(s.chainA, 0, payee.String(), tc.fee, tc.expErr)
		})

		// get updated balance
		newStakeBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
		s.Require().NoError(err)

		// check that the update balance was increased by at least the amount of stake tokens in the reward
		total := rewards.GetTotal().Add(sdk.NewDecCoinFromCoin(stakeBalance)).Sort()
		fmt.Println("total", total.String())
		fmt.Println("rewards", rewards.String())
		s.Require().Equal(total[0].Denom, stakeDenom)

		fmt.Println(sdk.NewDecCoinFromCoin(newStakeBalance).String(), total[0].String())
		s.Require().True(sdk.NewDecCoinFromCoin(newStakeBalance).IsGTE(total[0]))
	}
}
