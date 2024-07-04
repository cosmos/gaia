package e2e

import (
	"encoding/json"
	"math/rand"
	"path/filepath"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	delayedVestingKey    = "delayed_vesting"
	continuousVestingKey = "continuous_vesting"
	lockedVestingKey     = "locker_vesting"
	periodicVestingKey   = "periodic_vesting"

	vestingPeriodFile = "test_period.json"
	vestingTxDelay    = 5
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
	genesisVestingKeys      = []string{continuousVestingKey, delayedVestingKey, lockedVestingKey, periodicVestingKey}
	vestingAmountVested     = sdk.NewCoin(uatomDenom, math.NewInt(99900000000))
	vestingAmount           = sdk.NewCoin(uatomDenom, math.NewInt(350000))
	vestingBalance          = sdk.NewCoins(vestingAmountVested).Add(vestingAmount)
	vestingDelegationAmount = sdk.NewCoin(uatomDenom, math.NewInt(500000000))
	vestingDelegationFees   = sdk.NewCoin(uatomDenom, math.NewInt(1))
)

func (s *IntegrationTestSuite) testDelayedVestingAccount(api string) {
	var (
		valIdx            = 0
		chain             = s.chainA
		val               = chain.validators[valIdx]
		vestingDelayedAcc = chain.genesisVestingAccounts[delayedVestingKey]
	)
	sender, _ := val.keyInfo.GetAddress()
	valOpAddr := sdk.ValAddress(sender).String()

	s.Run("test delayed vesting genesis account", func() {
		acc, err := queryDelayedVestingAccount(api, vestingDelayedAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, vestingDelayedAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.execDelegate(chain, valIdx, vestingDelegationAmount.String(), valOpAddr,
			vestingDelayedAcc.String(), gaiaHomePath, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, vestingDelayedAcc.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(math.LegacyNewDecFromInt(vestingDelegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		waitTime := acc.EndTime - time.Now().Unix()
		if waitTime > vestingTxDelay {
			//	Transfer coins should fail
			balance, err := getSpecificBalance(api, vestingDelayedAcc.String(), uatomDenom)
			s.Require().NoError(err)
			s.execBankSend(
				chain,
				valIdx,
				vestingDelayedAcc.String(),
				Address(),
				balance.Sub(standardFees).String(),
				standardFees.String(),
				true,
			)
			waitTime = acc.EndTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitTime) * time.Second)
		}

		//	Transfer coins should succeed
		balance, err = getSpecificBalance(api, vestingDelayedAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.execBankSend(
			chain,
			valIdx,
			vestingDelayedAcc.String(),
			Address(),
			balance.Sub(standardFees).String(),
			standardFees.String(),
			false,
		)
	})
}

func (s *IntegrationTestSuite) testContinuousVestingAccount(api string) {
	s.Run("test continuous vesting genesis account", func() {
		var (
			valIdx               = 0
			chain                = s.chainA
			val                  = chain.validators[valIdx]
			continuousVestingAcc = chain.genesisVestingAccounts[continuousVestingKey]
		)
		sender, _ := val.keyInfo.GetAddress()
		valOpAddr := sdk.ValAddress(sender).String()

		acc, err := queryContinuousVestingAccount(api, continuousVestingAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, continuousVestingAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(uatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.execDelegate(chain, valIdx, vestingDelegationAmount.String(),
			valOpAddr, continuousVestingAcc.String(), gaiaHomePath, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, continuousVestingAcc.String())
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(math.LegacyNewDecFromInt(vestingDelegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		waitStartTime := acc.StartTime - time.Now().Unix()
		if waitStartTime > vestingTxDelay {
			//	Transfer coins should fail
			balance, err := getSpecificBalance(api, continuousVestingAcc.String(), uatomDenom)
			s.Require().NoError(err)
			s.execBankSend(
				chain,
				valIdx,
				continuousVestingAcc.String(),
				Address(),
				balance.Sub(standardFees).String(),
				standardFees.String(),
				true,
			)
			waitStartTime = acc.StartTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitStartTime) * time.Second)
		}

		waitEndTime := acc.EndTime - time.Now().Unix()
		if waitEndTime > vestingTxDelay {
			//	Transfer coins should fail
			balance, err := getSpecificBalance(api, continuousVestingAcc.String(), uatomDenom)
			s.Require().NoError(err)
			s.execBankSend(
				chain,
				valIdx,
				continuousVestingAcc.String(),
				Address(),
				balance.Sub(standardFees).String(),
				standardFees.String(),
				true,
			)
			waitEndTime = acc.EndTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitEndTime) * time.Second)
		}

		//	Transfer coins should succeed
		balance, err = getSpecificBalance(api, continuousVestingAcc.String(), uatomDenom)
		s.Require().NoError(err)
		s.execBankSend(
			chain,
			valIdx,
			continuousVestingAcc.String(),
			Address(),
			balance.Sub(standardFees).String(),
			standardFees.String(),
			false,
		)
	})
}

func (s *IntegrationTestSuite) testPeriodicVestingAccount(api string) { //nolint:unused

	s.Run("test periodic vesting genesis account", func() {
		var (
			valIdx              = 0
			chain               = s.chainA
			val                 = chain.validators[valIdx]
			periodicVestingAddr = chain.genesisVestingAccounts[periodicVestingKey].String()
		)
		sender, _ := val.keyInfo.GetAddress()
		valOpAddr := sdk.ValAddress(sender).String()

		s.execCreatePeriodicVestingAccount(
			chain,
			periodicVestingAddr,
			filepath.Join(gaiaHomePath, vestingPeriodFile),
			withKeyValue(flagFrom, sender.String()),
		)

		acc, err := queryPeriodicVestingAccount(api, periodicVestingAddr)
		s.Require().NoError(err)

		//	Check address balance
		balance, err := getSpecificBalance(api, periodicVestingAddr, uatomDenom)
		s.Require().NoError(err)

		expectedBalance := sdk.NewCoin(uatomDenom, math.NewInt(0))
		for _, period := range acc.VestingPeriods {
			// _, coin := ante.Find(period.Amount, uatomDenom)
			_, coin := period.Amount.Find(uatomDenom)
			expectedBalance = expectedBalance.Add(coin)
		}
		s.Require().Equal(expectedBalance, balance)

		waitStartTime := acc.StartTime - time.Now().Unix()
		if waitStartTime > vestingTxDelay {
			//	Transfer coins should fail
			balance, err = getSpecificBalance(api, periodicVestingAddr, uatomDenom)
			s.Require().NoError(err)
			s.execBankSend(
				chain,
				valIdx,
				periodicVestingAddr,
				Address(),
				balance.Sub(standardFees).String(),
				standardFees.String(),
				true,
			)
			waitStartTime = acc.StartTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitStartTime) * time.Second)
		}

		firstPeriod := acc.StartTime + acc.VestingPeriods[0].Length
		waitFirstPeriod := firstPeriod - time.Now().Unix()
		if waitFirstPeriod > vestingTxDelay {
			//	Transfer coins should fail
			balance, err = getSpecificBalance(api, periodicVestingAddr, uatomDenom)
			s.Require().NoError(err)
			s.execBankSend(
				chain,
				valIdx,
				periodicVestingAddr,
				Address(),
				balance.Sub(standardFees).String(),
				standardFees.String(),
				true,
			)
			waitFirstPeriod = firstPeriod - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitFirstPeriod) * time.Second)
		}

		// Delegate coins should succeed
		s.execDelegate(chain, valIdx, vestingDelegationAmount.String(), valOpAddr,
			periodicVestingAddr, gaiaHomePath, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := queryDelegation(api, valOpAddr, periodicVestingAddr)
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(math.LegacyNewDecFromInt(vestingDelegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should succeed
		balance, err = getSpecificBalance(api, periodicVestingAddr, uatomDenom)
		s.Require().NoError(err)
		s.execBankSend(
			chain,
			valIdx,
			periodicVestingAddr,
			Address(),
			balance.Sub(standardFees).String(),
			standardFees.String(),
			false,
		)

		secondPeriod := firstPeriod + acc.VestingPeriods[1].Length
		waitSecondPeriod := secondPeriod - time.Now().Unix()
		if waitSecondPeriod > vestingTxDelay {
			time.Sleep(time.Duration(waitSecondPeriod) * time.Second)

			//	Transfer coins should succeed
			balance, err = getSpecificBalance(api, periodicVestingAddr, uatomDenom)
			s.Require().NoError(err)
			s.execBankSend(
				chain,
				valIdx,
				periodicVestingAddr,
				Address(),
				balance.Sub(standardFees).String(),
				standardFees.String(),
				false,
			)
		}
	})
}

// generateVestingPeriod generate the vesting period file
func generateVestingPeriod() ([]byte, error) {
	p := vestingPeriod{
		StartTime: time.Now().Add(time.Duration(rand.Intn(20)+95) * time.Second).Unix(),
		Periods: []period{
			{
				Coins:  "850000000" + uatomDenom,
				Length: 35,
			},
			{
				Coins:  "2000000000" + uatomDenom,
				Length: 35,
			},
		},
	}
	return json.Marshal(p)
}
