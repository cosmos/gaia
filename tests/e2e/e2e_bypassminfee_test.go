package e2e

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *IntegrationTestSuite) testBypassMinFeeWithdrawReward() {
	// s.T().Skip()
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	s.Require().Eventually(
		func() bool {
			globalFees, err := queryGlobalFees(chainAAPIEndpoint)
			s.T().Logf("After gov new global fee proposal: %s", globalFees.String())
			s.Require().NoError(err)
			return true

			// attention: if global fee is empty, when query globalfee, it shows empty rather than default ante.DefaultZeroGlobalFee() = 0uatom.
		},
		15*time.Second,
		5*time.Second,
	)

	// gov propose withdraw to be bypass-msg first
	submitterAddr := s.chainA.validators[0].keyInfo.GetAddress()
	submitter := submitterAddr.String()
	s.govProposeNewBypassMsgs([]string{sdk.MsgTypeURL(&distributiontypes.MsgWithdrawDelegatorReward{})}, proposalCounter, submitter, standardFees.String())

	// GlobalFee == minGasPrice+uatomDenom
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
}
