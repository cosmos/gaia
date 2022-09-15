package e2e

import (
	"cosmossdk.io/math"
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	vestingPeriodFilePath = "test_period.json"
)

type (
	vestingPeriod struct {
		StartTime int64    `json:"start_time"`
		Periods   []period `json:"periods"`
	}
	period struct {
		Coins  string `json:"coins"`
		Length int64  `json:"length_seconds"`
	}
)

var (
	vestingAmountVested     = sdk.NewCoin(uatomDenom, math.NewInt(99900000000))
	vestingAmount           = sdk.NewCoin(uatomDenom, math.NewInt(350000))
	vestingBalance          = sdk.NewCoins(vestingAmountVested).Add(vestingAmount)
	vestingTransferAmount   = sdk.NewCoin(uatomDenom, sdk.NewInt(800000000))
	vestingDelegationAmount = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
	vestingDelegationFees   = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
)

func (s *IntegrationTestSuite) testDelayedVestingAccount(api, home string) {
	validatorA := s.chainA.validators[0]
	sender, err := validatorA.keyInfo.GetAddress()
	s.NoError(err)

	var (
		valOpAddr         = sdk.ValAddress(sender).String()
		vestingDelayedAcc = s.chainA.delayedVestingAcc
		delegationAmount  = sdk.NewCoin(uatomDenom, sdk.NewInt(500000000))
		delegationFees    = sdk.NewCoin(uatomDenom, sdk.NewInt(10))
	)
	s.Run("test delayed vesting genesis account", func() {
		s.T().Parallel()

		acc, err := queryDelayedVestingAccount(api, vestingDelayedAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, vestingDelayedAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, delegationAmount.String(), valOpAddr, vestingDelayedAcc.String(), home, delegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, vestingDelayedAcc.String())
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
			vestingTransferAmount.String(),
			fees.String(),
			true,
		)

		waitTime := acc.EndTime - time.Now().Unix() + 5
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Second)
		}

		//	Transfer coins should succeed
		s.sendMsgSend(s.chainA, 0, vestingDelayedAcc.String(), Address(), vestingTransferAmount.String(), fees.String(), false)
	})
}

func (s *IntegrationTestSuite) testContinuousVestingAccount(api, home string) {
	s.Run("test continuous vesting genesis account", func() {
		s.T().Parallel()

		validatorA := s.chainA.validators[1]
		sender, err := validatorA.keyInfo.GetAddress()
		s.NoError(err)

		var (
			valOpAddr            = sdk.ValAddress(sender).String()
			continuousVestingAcc = s.chainA.continuousVestingAcc
		)

		acc, err := queryContinuousVestingAccount(api, continuousVestingAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, continuousVestingAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, vestingDelegationAmount.String(),
			valOpAddr, continuousVestingAcc.String(), home, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, continuousVestingAcc.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(vestingDelegationAmount.Amount))
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
			vestingTransferAmount.String(),
			fees.String(),
			true,
		)

		waitStartTime := acc.StartTime - time.Now().Unix()
		if waitStartTime > 0 {
			time.Sleep(time.Duration(waitStartTime) * time.Second)
		}

		//	Transfer coins should fail
		s.sendMsgSend(
			s.chainA,
			0,
			continuousVestingAcc.String(),
			Address(),
			vestingTransferAmount.String(),
			fees.String(),
			true,
		)

		waitEndTime := acc.EndTime - time.Now().Unix()
		if waitEndTime > 0 {
			time.Sleep(time.Duration(waitEndTime) * time.Second)
		}

		//	Transfer coins should succeed
		s.sendMsgSend(s.chainA, 0, continuousVestingAcc.String(), Address(), vestingTransferAmount.String(), fees.String(), false)
	})
}

func (s *IntegrationTestSuite) testPermanentLockedAccount(api, home string) {
	s.Run("test permanent locked vesting genesis account", func() {
		s.T().Parallel()

		val := s.chainA.validators[0]
		sender, err := val.keyInfo.GetAddress()
		s.NoError(err)

		valOpAddr := sdk.ValAddress(sender).String()
		kb, err := keyring.New(keyringAppName, keyring.BackendTest, val.configDir(), nil, cdc)
		s.Require().NoError(err)

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
		s.Require().NoError(err)

		mnemoinic, err := createMnemonic()
		s.Require().NoError(err)

		// Use the first wallet from the same mnemonic by HD path
		account, err := kb.NewAccount("permanent_locked_vesting", mnemoinic, "", HDPathZero, algo)
		s.Require().NoError(err)
		permanentLockedAddr, err := account.GetAddress()
		s.Require().NoError(err)

		s.execCreatePermanentLockedAccount(s.chainA, home, permanentLockedAddr.String(),
			vestingAmountVested.String(), withKeyValue("from", sender.String()),
		)

		_, err = queryPermanentLockedAccount(api, permanentLockedAddr.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, permanentLockedAddr.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingAmountVested.Amount, balance.Amount)

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, vestingDelegationAmount.String(), valOpAddr,
			permanentLockedAddr.String(), home, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, permanentLockedAddr.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(vestingDelegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should fail
		s.sendMsgSend(s.chainA, 0, permanentLockedAddr.String(), Address(),
			vestingTransferAmount.String(), fees.String(), true)
	})
}

func (s *IntegrationTestSuite) testPeriodicVestingAccount(api, home string) {
	s.Run("test periodic vesting genesis account", func() {
		s.T().Parallel()

		val := s.chainB.validators[1]
		sender, err := val.keyInfo.GetAddress()
		s.NoError(err)

		valOpAddr := sdk.ValAddress(sender).String()
		kb, err := keyring.New(keyringAppName, keyring.BackendTest, val.configDir(), nil, cdc)
		s.Require().NoError(err)

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
		s.Require().NoError(err)

		mnemoinic, err := createMnemonic()
		s.Require().NoError(err)

		// Use the first wallet from the same mnemonic by HD path
		account, err := kb.NewAccount("periodic_vesting", mnemoinic, "", HDPathZero, algo)
		s.Require().NoError(err)
		periodicVestingAddr, err := account.GetAddress()
		s.Require().NoError(err)

		s.execCreatePeriodicVestingAccount(
			s.chainA,
			home,
			periodicVestingAddr.String(),
			withKeyValue("from", sender.String()),
		)

		acc, err := queryPeriodicVestingAccount(api, periodicVestingAddr.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, periodicVestingAddr.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(sdk.NewCoin(uatomDenom, sdk.NewInt(1700000000)), balance)

		waitStartTime := acc.StartTime - time.Now().Unix()
		if waitStartTime > 0 {
			time.Sleep(time.Duration(waitStartTime) * time.Second)
		}

		//	Transfer coins should fail
		s.sendMsgSend(
			s.chainA,
			0,
			periodicVestingAddr.String(),
			Address(),
			vestingTransferAmount.String(),
			fees.String(),
			true,
		)

		waitFirstPeriod := (acc.StartTime + acc.VestingPeriods[0].Length) - time.Now().Unix()
		if waitFirstPeriod > 0 {
			time.Sleep(time.Duration(waitFirstPeriod) * time.Second)
		}

		// Delegate coins should succeed
		s.executeDelegate(s.chainA, 0, api, vestingDelegationAmount.String(), valOpAddr,
			periodicVestingAddr.String(), home, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, periodicVestingAddr.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(sdk.NewDecFromInt(vestingDelegationAmount.Amount))
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
			vestingTransferAmount.String(),
			fees.String(),
			false,
		)

		waitSecondPeriod := acc.StartTime + acc.VestingPeriods[0].Length + acc.VestingPeriods[1].Length
		waitSecondPeriod -= time.Now().Unix()
		if waitSecondPeriod > 0 {
			time.Sleep(time.Duration(waitSecondPeriod) * time.Second)
		}

		//	Transfer coins should succeed
		s.sendMsgSend(
			s.chainA,
			0,
			periodicVestingAddr.String(),
			Address(),
			vestingTransferAmount.String(),
			fees.String(),
			false,
		)
	})
}

func generateVestingPeriod() ([]byte, error) {
	p := vestingPeriod{
		StartTime: time.Now().Add(90 * time.Second).Unix(),
		Periods: []period{
			{
				Coins:  "850000000uatom",
				Length: 30,
			},
			{
				Coins:  "850000000uatom",
				Length: 30,
			},
		},
	}
	return json.Marshal(p)
}
