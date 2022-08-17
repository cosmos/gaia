package e2e

import (
	"cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

func (s *IntegrationTestSuite) testStaking() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	seedAmount := sdk.NewCoin(uatomDenom, math.NewInt(1000000000)) // 2,200uatom
	delegationAmount := math.NewInt(500000000)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount) // 2,200uatom

	validatorA := s.chainA.validators[0]
	validatorB := s.chainA.validators[1]
	sender, err := validatorA.keyInfo.GetAddress()
	s.NoError(err)
	validatorBAddr, err := validatorB.keyInfo.GetAddress()
	s.NoError(err)

	valOperA := sdk.ValAddress(sender)
	valOperB := sdk.ValAddress(validatorBAddr)

	alice := s.executeGKeysAddCommand(s.chainA, 0, "alice", dataDirectoryHome)
	// up the amount
	delegationFees := sdk.NewCoin(uatomDenom, math.NewInt(10))

	// Fund Alice
	s.sendMsgSend(s.chainA, 0, sender.String(), alice, seedAmount.String(), fees.String(), false)
	s.verifyBalanceChange(chainAAPIEndpoint, seedAmount, alice)

	// Alice delegate uatom to Validator A
	s.executeDelegate(s.chainA, 0, chainAAPIEndpoint, delegation.String(), valOperA.String(), alice, dataDirectoryHome, delegationFees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainAAPIEndpoint, valOperA.String(), alice)
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Alice redelegate uatom from Validator A to Validator B
	s.executeRedelegate(s.chainA, 0, chainAAPIEndpoint, delegation.String(), valOperA.String(), valOperB.String(), alice, dataDirectoryHome, delegationFees.String())

	// Validate redelegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainAAPIEndpoint, valOperB.String(), alice)
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)
}
