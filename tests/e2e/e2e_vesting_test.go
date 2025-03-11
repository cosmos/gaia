package e2e

import (
	"encoding/json"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
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
	vestingAmountVested     = sdk.NewCoin(common.UatomDenom, math.NewInt(99900000000))
	vestingAmount           = sdk.NewCoin(common.UatomDenom, math.NewInt(350000))
	vestingBalance          = sdk.NewCoins(vestingAmountVested).Add(vestingAmount)
	vestingDelegationAmount = sdk.NewCoin(common.UatomDenom, math.NewInt(500000000))
	vestingDelegationFees   = sdk.NewCoin(common.UatomDenom, math.NewInt(1))
)

func (s *IntegrationTestSuite) testDelayedVestingAccount(api string) {
	var (
		valIdx            = 0
		chain             = s.commonHelper.Resources.ChainA
		val               = chain.Validators[valIdx]
		vestingDelayedAcc = chain.GenesisVestingAccounts[delayedVestingKey]
	)
	sender, _ := val.KeyInfo.GetAddress()
	valOpAddr := sdk.ValAddress(sender).String()

	s.Run("test delayed vesting genesis account", func() {
		acc, err := query.QueryDelayedVestingAccount(api, vestingDelayedAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := query.GetSpecificBalance(api, vestingDelayedAcc.String(), common.UatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(common.UatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.tx.ExecDelegate(chain, valIdx, vestingDelegationAmount.String(), valOpAddr,
			vestingDelayedAcc.String(), common.GaiaHomePath, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := query.QueryDelegation(api, valOpAddr, vestingDelayedAcc.String())
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
			balance, err := query.GetSpecificBalance(api, vestingDelayedAcc.String(), common.UatomDenom)
			s.Require().NoError(err)
			s.tx.ExecBankSend(
				chain,
				valIdx,
				vestingDelayedAcc.String(),
				common.Address(),
				balance.Sub(common.StandardFees).String(),
				common.StandardFees.String(),
				true,
			)
			waitTime = acc.EndTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitTime) * time.Second)
		}

		//	Transfer coins should succeed
		balance, err = query.GetSpecificBalance(api, vestingDelayedAcc.String(), common.UatomDenom)
		s.Require().NoError(err)
		s.tx.ExecBankSend(
			chain,
			valIdx,
			vestingDelayedAcc.String(),
			common.Address(),
			balance.Sub(common.StandardFees).String(),
			common.StandardFees.String(),
			false,
		)
	})
}

func (s *IntegrationTestSuite) testContinuousVestingAccount(api string) {
	s.Run("test continuous vesting genesis account", func() {
		var (
			valIdx               = 0
			chain                = s.commonHelper.Resources.ChainA
			val                  = chain.Validators[valIdx]
			continuousVestingAcc = chain.GenesisVestingAccounts[continuousVestingKey]
		)
		sender, _ := val.KeyInfo.GetAddress()
		valOpAddr := sdk.ValAddress(sender).String()

		acc, err := query.QueryContinuousVestingAccount(api, continuousVestingAcc.String())
		s.Require().NoError(err)

		//	Check address balance
		balance, err := query.GetSpecificBalance(api, continuousVestingAcc.String(), common.UatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(vestingBalance.AmountOf(common.UatomDenom), balance.Amount)

		// Delegate coins should succeed
		s.tx.ExecDelegate(chain, valIdx, vestingDelegationAmount.String(),
			valOpAddr, continuousVestingAcc.String(), common.GaiaHomePath, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := query.QueryDelegation(api, valOpAddr, continuousVestingAcc.String())
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
			balance, err := query.GetSpecificBalance(api, continuousVestingAcc.String(), common.UatomDenom)
			s.Require().NoError(err)
			s.tx.ExecBankSend(
				chain,
				valIdx,
				continuousVestingAcc.String(),
				common.Address(),
				balance.Sub(common.StandardFees).String(),
				common.StandardFees.String(),
				true,
			)
			waitStartTime = acc.StartTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitStartTime) * time.Second)
		}

		waitEndTime := acc.EndTime - time.Now().Unix()
		if waitEndTime > vestingTxDelay {
			//	Transfer coins should fail
			balance, err := query.GetSpecificBalance(api, continuousVestingAcc.String(), common.UatomDenom)
			s.Require().NoError(err)
			s.tx.ExecBankSend(
				chain,
				valIdx,
				continuousVestingAcc.String(),
				common.Address(),
				balance.Sub(common.StandardFees).String(),
				common.StandardFees.String(),
				true,
			)
			waitEndTime = acc.EndTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitEndTime) * time.Second)
		}

		//	Transfer coins should succeed
		balance, err = query.GetSpecificBalance(api, continuousVestingAcc.String(), common.UatomDenom)
		s.Require().NoError(err)
		s.tx.ExecBankSend(
			chain,
			valIdx,
			continuousVestingAcc.String(),
			common.Address(),
			balance.Sub(common.StandardFees).String(),
			common.StandardFees.String(),
			false,
		)
	})
}

func (s *IntegrationTestSuite) testPeriodicVestingAccount(api string) { //nolint:unused

	s.Run("test periodic vesting genesis account", func() {
		var (
			valIdx              = 0
			chain               = s.commonHelper.Resources.ChainA
			val                 = chain.Validators[valIdx]
			periodicVestingAddr = chain.GenesisVestingAccounts[periodicVestingKey].String()
		)
		sender, _ := val.KeyInfo.GetAddress()
		valOpAddr := sdk.ValAddress(sender).String()

		s.tx.ExecCreatePeriodicVestingAccount(
			chain,
			periodicVestingAddr,
			filepath.Join(common.GaiaHomePath, vestingPeriodFile),
			common.WithKeyValue(common.FlagFrom, sender.String()),
		)

		acc, err := query.QueryPeriodicVestingAccount(api, periodicVestingAddr)
		s.Require().NoError(err)

		//	Check address balance
		balance, err := query.GetSpecificBalance(api, periodicVestingAddr, common.UatomDenom)
		s.Require().NoError(err)

		expectedBalance := sdk.NewCoin(common.UatomDenom, math.NewInt(0))
		for _, period := range acc.VestingPeriods {
			// _, coin := ante.Find(period.Amount, uatomDenom)
			_, coin := period.Amount.Find(common.UatomDenom)
			expectedBalance = expectedBalance.Add(coin)
		}
		s.Require().Equal(expectedBalance, balance)

		waitStartTime := acc.StartTime - time.Now().Unix()
		if waitStartTime > vestingTxDelay {
			//	Transfer coins should fail
			balance, err = query.GetSpecificBalance(api, periodicVestingAddr, common.UatomDenom)
			s.Require().NoError(err)
			s.tx.ExecBankSend(
				chain,
				valIdx,
				periodicVestingAddr,
				common.Address(),
				balance.Sub(common.StandardFees).String(),
				common.StandardFees.String(),
				true,
			)
			waitStartTime = acc.StartTime - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitStartTime) * time.Second)
		}

		firstPeriod := acc.StartTime + acc.VestingPeriods[0].Length
		waitFirstPeriod := firstPeriod - time.Now().Unix()
		if waitFirstPeriod > vestingTxDelay {
			//	Transfer coins should fail
			balance, err = query.GetSpecificBalance(api, periodicVestingAddr, common.UatomDenom)
			s.Require().NoError(err)
			s.tx.ExecBankSend(
				chain,
				valIdx,
				periodicVestingAddr,
				common.Address(),
				balance.Sub(common.StandardFees).String(),
				common.StandardFees.String(),
				true,
			)
			waitFirstPeriod = firstPeriod - time.Now().Unix() + vestingTxDelay
			time.Sleep(time.Duration(waitFirstPeriod) * time.Second)
		}

		// Delegate coins should succeed
		s.tx.ExecDelegate(chain, valIdx, vestingDelegationAmount.String(), valOpAddr,
			periodicVestingAddr, common.GaiaHomePath, vestingDelegationFees.String())

		// Validate delegation successful
		s.Require().Eventually(
			func() bool {
				res, err := query.QueryDelegation(api, valOpAddr, periodicVestingAddr)
				amt := res.GetDelegationResponse().GetDelegation().GetShares()
				s.Require().NoError(err)

				return amt.Equal(math.LegacyNewDecFromInt(vestingDelegationAmount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)

		//	Transfer coins should succeed
		balance, err = query.GetSpecificBalance(api, periodicVestingAddr, common.UatomDenom)
		s.Require().NoError(err)
		s.tx.ExecBankSend(
			chain,
			valIdx,
			periodicVestingAddr,
			common.Address(),
			balance.Sub(common.StandardFees).String(),
			common.StandardFees.String(),
			false,
		)

		secondPeriod := firstPeriod + acc.VestingPeriods[1].Length
		waitSecondPeriod := secondPeriod - time.Now().Unix()
		if waitSecondPeriod > vestingTxDelay {
			time.Sleep(time.Duration(waitSecondPeriod) * time.Second)

			//	Transfer coins should succeed
			balance, err = query.GetSpecificBalance(api, periodicVestingAddr, common.UatomDenom)
			s.Require().NoError(err)
			s.tx.ExecBankSend(
				chain,
				valIdx,
				periodicVestingAddr,
				common.Address(),
				balance.Sub(common.StandardFees).String(),
				common.StandardFees.String(),
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
				Coins:  "850000000" + common.UatomDenom,
				Length: 35,
			},
			{
				Coins:  "2000000000" + common.UatomDenom,
				Length: 35,
			},
		},
	}
	return json.Marshal(p)
}
