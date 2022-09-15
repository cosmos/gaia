package e2e

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testDelayedVestingAccount(api, home string) {
	validatorA := s.chainA.validators[0]
	sender, err := validatorA.keyInfo.GetAddress()
	s.NoError(err)

	var (
		valOpAddr         = sdk.ValAddress(sender)
		vestingDelayedAcc = s.chainA.delayedVestingAcc
		transferAmount    = sdk.NewCoin(uatomDenom, sdk.NewInt(80000000))
		delegationAmount  = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
		delegationFees    = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
	)
	s.Run("test continuous vesting genesis account", func() {
		acc, err := queryDelayedVestingAccount(api, vestingDelayedAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, vestingDelayedAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(), vestingDelayedAcc.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr.String(), vestingDelayedAcc.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(delegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should fail
		s.sendMsgSend(
			s.chainA,
			0,
			vestingDelayedAcc.String(),
			Address(),
			transferAmount.String(),
			fees.String(),
			true,
		)

		waitTime := time.Duration(time.Now().Unix() - acc.EndTime)
		time.Sleep(waitTime * time.Second)

		//	Transfer coins should succeed
		s.sendMsgSend(s.chainA, 0, vestingDelayedAcc.String(), Address(), transferAmount.String(), fees.String(), false)
	})
}

func (s *IntegrationTestSuite) testContinuousVestingAccount(api, home string) {
	s.Run("test continuous vesting genesis account", func() {
		validatorA := s.chainA.validators[0]
		sender, err := validatorA.keyInfo.GetAddress()
		s.NoError(err)

		var (
			valOpAddr            = sdk.ValAddress(sender)
			continuousVestingAcc = s.chainA.continuousVestingAcc
			transferAmount       = sdk.NewCoin(uatomDenom, sdk.NewInt(80000000))
			delegationAmount     = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
			delegationFees       = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
		)

		acc, err := queryContinuousVestingAccount(api, continuousVestingAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, continuousVestingAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should fail
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(), continuousVestingAcc.String(), home, delegationFees.String())

		// Validate delegation fail
		s.Require().Eventually(
			func() bool {
				_, err := queryDelegation(api, valOpAddr.String(), continuousVestingAcc.String())
				return err != nil
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should fail
		s.sendMsgSend(
			s.chainA,
			0,
			continuousVestingAcc.String(),
			Address(),
			transferAmount.String(),
			fees.String(),
			true,
		)

		waitStartTime := time.Duration(time.Now().Unix() - acc.StartTime)
		time.Sleep(waitStartTime * time.Second)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(), continuousVestingAcc.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr.String(), continuousVestingAcc.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(delegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should fail
		s.sendMsgSend(
			s.chainA,
			0,
			continuousVestingAcc.String(),
			Address(),
			transferAmount.String(),
			fees.String(),
			true,
		)

		waitEndTime := time.Duration(time.Now().Unix() - acc.GetStartTime())
		time.Sleep(waitEndTime * time.Second)

		//	Transfer coins should succeed
		s.sendMsgSend(s.chainA, 0, continuousVestingAcc.String(), Address(), transferAmount.String(), fees.String(), false)
	})
}

func (s *IntegrationTestSuite) testPermanentLockedAccount(api, home string) {
	s.Run("test permanent locked vesting genesis account", func() {
		validatorA := s.chainA.validators[0]
		sender, err := validatorA.keyInfo.GetAddress()
		s.NoError(err)

		var (
			valOpAddr        = sdk.ValAddress(sender)
			transferAmount   = sdk.NewCoin(uatomDenom, sdk.NewInt(80000000))
			delegationAmount = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
			delegationFees   = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
			val0ConfigDir    = s.chainA.validators[0].configDir()
		)
		kb, err := keyring.New(keyringAppName, keyring.BackendTest, val0ConfigDir, nil, cdc)
		s.Require().NoError(err)

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
		s.Require().NoError(err)

		mnemoinic, err := createMnemonic()
		s.Require().NoError(err)

		// Use the first wallet from the same mnemonic by HD path
		account, err := kb.NewAccount("continuous_vesting", mnemoinic, "", HDPathZero, algo)
		s.Require().NoError(err)
		permanentLockedAddr, err := account.GetAddress()
		s.Require().NoError(err)

		s.execCreatePeriodicVestingAccount(s.chainA, home, permanentLockedAddr.String(), periodJSONFile,
			withKeyValue("from", sender.String()),
		)

		_, err = queryPermanentLockedAccount(api, permanentLockedAddr.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, permanentLockedAddr.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingAmountVested.Amount, balance.Amount)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(),
			permanentLockedAddr.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr.String(), permanentLockedAddr.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(delegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should fail
		s.sendMsgSend(s.chainA, 0, permanentLockedAddr.String(), Address(),
			transferAmount.String(), fees.String(), true)
	})
}

func (s *IntegrationTestSuite) testPeriodicVestingAccount(api, home string) {
	s.Run("test periodic vesting genesis account", func() {
		validatorA := s.chainA.validators[0]
		sender, err := validatorA.keyInfo.GetAddress()
		s.NoError(err)

		var (
			delegationFees   = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
			valOpAddr        = sdk.ValAddress(sender)
			transferAmount   = sdk.NewCoin(uatomDenom, sdk.NewInt(80000000))
			delegationAmount = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
			val0ConfigDir    = s.chainA.validators[0].configDir()
		)
		kb, err := keyring.New(keyringAppName, keyring.BackendTest, val0ConfigDir, nil, cdc)
		s.Require().NoError(err)

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
		s.Require().NoError(err)

		mnemoinic, err := createMnemonic()
		s.Require().NoError(err)

		// Use the first wallet from the same mnemonic by HD path
		account, err := kb.NewAccount("continuous_vesting", mnemoinic, "", HDPathZero, algo)
		s.Require().NoError(err)
		periodicVestingAddr, err := account.GetAddress()
		s.Require().NoError(err)

		s.execCreatePeriodicVestingAccount(s.chainA, home, periodicVestingAddr.String(), periodJSONFile)

		acc, err := queryPeriodicVestingAccount(api, periodicVestingAddr.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, periodicVestingAddr.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should fail
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(),
			periodicVestingAddr.String(), home, delegationFees.String())

		// Validate delegation fail
		s.Require().Eventually(
			func() bool {
				_, err := queryDelegation(api, valOpAddr.String(), periodicVestingAddr.String())
				return err != nil
			},
			20*time.Second,
			5*time.Second,
		)

		waitFirstPeriod := time.Duration(time.Now().Unix() - (acc.StartTime + acc.VestingPeriods[0].Length))
		time.Sleep(waitFirstPeriod * time.Second)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(),
			periodicVestingAddr.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr.String(), periodicVestingAddr.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(delegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should succeed
		s.sendMsgSend(
			s.chainA,
			0,
			periodicVestingAddr.String(),
			Address(),
			transferAmount.String(),
			fees.String(),
			false,
		)

		waitSecondPeriod := time.Duration(time.Now().Unix() - (acc.StartTime + acc.VestingPeriods[1].Length))
		time.Sleep(waitSecondPeriod * time.Second)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr.String(),
			periodicVestingAddr.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr.String(), periodicVestingAddr.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(delegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should succeed
		s.sendMsgSend(
			s.chainA,
			0,
			periodicVestingAddr.String(),
			Address(),
			transferAmount.String(),
			fees.String(),
			false,
		)
	})
}
