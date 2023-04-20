package e2e

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *IntegrationTestSuite) testBypassMinFeeWithdrawReward() {

	// submit gov prop to change bypass-msg param to MsgWithdrawDelegatorReward
	submitterAddr := s.chainA.validators[0].keyInfo.GetAddress()
	submitter := submitterAddr.String()
	proposalCounter++
	s.govProposeNewBypassMsgs([]string{sdk.MsgTypeURL(&distributiontypes.MsgWithdrawDelegatorReward{})}, proposalCounter, submitter, standardFees.String())

	// Note that the global fee min gas prices is equal to minGasPrice+uatomDenom
	paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
	payee := s.chainA.validators[0].keyInfo.GetAddress()

	// pass
	s.T().Logf("bypass-msg with fee in the denom of global fee, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), paidFeeAmt+uatomDenom, false)
	// pass
	s.T().Logf("bypass-msg with zero coin in the denom of global fee, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), "0"+uatomDenom, false)
	// pass
	s.T().Logf("bypass-msg with zero coin not in the denom of global fee, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), "0"+photonDenom, false)
	// fail
	s.T().Logf("bypass-msg with non-zero coin not in the denom of global fee, fail")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), paidFeeAmt+photonDenom, true)

	proposalCounter++
	// change MaxTotalBypassMinFeeMsgGasUsage through governance proposal from 1_000_0000 to 1
	s.govProposeNewMaxTotalBypassMinFeeMsgGasUsage(1, proposalCounter, submitter)

	// fail
	s.T().Logf("bypass-msg has zero coin and maxTotalBypassMinFeeMsgGasUsage set to 1, fail")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), "0"+uatomDenom, true)

	// pass
	s.T().Logf("bypass-msg has non zero coin and maxTotalBypassMinFeeMsgGasUsage set to 1, pass")
	s.execWithdrawAllRewards(s.chainA, 0, payee.String(), paidFeeAmt+uatomDenom, false)
}
