package integration

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	liquidkeeper "github.com/cosmos/gaia/v26/x/liquid/keeper"
	liquidtypes "github.com/cosmos/gaia/v26/x/liquid/types"
)

func TestTokenizeSharesAndRedeemTokens(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	stakingKeeper := f.stakingKeeper

	liquidStakingCapStrict := math.LegacyZeroDec()
	liquidStakingCapConservative := math.LegacyMustNewDecFromStr("0.8")
	liquidStakingCapDisabled := math.LegacyOneDec()

	testCases := []struct {
		name                          string
		vestingAmount                 math.Int
		delegationAmount              math.Int
		tokenizeShareAmount           math.Int
		redeemAmount                  math.Int
		targetVestingDelAfterShare    math.Int
		targetVestingDelAfterRedeem   math.Int
		globalLiquidStakingCap        math.LegacyDec
		slashFactor                   math.LegacyDec
		validatorLiquidStakingCap     math.LegacyDec
		expTokenizeErr                bool
		expRedeemErr                  bool
		prevAccountDelegationExists   bool
		recordAccountDelegationExists bool
	}{
		{
			name:                          "full amount tokenize and redeem",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                   math.LegacyZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: false,
		},
		{
			name:                          "full amount tokenize and partial redeem",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   math.LegacyZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: true,
		},
		{
			name:                          "partial amount tokenize and full redeem",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   math.LegacyZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   true,
			recordAccountDelegationExists: false,
		},
		{
			name:                          "tokenize and redeem with slash",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   math.LegacyMustNewDecFromStr("0.1"),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: true,
		},
		{
			name:                      "over tokenize",
			vestingAmount:             math.NewInt(0),
			delegationAmount:          stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:       stakingKeeper.TokensFromConsensusPower(ctx, 30),
			redeemAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:               math.LegacyZeroDec(),
			globalLiquidStakingCap:    liquidStakingCapDisabled,
			validatorLiquidStakingCap: liquidStakingCapDisabled,
			expTokenizeErr:            true,
			expRedeemErr:              false,
		},
		{
			name:                      "over redeem",
			vestingAmount:             math.NewInt(0),
			delegationAmount:          stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:       stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 40),
			slashFactor:               math.LegacyZeroDec(),
			globalLiquidStakingCap:    liquidStakingCapDisabled,
			validatorLiquidStakingCap: liquidStakingCapDisabled,
			expTokenizeErr:            false,
			expRedeemErr:              true,
		},
		{
			name:                        "vesting account tokenize share failure",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "vesting account tokenize share success",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict global liquid staking cap - tokenization fails",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapStrict,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative global liquid staking cap - successful tokenization",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict validator liquid staking cap - tokenization fails",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapStrict,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative validator liquid staking cap - successful tokenization",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "all caps set conservatively - successful tokenize share",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "delegator is a liquid staking provider - accounting should not update",
			vestingAmount:               math.ZeroInt(),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			f := initFixture(t)

			ctx := f.sdkCtx
			var (
				bankKeeper    = f.bankKeeper
				accountKeeper = f.accountKeeper
				stakingKeeper = f.stakingKeeper
				liquidKeeper  = f.liquidKeeper
			)
			addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
			addrAcc1, addrAcc2 := addrs[0], addrs[1]
			addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

			// Fund module account
			bondDenom, err := stakingKeeper.BondDenom(ctx)
			require.NoError(t, err)
			delegationCoin := sdk.NewCoin(bondDenom, tc.delegationAmount)
			err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(delegationCoin))
			require.NoError(t, err)

			// set the delegator address depending on whether the delegator should be a liquid staking provider
			delegatorAccount := addrAcc2

			// set validator bond factor and global liquid staking cap
			params, err := liquidKeeper.GetParams(ctx)
			require.NoError(t, err)
			params.GlobalLiquidStakingCap = tc.globalLiquidStakingCap
			params.ValidatorLiquidStakingCap = tc.validatorLiquidStakingCap
			require.NoError(t, liquidKeeper.SetParams(ctx, params))

			// set the total liquid staked tokens
			liquidKeeper.SetTotalLiquidStakedTokens(ctx, math.ZeroInt())

			if !tc.vestingAmount.IsZero() {
				// create vesting account
				acc2 := accountKeeper.GetAccount(ctx, addrAcc2).(*authtypes.BaseAccount)
				initialVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tc.vestingAmount))
				baseVestingWithCoins, err := vestingtypes.NewBaseVestingAccount(acc2, initialVesting, time.Now().Unix()+86400*365)
				require.NoError(t, err)
				delayedVestingAccount := vestingtypes.NewDelayedVestingAccountRaw(baseVestingWithCoins)
				accountKeeper.SetAccount(ctx, delayedVestingAccount)
			}

			pubKeys := simtestutil.CreateTestPubKeys(2)
			pk1, pk2 := pubKeys[0], pubKeys[1]

			// Create Validators and Delegation
			val1 := testutil.NewValidator(t, addrVal1, pk1)
			val1.Status = stakingtypes.Bonded
			err = stakingKeeper.SetValidator(ctx, val1)
			require.NoError(t, err)
			err = stakingKeeper.SetValidatorByPowerIndex(ctx, val1)
			require.NoError(t, err)
			err = stakingKeeper.SetValidatorByConsAddr(ctx, val1)
			require.NoError(t, err)
			err = liquidKeeper.SetLiquidValidator(ctx, liquidtypes.NewLiquidValidator(val1.OperatorAddress))
			require.NoError(t, err)

			val2 := testutil.NewValidator(t, addrVal2, pk2)
			val2.Status = stakingtypes.Bonded
			err = stakingKeeper.SetValidator(ctx, val2)
			require.NoError(t, err)
			err = stakingKeeper.SetValidatorByPowerIndex(ctx, val2)
			require.NoError(t, err)
			err = stakingKeeper.SetValidatorByConsAddr(ctx, val2)
			require.NoError(t, err)
			err = liquidKeeper.SetLiquidValidator(ctx, liquidtypes.NewLiquidValidator(val2.OperatorAddress))
			require.NoError(t, err)

			// Delegate from both the main delegator as well as a random account so there is a
			// non-zero delegation after redemption
			err = delegateCoinsFromAccount(ctx, *stakingKeeper, delegatorAccount, tc.delegationAmount, val1)
			require.NoError(t, err)

			// apply TM updates
			applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)

			_, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			require.NoError(t, err, "delegation not found after delegate")

			lastRecordID := liquidKeeper.GetLastTokenizeShareRecordID(ctx)
			oldValidator, err := stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)

			// skServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
			msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)

			resp, err := msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
				DelegatorAddress:    delegatorAccount.String(),
				ValidatorAddress:    addrVal1.String(),
				Amount:              sdk.NewCoin(bondDenom, tc.tokenizeShareAmount),
				TokenizedShareOwner: delegatorAccount.String(),
			})
			if tc.expTokenizeErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check last record id increase
			require.Equal(t, lastRecordID+1, liquidKeeper.GetLastTokenizeShareRecordID(ctx))

			// ensure validator's total tokens is consistent
			newValidator, err := stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)
			newLiquidVal, err := liquidKeeper.GetLiquidValidator(ctx, addrVal1)
			require.NoError(t, err)

			// if the delegator was not a provider, check that the total liquid staked and validator liquid shares increased
			totalLiquidTokensAfterTokenization := liquidKeeper.GetTotalLiquidStakedTokens(ctx)
			validatorLiquidSharesAfterTokenization := newLiquidVal.LiquidShares
			require.Equal(t, tc.tokenizeShareAmount.String(), totalLiquidTokensAfterTokenization.String(), "total liquid tokens after tokenization")
			require.Equal(t, tc.tokenizeShareAmount.String(), validatorLiquidSharesAfterTokenization.TruncateInt().String(), "validator liquid shares after tokenization")

			if tc.vestingAmount.IsPositive() {
				acc := accountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(bondDenom).String(), tc.targetVestingDelAfterShare.String())
			}

			if tc.prevAccountDelegationExists {
				_, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
				require.NoError(t, err, "delegation not found after partial tokenize share")
			} else {
				_, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
				require.ErrorIs(t, err, stakingtypes.ErrNoDelegation, "delegation found after full tokenize share")
			}

			shareToken := bankKeeper.GetBalance(ctx, delegatorAccount, resp.Amount.Denom)
			require.Equal(t, resp.Amount, shareToken)
			_, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err, "validator not found")

			records := liquidKeeper.GetAllTokenizeShareRecords(ctx)
			require.Len(t, records, 1)
			delegation, err := stakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
			require.NoError(t, err, "delegation not found from tokenize share module account after tokenize share")

			// slash before redeem
			slashedTokens := math.ZeroInt()
			redeemedShares := tc.redeemAmount
			redeemedTokens := tc.redeemAmount
			if tc.slashFactor.IsPositive() {
				consAddr, err := val1.GetConsAddr()
				require.NoError(t, err)
				ctx = ctx.WithBlockHeight(100)
				val1, err = stakingKeeper.GetValidator(ctx, addrVal1)
				require.NoError(t, err)
				power := stakingKeeper.TokensToConsensusPower(ctx, val1.Tokens)
				_, err = stakingKeeper.Slash(ctx, consAddr, 10, power, tc.slashFactor)
				require.NoError(t, err)
				slashedTokens = math.LegacyNewDecFromInt(val1.Tokens).Mul(tc.slashFactor).TruncateInt()

				val1, _ := stakingKeeper.GetValidator(ctx, addrVal1)
				redeemedTokens = val1.TokensFromShares(math.LegacyNewDecFromInt(redeemedShares)).TruncateInt()
			}

			// get delegator balance and delegation
			bondDenomAmountBefore := bankKeeper.GetBalance(ctx, delegatorAccount, bondDenom)
			val1, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			delegation, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			if errors.Is(err, stakingtypes.ErrNoDelegation) {
				delegation = stakingtypes.Delegation{Shares: math.LegacyZeroDec()}
			}
			delAmountBefore := val1.TokensFromShares(delegation.Shares)
			oldValidator, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)

			_, err = msgServer.RedeemTokensForShares(ctx, &liquidtypes.MsgRedeemTokensForShares{
				DelegatorAddress: delegatorAccount.String(),
				Amount:           sdk.NewCoin(resp.Amount.Denom, tc.redeemAmount),
			})
			if tc.expRedeemErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// ensure validator's total tokens is consistent
			newLiquidVal, err = liquidKeeper.GetLiquidValidator(ctx, addrVal1)
			require.NoError(t, err)
			newValidator, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)

			// if the delegator was not a liquid staking provider, check that the total liquid staked
			// and liquid shares decreased
			totalLiquidTokensAfterRedemption := liquidKeeper.GetTotalLiquidStakedTokens(ctx)
			validatorLiquidSharesAfterRedemption := newLiquidVal.LiquidShares
			expectedLiquidTokens := totalLiquidTokensAfterTokenization.Sub(redeemedTokens).Sub(slashedTokens)
			expectedLiquidShares := validatorLiquidSharesAfterTokenization.Sub(math.LegacyNewDecFromInt(redeemedShares))
			require.Equal(t, expectedLiquidTokens.String(), totalLiquidTokensAfterRedemption.String(), "total liquid tokens after redemption")
			require.Equal(t, expectedLiquidShares.String(), validatorLiquidSharesAfterRedemption.String(), "validator liquid shares after tokenization")

			if tc.vestingAmount.IsPositive() {
				acc := accountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(bondDenom).String(), tc.targetVestingDelAfterRedeem.String())
			}

			expectedDelegatedShares := math.LegacyNewDecFromInt(tc.delegationAmount.Sub(tc.tokenizeShareAmount).Add(tc.redeemAmount))
			delegation, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			require.NoError(t, err, "delegation not found after redeem tokens")
			require.Equal(t, delegatorAccount.String(), delegation.DelegatorAddress)
			require.Equal(t, addrVal1.String(), delegation.ValidatorAddress)
			require.Equal(t, expectedDelegatedShares, delegation.Shares, "delegation shares after redeem")

			// check delegator balance is not changed
			bondDenomAmountAfter := bankKeeper.GetBalance(ctx, delegatorAccount, bondDenom)
			require.Equal(t, bondDenomAmountAfter.Amount.String(), bondDenomAmountBefore.Amount.String())

			// get delegation amount is changed correctly
			val1, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			delegation, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			if errors.Is(err, stakingtypes.ErrNoDelegation) {
				delegation = stakingtypes.Delegation{Shares: math.LegacyZeroDec()}
			}
			delAmountAfter := val1.TokensFromShares(delegation.Shares)
			require.Equal(t, delAmountAfter.String(), delAmountBefore.Add(math.LegacyNewDecFromInt(tc.redeemAmount).Mul(math.LegacyOneDec().Sub(tc.slashFactor))).String())

			shareToken = bankKeeper.GetBalance(ctx, delegatorAccount, resp.Amount.Denom)
			require.Equal(t, shareToken.Amount.String(), tc.tokenizeShareAmount.Sub(tc.redeemAmount).String())
			_, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err, "validator not found")

			if tc.recordAccountDelegationExists {
				_, err = stakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
				require.NoError(t, err, "delegation not found from tokenize share module account after redeem partial amount")

				records = liquidKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 1)
			} else {
				_, err = stakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
				require.True(t, errors.Is(err, stakingtypes.ErrNoDelegation),
					"delegation found from tokenize share module account after redeem full amount")

				records = liquidKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 0)
			}
		})
	}
}

func TestRedelegationTokenization(t *testing.T) {
	// Test that a delegator with ongoing redelegation cannot
	// tokenize any shares until the redelegation is complete.
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		liquidKeeper  = f.liquidKeeper
	)
	skServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)
	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk1 := pubKeys[0]

	// Create Validators and Delegation
	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	alice := addrs[0]

	validatorAAddress := sdk.ValAddress(addrs[1])
	val1 := testutil.NewValidator(t, validatorAAddress, pk1)
	val1.Status = stakingtypes.Bonded
	require.NoError(t, stakingKeeper.SetValidator(ctx, val1))
	require.NoError(t, stakingKeeper.SetValidatorByPowerIndex(ctx, val1))
	err := stakingKeeper.SetValidatorByConsAddr(ctx, val1)
	require.NoError(t, err)

	_, validatorBAddress := setupTestTokenizeAndRedeemConversion(t, *liquidKeeper, *stakingKeeper, bankKeeper, ctx)

	delegateAmount := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	bondedDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondedDenom, delegateAmount)

	// Alice delegates to validatorA
	_, err = skServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: alice.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err)

	// Alice redelegates to validatorB
	redelegateAmount := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	redelegateCoin := sdk.NewCoin(bondedDenom, redelegateAmount)
	_, err = skServer.BeginRedelegate(ctx, &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    alice.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorBAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err)

	redelegation, err := stakingKeeper.GetRedelegations(ctx, alice, uint16(10))
	require.NoError(t, err)
	require.Len(t, redelegation, 1, "expect one redelegation")
	require.Len(t, redelegation[0].Entries, 1, "expect one redelegation entry")

	// Alice attempts to tokenize the redelegation, but this fails because the redelegation is ongoing
	tokenizedAmount := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	tokenizedCoin := sdk.NewCoin(bondedDenom, tokenizedAmount)
	_, err = msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    alice.String(),
		ValidatorAddress:    validatorBAddress.String(),
		Amount:              tokenizedCoin,
		TokenizedShareOwner: alice.String(),
	})
	require.Error(t, err)
	require.Equal(t, liquidtypes.ErrRedelegationInProgress, err)

	// Check that the redelegation is still present
	redelegation, err = stakingKeeper.GetRedelegations(ctx, alice, uint16(10))
	require.NoError(t, err)
	require.Len(t, redelegation, 1, "expect one redelegation")
	require.Len(t, redelegation[0].Entries, 1, "expect one redelegation entry")

	// advance time until the redelegations should mature
	// end block
	_, err = f.stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// advance by 22 days
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(22 * 24 * time.Hour))
	headerInfo := ctx.HeaderInfo()
	headerInfo.Time = ctx.BlockHeader().Time
	headerInfo.Height = ctx.BlockHeader().Height
	ctx = ctx.WithHeaderInfo(headerInfo)
	// begin block
	require.NoError(t, f.stakingKeeper.BeginBlocker(ctx))
	// end block
	_, err = f.stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// check that the redelegation is removed
	redelegation, err = stakingKeeper.GetRedelegations(ctx, alice, uint16(10))
	require.NoError(t, err)
	require.Len(t, redelegation, 0, "expect no redelegations")

	// Alice attempts to tokenize the redelegation again, and this time it should succeed
	// because there is no ongoing redelegation
	_, err = msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    alice.String(),
		ValidatorAddress:    validatorBAddress.String(),
		Amount:              tokenizedCoin,
		TokenizedShareOwner: alice.String(),
	})
	require.NoError(t, err)

	// Check that the tokenization was successful
	shareRecord, err := liquidKeeper.GetTokenizeShareRecord(ctx, liquidKeeper.GetLastTokenizeShareRecordID(ctx))
	require.NoError(t, err, "expect to find token share record")
	require.Equal(t, alice.String(), shareRecord.Owner)
	require.Equal(t, validatorBAddress.String(), shareRecord.Validator)
}

// Helper function to setup a delegator and validator for the Tokenize/Redeem conversion tests
func setupTestTokenizeAndRedeemConversion(
	t *testing.T,
	lk liquidkeeper.Keeper,
	sk stakingkeeper.Keeper,
	bk bankkeeper.Keeper,
	ctx sdk.Context,
) (delAddress sdk.AccAddress, valAddress sdk.ValAddress) {
	t.Helper()
	addresses := simtestutil.AddTestAddrs(bk, sk, ctx, 2, math.NewInt(1_000_000))

	pubKeys := simtestutil.CreateTestPubKeys(1)

	delegatorAddress := addresses[0]
	validatorAddress := sdk.ValAddress(addresses[1])

	validator, err := stakingtypes.NewValidator(validatorAddress.String(), pubKeys[0], stakingtypes.Description{})
	require.NoError(t, err)
	liquidVal := liquidtypes.NewLiquidValidator(validatorAddress.String())
	validator.DelegatorShares = math.LegacyNewDec(1_000_000)
	validator.Tokens = math.NewInt(1_000_000)
	validator.Status = stakingtypes.Bonded

	_ = sk.SetValidator(ctx, validator)
	_ = sk.SetValidatorByConsAddr(ctx, validator)
	_ = lk.SetLiquidValidator(ctx, liquidVal)

	return delegatorAddress, validatorAddress
}

// Simulate a slash by decrementing the validator's tokens
// We'll do this in a way such that the exchange rate is not an even integer
// and the shares associated with a delegation will have a long decimal
func simulateSlashWithImprecision(t *testing.T, sk stakingkeeper.Keeper, ctx sdk.Context, valAddress sdk.ValAddress) {
	t.Helper()
	validator, err := sk.GetValidator(ctx, valAddress)
	require.NoError(t, err)

	slashMagnitude := math.LegacyMustNewDecFromStr("0.1111111111")
	slashTokens := math.LegacyNewDecFromInt(validator.Tokens).Mul(slashMagnitude).TruncateInt()
	validator.Tokens = validator.Tokens.Sub(slashTokens)

	require.NoError(t, sk.SetValidator(ctx, validator))
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Slash -> Delegate -> Tokenize -> Redeem
// Note, in this example, there 2 tokens are lost during the decimal to int conversion
// during the unbonding step within tokenization and redemption
func TestTokenizeAndRedeemConversion_SlashBeforeDelegation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		liquidKepeer  = f.liquidKeeper
	)
	skServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	liquidServer := liquidkeeper.NewMsgServerImpl(liquidKepeer)
	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, *liquidKepeer, *stakingKeeper,
		bankKeeper, ctx)

	// slash the validator
	simulateSlashWithImprecision(t, *stakingKeeper, ctx, validatorAddress)
	validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err)

	// Delegate and confirm the delegation record was created
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	_, err = skServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	delegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// Tokenize the full delegation amount
	_, err = liquidServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              delegateCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// Confirm the number of shareTokens equals the number of shares truncated
	// Note: 1 token is lost during unbonding due to rounding
	shareDenom := validatorAddress.String() + "/1"
	shareToken := bankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens := delegation.Shares.TruncateInt().Int64() - 1 // 1 token was lost during unbonding
	require.Equal(t, expectedShareTokens, shareToken.Amount.Int64(), "share token amount")

	// Redeem the share tokens
	_, err = liquidServer.RedeemTokensForShares(ctx, &liquidtypes.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm (almost) the full delegation was recovered - minus the 2 tokens from the precision error
	// (1 occurs during tokenization, and 1 occurs during redemption)
	newDelegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	expectedDelegationTokens := delegateAmount.Int64() - 2
	require.Equal(t, expectedDelegationTokens, endDelegationTokens, "final delegation tokens")
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Delegate -> Slash -> Tokenize -> Redeem
// Note, in this example, there 1 token lost during the decimal to int conversion
// during the unbonding step within tokenization
func TestTokenizeAndRedeemConversion_SlashBeforeTokenization(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		liquidKeeper  = f.liquidKeeper
	)
	msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)
	skServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, *liquidKeeper, *stakingKeeper,
		bankKeeper, ctx)

	// Delegate and confirm the delegation record was created
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	_, err = skServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	_, err = stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// slash the validator
	simulateSlashWithImprecision(t, *stakingKeeper, ctx, validatorAddress)
	validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err)

	// Tokenize the new amount after the slash
	delegationAmountAfterSlash := validator.TokensFromShares(math.LegacyNewDecFromInt(delegateAmount)).TruncateInt()
	tokenizationCoin := sdk.NewCoin(bondDenom, delegationAmountAfterSlash)

	_, err = msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              tokenizationCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// The number of share tokens should line up with the **new** number of shares associated
	// with the original delegated amount
	// Note: 1 token is lost during unbonding due to rounding
	shareDenom := validatorAddress.String() + "/1"
	shareToken := bankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens, err := validator.SharesFromTokens(tokenizationCoin.Amount)
	require.NoError(t, err)
	require.Equal(t, expectedShareTokens.TruncateInt().Int64()-1, shareToken.Amount.Int64(), "share token amount")

	// // Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(ctx, &liquidtypes.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm the full tokenization amount was recovered - minus the 1 token from the precision error
	newDelegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	expectedDelegationTokens := delegationAmountAfterSlash.Int64() - 1
	require.Equal(t, expectedDelegationTokens, endDelegationTokens, "final delegation tokens")
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Delegate -> Tokenize -> Slash -> Redeem
// Note, in this example, there 1 token lost during the decimal to int conversion
// during the unbonding step within redemption
func TestTokenizeAndRedeemConversion_SlashBeforeRedemption(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		liquidKeeper  = f.liquidKeeper
	)
	msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)
	skServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, *liquidKeeper, *stakingKeeper,
		bankKeeper, ctx)

	// Delegate and confirm the delegation record was created
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	_, err = skServer.Delegate(ctx, &stakingtypes.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	_, err = stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// Tokenize the full delegation amount
	_, err = msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              delegateCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// The number of share tokens should line up 1:1 with the number of issued shares
	// Since the validator has not been slashed, the shares also line up 1;1
	// with the original delegation amount
	shareDenom := validatorAddress.String() + "/1"
	shareToken := bankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens := delegateAmount
	require.Equal(t, expectedShareTokens.Int64(), shareToken.Amount.Int64(), "share token amount")

	// slash the validator
	simulateSlashWithImprecision(t, *stakingKeeper, ctx, validatorAddress)
	validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err)

	// Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(ctx, &liquidtypes.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm the original delegation, minus the slash, was recovered
	// There's an additional 1 token lost from precision error during unbonding
	delegationAmountAfterSlash := validator.TokensFromShares(math.LegacyNewDecFromInt(delegateAmount)).TruncateInt().Int64()
	newDelegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	require.Equal(t, delegationAmountAfterSlash-1, endDelegationTokens, "final delegation tokens")
}

func TestTransferTokenizeShareRecord(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		liquidKeeper  = f.liquidKeeper
	)
	msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 3, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1, addrAcc2, valAcc := addrs[0], addrs[1], addrs[2]
	addrVal := sdk.ValAddress(valAcc)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk := pubKeys[0]

	val, err := stakingtypes.NewValidator(addrVal.String(), pk, stakingtypes.Description{})
	require.NoError(t, err)

	require.NoError(t, stakingKeeper.SetValidator(ctx, val))
	require.NoError(t, stakingKeeper.SetValidatorByPowerIndex(ctx, val))

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)

	err = liquidKeeper.AddTokenizeShareRecord(ctx, liquidtypes.TokenizeShareRecord{
		Id:            1,
		Owner:         addrAcc1.String(),
		ModuleAccount: "module_account",
		Validator:     val.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.TransferTokenizeShareRecord(ctx, &liquidtypes.MsgTransferTokenizeShareRecord{
		TokenizeShareRecordId: 1,
		Sender:                addrAcc1.String(),
		NewOwner:              addrAcc2.String(),
	})
	require.NoError(t, err)

	record, err := liquidKeeper.GetTokenizeShareRecord(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, record.Owner, addrAcc2.String())

	records := liquidKeeper.GetTokenizeShareRecordsByOwner(ctx, addrAcc1)
	require.Len(t, records, 0)
	records = liquidKeeper.GetTokenizeShareRecordsByOwner(ctx, addrAcc2)
	require.Len(t, records, 1)
}

func TestEnableDisableTokenizeShares(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		liquidKeeper  = f.liquidKeeper
	)
	// Create a delegator and validator
	stakeAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	stakeToken := sdk.NewCoin(bondDenom, stakeAmount)

	addresses := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakeAmount)
	delegatorAddress := addresses[0]

	pubKeys := simtestutil.CreateTestPubKeys(1)
	validatorAddress := sdk.ValAddress(addresses[1])
	validator, err := stakingtypes.NewValidator(validatorAddress.String(), pubKeys[0], stakingtypes.Description{})
	require.NoError(t, err)

	validator.DelegatorShares = math.LegacyNewDec(1_000_000)
	validator.Tokens = math.NewInt(1_000_000)
	validator.Status = stakingtypes.Bonded
	require.NoError(t, stakingKeeper.SetValidator(ctx, validator))
	require.NoError(t, liquidKeeper.SetLiquidValidator(ctx, liquidtypes.NewLiquidValidator(validator.OperatorAddress)))

	// Fix block time and set unbonding period to 1 day
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(blockTime)

	unbondingPeriod := time.Hour * 24
	params, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.UnbondingTime = unbondingPeriod
	require.NoError(t, stakingKeeper.SetParams(ctx, params))
	unlockTime := blockTime.Add(unbondingPeriod)

	// Build test messages (some of which will be reused)
	delegateMsg := stakingtypes.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           stakeToken,
	}
	tokenizeMsg := liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              stakeToken,
		TokenizedShareOwner: delegatorAddress.String(),
	}
	redeemMsg := liquidtypes.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
	}
	disableMsg := liquidtypes.MsgDisableTokenizeShares{
		DelegatorAddress: delegatorAddress.String(),
	}
	enableMsg := liquidtypes.MsgEnableTokenizeShares{
		DelegatorAddress: delegatorAddress.String(),
	}

	msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)
	skServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	// Delegate normally
	_, err = skServer.Delegate(ctx, &delegateMsg)
	require.NoError(t, err, "no error expected when delegating")

	// Tokenize shares - it should succeed
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.NoError(t, err, "no error expected when tokenizing shares for the first time")

	liquidToken := bankKeeper.GetBalance(ctx, delegatorAddress, validatorAddress.String()+"/1")
	require.Equal(t, stakeAmount.Int64(), liquidToken.Amount.Int64(), "user received token after tokenizing share")

	// Redeem to remove all tokenized shares
	redeemMsg.Amount = liquidToken
	_, err = msgServer.RedeemTokensForShares(ctx, &redeemMsg)
	require.NoError(t, err, "no error expected when redeeming")

	// Attempt to enable tokenizing shares when there is no lock in place, it should error
	_, err = msgServer.EnableTokenizeShares(ctx, &enableMsg)
	require.ErrorIs(t, err, liquidtypes.ErrTokenizeSharesAlreadyEnabledForAccount)

	// Attempt to disable when no lock is in place, it should succeed
	_, err = msgServer.DisableTokenizeShares(ctx, &disableMsg)
	require.NoError(t, err, "no error expected when disabling tokenization")

	// Disabling again while the lock is already in place, should error
	_, err = msgServer.DisableTokenizeShares(ctx, &disableMsg)
	require.ErrorIs(t, err, liquidtypes.ErrTokenizeSharesAlreadyDisabledForAccount)

	// Attempt to tokenize, it should fail since tokenization is disabled
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.ErrorIs(t, err, liquidtypes.ErrTokenizeSharesDisabledForAccount)

	// Now enable tokenization
	_, err = msgServer.EnableTokenizeShares(ctx, &enableMsg)
	require.NoError(t, err, "no error expected when enabling tokenization")

	// Attempt to tokenize again, it should still fail since the unbonding period has
	// not passed and the lock is still active
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.ErrorIs(t, err, liquidtypes.ErrTokenizeSharesDisabledForAccount)
	require.ErrorContains(t, err, fmt.Sprintf("tokenization will be allowed at %s",
		blockTime.Add(unbondingPeriod)))

	// Confirm the unlock is queued
	authorizations := liquidKeeper.GetPendingTokenizeShareAuthorizations(ctx, unlockTime)
	require.Equal(t, []string{delegatorAddress.String()}, authorizations.Addresses,
		"pending tokenize share authorizations")

	// Disable tokenization again - it should remove the pending record from the queue
	_, err = msgServer.DisableTokenizeShares(ctx, &disableMsg)
	require.NoError(t, err, "no error expected when re-enabling tokenization")

	authorizations = liquidKeeper.GetPendingTokenizeShareAuthorizations(ctx, unlockTime)
	require.Empty(t, authorizations.Addresses, "there should be no pending authorizations in the queue")

	// Enable one more time
	_, err = msgServer.EnableTokenizeShares(ctx, &enableMsg)
	require.NoError(t, err, "no error expected when enabling tokenization again")

	// Increment the block time by the unbonding period and remove the expired locks
	ctx = ctx.WithBlockTime(unlockTime)
	_, err = liquidKeeper.RemoveExpiredTokenizeShareLocks(ctx, ctx.BlockTime())
	require.NoError(t, err)

	// Attempt to tokenize again, it should succeed this time since the lock has expired
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.NoError(t, err, "no error expected when tokenizing after lock has expired")
}

func TestTokenizeAndRedeemVestedDelegation(t *testing.T) {
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		accountKeeper = f.accountKeeper
		liquidKeeper  = f.liquidKeeper
	)

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1 := addrs[0]
	addrVal1 := sdk.ValAddress(addrAcc1)

	// Original vesting mount (OV)
	originalVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100_000)))
	startTime := time.Now()
	endTime := time.Now().Add(24 * time.Hour)

	// Create vesting account
	lastAccNum := uint64(1000)
	baseAcc := authtypes.NewBaseAccountWithAddress(addrAcc1)
	require.NoError(t, baseAcc.SetAccountNumber(atomic.AddUint64(&lastAccNum, 1)))

	continuousVestingAccount, err := vestingtypes.NewContinuousVestingAccount(
		baseAcc,
		originalVesting,
		startTime.Unix(),
		endTime.Unix(),
	)
	require.NoError(t, err)
	accountKeeper.SetAccount(ctx, continuousVestingAccount)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk1 := pubKeys[0]

	// Create Validators and Delegation
	val1 := testutil.NewValidator(t, addrVal1, pk1)
	val1.Status = stakingtypes.Bonded
	require.NoError(t, stakingKeeper.SetValidator(ctx, val1))
	require.NoError(t, liquidKeeper.SetLiquidValidator(ctx, liquidtypes.NewLiquidValidator(val1.OperatorAddress)))
	require.NoError(t, stakingKeeper.SetValidatorByPowerIndex(ctx, val1))
	err = stakingKeeper.SetValidatorByConsAddr(ctx, val1)
	require.NoError(t, err)

	// Delegate all the vesting coins
	originalVestingAmount := originalVesting.AmountOf(sdk.DefaultBondDenom)
	err = delegateCoinsFromAccount(ctx, *stakingKeeper, addrAcc1, originalVestingAmount, val1)
	require.NoError(t, err)

	// Apply TM updates
	applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)

	_, err = stakingKeeper.GetDelegation(ctx, addrAcc1, addrVal1)
	require.NoError(t, err)

	// Check vesting account data
	// V=100, V'=0, DV=100, DF=0
	acc := accountKeeper.GetAccount(ctx, addrAcc1).(*vestingtypes.ContinuousVestingAccount)
	require.Equal(t, originalVesting, acc.GetVestingCoins(ctx.BlockTime()))
	require.Empty(t, acc.GetVestedCoins(ctx.BlockTime()))
	require.Equal(t, originalVesting, acc.GetDelegatedVesting())
	require.Empty(t, acc.GetDelegatedFree())

	msgServer := liquidkeeper.NewMsgServerImpl(liquidKeeper)

	// Vest half the original vesting coins
	vestHalfTime := startTime.Add(time.Duration(float64(endTime.Sub(startTime).Nanoseconds()) / float64(2)))
	ctx = ctx.WithBlockTime(vestHalfTime)

	// expect that half of the orignal vesting coins are vested
	expVestedCoins := originalVesting.QuoInt(math.NewInt(2))

	// Check vesting account data
	// V=50, V'=50, DV=100, DF=0
	acc = accountKeeper.GetAccount(ctx, addrAcc1).(*vestingtypes.ContinuousVestingAccount)
	require.Equal(t, expVestedCoins, acc.GetVestingCoins(ctx.BlockTime()))
	require.Equal(t, expVestedCoins, acc.GetVestedCoins(ctx.BlockTime()))
	require.Equal(t, originalVesting, acc.GetDelegatedVesting())
	require.Empty(t, acc.GetDelegatedFree())

	// Expect that tokenizing all the delegated coins fails
	// since only the half are vested
	_, err = msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    addrAcc1.String(),
		ValidatorAddress:    addrVal1.String(),
		Amount:              originalVesting[0],
		TokenizedShareOwner: addrAcc1.String(),
	})
	require.Error(t, err)

	// Tokenize the delegated vested coins
	_, err = msgServer.TokenizeShares(ctx, &liquidtypes.MsgTokenizeShares{
		DelegatorAddress:    addrAcc1.String(),
		ValidatorAddress:    addrVal1.String(),
		Amount:              sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: originalVestingAmount.Quo(math.NewInt(2))},
		TokenizedShareOwner: addrAcc1.String(),
	})
	require.NoError(t, err)

	shareDenom := addrVal1.String() + "/1"

	// Redeem the tokens
	_, err = msgServer.RedeemTokensForShares(ctx,
		&liquidtypes.MsgRedeemTokensForShares{
			DelegatorAddress: addrAcc1.String(),
			Amount:           sdk.Coin{Denom: shareDenom, Amount: originalVestingAmount.Quo(math.NewInt(2))},
		},
	)
	require.NoError(t, err)

	// After the redemption of the tokens, the vesting delegations should be evenly distributed
	// V=50, V'=50, DV=100, DF=50
	acc = accountKeeper.GetAccount(ctx, addrAcc1).(*vestingtypes.ContinuousVestingAccount)
	require.Equal(t, expVestedCoins, acc.GetVestingCoins(ctx.BlockTime()))
	require.Equal(t, expVestedCoins, acc.GetVestedCoins(ctx.BlockTime()))
	require.Equal(t, expVestedCoins, acc.GetDelegatedVesting())
	require.Equal(t, expVestedCoins, acc.GetDelegatedFree())
}
