package e2e

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *IntegrationTestSuite) testBypassMinFeeWithdrawReward(endpoint string) {
	// submit gov prop to change bypass-msg param to MsgWithdrawDelegatorReward
	submitterAddr := s.chainA.validators[0].keyInfo.GetAddress()
	submitter := submitterAddr.String()
	proposalCounter++
	s.govProposeNewBypassMsgs([]string{sdk.MsgTypeURL(&distributiontypes.MsgWithdrawDelegatorReward{})}, proposalCounter, submitter, standardFees.String())

	paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
	payee := s.chainA.validators[0].keyInfo.GetAddress()

	testCases := []struct {
		name                    string
		fee                     string
		changeMaxBypassGasUsage bool
		expErr                  bool
	}{
		{
			"bypass-msg with fee in the denom of global fee, pass",
			paidFeeAmt + uatomDenom,
			false,
			false,
		},
		{
			"bypass-msg with zero coin in the denom of global fee, pass",
			"0" + uatomDenom,
			false,
			false,
		},
		{
			"bypass-msg with zero coin not in the denom of global fee, pass",
			"0" + photonDenom,
			false,
			false,
		},
		{
			"bypass-msg with non-zero coin not in the denom of global fee, fail",
			paidFeeAmt + photonDenom,
			false,
			true,
		},
		{
			"bypass-msg with zero coin in the denom of global fee and maxTotalBypassMinFeeMsgGasUsage set to 1, fail",
			"0" + uatomDenom,
			true,
			true,
		},
		{
			"bypass-msg with non zero coin in the denom of global fee and maxTotalBypassMinFeeMsgGasUsage set to 1, pass",
			paidFeeAmt + uatomDenom,
			false,
			false,
		},
	}

	for _, tc := range testCases {

		if tc.changeMaxBypassGasUsage {
			proposalCounter++
			// change MaxTotalBypassMinFeeMsgGasUsage through governance proposal from 1_000_0000 to 1
			s.govProposeNewMaxTotalBypassMinFeeMsgGasUsage(1, proposalCounter, submitter)
		}

		// get delegator rewards
		rewards, err := queryDelegatorTotalRewards(endpoint, payee.String())
		s.Require().NoError(err)

		// get delegator stake balance
		oldBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
		s.Require().NoError(err)

		// withdraw rewards
		s.Run(tc.name, func() {
			s.execWithdrawAllRewards(s.chainA, 0, payee.String(), tc.fee, tc.expErr)
		})

		if !tc.expErr {
			// get updated balance
			incrBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
			s.Require().NoError(err)

			// compute sum of old balance and stake token rewards
			oldBalancePlusReward := rewards.GetTotal().Add(sdk.NewDecCoinFromCoin(oldBalance))
			s.Require().Equal(oldBalancePlusReward[0].Denom, stakeDenom)

			// check updated balance got increased by at least oldBalancePlusReward
			s.Require().True(sdk.NewDecCoinFromCoin(incrBalance).IsGTE(oldBalancePlusReward[0]))
		}
	}
}
