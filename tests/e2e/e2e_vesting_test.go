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
	valOpAddr := sdk.ValAddress(sender)

	s.Run("test delayed vesting genesis account", func() {
		//	Balance should be equal to original vesting coins
		s.verifyBalanceChange(api, vestingAmount, vestingDelayedAcc.String())
	})

	s.Run("test delayed vesting created by API", func() {
		newVestingAcc := Address()
		s.execCreateVestingAccount(s.chainA, newVestingAcc, vestingAmount.String(), vestingEndTime)

		//	Balance should be equal to original vesting coins
		s.verifyBalanceChange(api, vestingAmount, vestingDelayedAcc.String())
		//	Transfer coins should fail
		s.sendMsgSend(s.chainA, 0, newVestingAcc, Address(), vestingAmount.String(), fees.String(), true)
		//	Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, valOpAddr.String(), newVestingAcc, home, fees.String(), vestingAmount, false)

		time.Sleep(10 * time.Second)

		//	Balance should be equal to original vesting coins
		s.verifyBalanceChange(api, vestingAmount, vestingDelayedAcc.String())
		//	Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, valOpAddr.String(), newVestingAcc, home, fees.String(), vestingAmount, false)
		//	Transfer coins should fail
		s.sendMsgSend(s.chainA, 0, newVestingAcc, Address(), vestingAmount.String(), fees.String(), false)
	})
}

func (s *IntegrationTestSuite) testContinuousVestingAccount(api, home string) {
	// TODO test genesis account
	_ = vestingContinuousAcc
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
