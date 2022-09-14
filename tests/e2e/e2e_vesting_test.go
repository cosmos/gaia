package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testDelayedVestingAccount(api, home string) {
	validatorA := s.chainA.validators[0]
	sender, err := validatorA.keyInfo.GetAddress()
	s.NoError(err)

	var (
		delegationFees    = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
		valOpAddr         = sdk.ValAddress(sender)
		vestingDelayedAcc = s.chainA.delayedVestingAcc
		transferAmount    = sdk.NewCoin(uatomDenom, sdk.NewInt(1550000))
		delegationAmount  = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
	)
	s.Run("test delayed vesting genesis account", func() {
		//	Balance should be zero
		afterAtomBalance, err := getSpecificBalance(api, vestingDelayedAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingAmount, afterAtomBalance)

		//	Transfer coins should fail
		ad := Address()
		fmt.Printf("transfer form %s to %s\n", vestingDelayedAcc.String(), ad)
		s.sendMsgSend(
			s.chainA,
			0,
			vestingDelayedAcc.String(),
			ad,
			transferAmount.String(),
			fees.String(),
			true,
		)
		//	Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(), vestingDelayedAcc.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr.String(), vestingDelayedAcc.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(vestingAmount.Amount.Sub(delegationAmount.Amount)))
			},
			20*time.Second,
			5*time.Second,
		)

		time.Sleep(vestingLength)

		//	Balance should be equal to original vesting coins
		s.verifyBalanceChange(api, transferAmount, vestingDelayedAcc.String())
		//	Transfer coins should fail
		s.sendMsgSend(s.chainA, 0, vestingDelayedAcc.String(), Address(), transferAmount.String(), fees.String(), false)
	})

	//s.Run("test delayed vesting created by API", func() {
	//	newVestingAddr := Address()
	//	s.execCreateVestingAccount(s.chainA, newVestingAddr, vestingAmount.String(), vestingEndTime)
	//
	//	//	Balance should be equal to original vesting coins
	//	s.verifyBalanceChange(api, amount, vestingDelayedAcc.String())
	//	//	Transfer coins should fail
	//	s.sendMsgSend(s.chainA, 0, newVestingAddr, Address(), amount.String(), fees.String(), true)
	//	//	Delegate coins should succeed
	//	s.executeDelegate(s.chainA, 0, api, amount.String(), valOpAddr.String(), newVestingAddr, home, fees.String())
	//
	//	// Validate delegation successful
	//	s.Require().Eventually(
	//		func() bool {
	//			res, err := queryDelegation(api, valOpAddr.String(), newVestingAddr)
	//			amt := res.GetDelegationResponse().GetDelegation().GetShares()
	//			s.Require().NoError(err)
	//
	//			return amt.Equal(sdk.NewDecFromInt(amount.Amount))
	//		},
	//		20*time.Second,
	//		5*time.Second,
	//	)
	//
	//	time.Sleep(vestingLength)
	//
	//	//	Balance should be equal to original vesting coins
	//	s.verifyBalanceChange(api, amount, vestingDelayedAcc.String())
	//	//	Transfer coins should fail
	//	s.sendMsgSend(s.chainA, 0, newVestingAddr, Address(), amount.String(), fees.String(), false)
	//})
}

func (s *IntegrationTestSuite) testContinuousVestingAccount(api, home string) {
	// TODO test genesis account
	vestingContinuousAcc := s.chainA.continuousVestingAcc
	fmt.Println(vestingContinuousAcc.String())

	// Create a continuous vesting account
	//	Balance should be equal to original vesting coins
	//	Transfer coins should fail
	//	Delegate coins should fail
	//	Balance should be equal to original vesting coins - delegated coins

	// Wait until the StartTime reach
	//	Balance should be equal to original vesting coins
	//	Transfer coins should fail
	//	Delegate coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins

	// Wait the (EndTime - StartTime) / 2 reach
	//	Balance should be equal to original vesting coins - delegated coins
	//	Delegate coins should succeed
	//	Transfer all coins should fail
	//	Transfer half of the coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins - sent coins

	// Wait until the EndTime reach
	//	Balance should be equal to original vesting coins - delegated coins - sent coins
	//	Delegate coins should succeed
	//	Transfer all coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins - sent coins

	// Formula:
	// X := T - StartTime
	// Y := EndTime - StartTime
	// V' := OV * (X / Y)
	// V := OV - V'
}

func (s *IntegrationTestSuite) testPermanentLockedAccount(api, home string) {
	newVestingAcc := Address()
	s.execCreatePermanentLockedAccount(s.chainA, newVestingAcc, vestingAmount.String())
	fmt.Println(newVestingAcc)

	// Create a permanently locked account
	//	Balance should be equal to original vesting coins
	//	Transfer coins should fail
	//	Delegate coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins
}

func (s *IntegrationTestSuite) testPeriodicVestingAccount(api, home string) {
	newVestingAcc := Address()
	s.execCreatePeriodicVestingAccount(s.chainA, newVestingAcc, periodJSONFile)
	fmt.Println(newVestingAcc)

	// Create a periodic vesting account
	//	Balance should be equal to original vesting coins
	//	Transfer coins should fail
	//	Delegate coins should fail
	//	Balance should be equal to original vesting coins - delegated coins

	// Wait until the first-period reach
	//	Balance should be equal to original vesting coins
	//	Delegate coins should succeed
	//	Transfer all coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins - sent coins

	// Wait until the next period reach
	//	Balance should be equal to original vesting coins
	//	Delegate coins should succeed
	//	Transfer all coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins - sent coins

	// Wait until the EndTime reach
	//	Balance should be equal to original vesting coins - delegated coins - sent coins
	//	Delegate coins should succeed
	//	Transfer all coins should succeed
	//	Balance should be equal to original vesting coins - delegated coins - sent coins

	// Formula:
	// CT := StartTime
	// Set V' := 0
	// # For each Period P:
	// X := T - CT
	// IF X >= P.Length {
	//	V' += P.Amount
	//	CT += P.Length
	// } ELSE {
	//	break
	// }
	// V := OV - V'
}
