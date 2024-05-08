package v17_test

import (
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v17 "github.com/cosmos/gaia/v17/app/upgrades/v17"

	"github.com/cosmos/gaia/v17/app/helpers"
)

func TestUpgradeRelegations(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	delAddr1 := sdk.AccAddress([]byte("delAddr1"))
	delAddr2 := sdk.AccAddress([]byte("delAddr2"))

	valSrcAddr1 := sdk.ValAddress([]byte("ValSrcAddr1"))
	valSrcAddr2 := sdk.ValAddress([]byte("ValSrcAddr2"))
	valSrcAddr3 := sdk.ValAddress([]byte("ValSrcAddr3"))
	valSrcAddr4 := sdk.ValAddress([]byte("ValSrcAddr4"))

	valDstAddr1 := sdk.ValAddress([]byte("ValDstAddr1"))
	valDstAddr2 := sdk.ValAddress([]byte("ValDstAddr2"))
	valDstAddr3 := sdk.ValAddress([]byte("ValDstAddr3"))
	valDstAddr4 := sdk.ValAddress([]byte("ValDstAddr4"))

	timeNow := time.Now()

	// define redelegations for a first delegator
	del1Reds := []stakingtypes.Redelegation{
		{
			DelegatorAddress:    delAddr1.String(),
			ValidatorSrcAddress: valSrcAddr1.String(),
			ValidatorDstAddress: valDstAddr3.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow,
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					CompletionTime: timeNow.Add(5 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr1.String(),
			ValidatorSrcAddress: valSrcAddr2.String(),
			ValidatorDstAddress: valDstAddr3.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(10 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr1.String(),
			ValidatorSrcAddress: valSrcAddr3.String(),
			ValidatorDstAddress: valDstAddr3.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(15 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr1.String(),
			ValidatorSrcAddress: valSrcAddr4.String(),
			ValidatorDstAddress: valDstAddr4.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(40 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(10)),
				},
				{
					CompletionTime: timeNow.Add(50 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(10)),
				},
			},
		},
	}

	// define redelegations for a second validator
	del2Reds := []stakingtypes.Redelegation{
		{
			DelegatorAddress:    delAddr2.String(),
			ValidatorSrcAddress: valSrcAddr1.String(),
			ValidatorDstAddress: valDstAddr1.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow,
					SharesDst:      sdk.NewDecFromInt(math.NewInt(1)),
				},
				{
					CompletionTime: timeNow.Add(5 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr2.String(),
			ValidatorSrcAddress: valSrcAddr2.String(),
			ValidatorDstAddress: valDstAddr2.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(10 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(100)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr2.String(),
			ValidatorSrcAddress: valSrcAddr3.String(),
			ValidatorDstAddress: valDstAddr3.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(15 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(20)),
				},
				{
					CompletionTime: timeNow.Add(20 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(1)),
				},
				{
					CompletionTime: timeNow.Add(30 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(1)),
				},
			},
		},
	}

	// define unbonding delegation for second validator
	del1Ubd := stakingtypes.UnbondingDelegation{
		DelegatorAddress: delAddr1.String(),
		ValidatorAddress: valDstAddr3.String(),
		Entries: []stakingtypes.UnbondingDelegationEntry{
			{
				CompletionTime: timeNow.Add(3 * time.Hour),
				InitialBalance: math.NewInt(5), // 5 - 5 => 0
			},
			{
				CompletionTime: timeNow.Add(8 * time.Hour),
				InitialBalance: math.NewInt(4), // 5 - 4 => 1
			},
			{
				CompletionTime: timeNow.Add(12 * time.Hour),
				InitialBalance: math.NewInt(1), // 1 + 2 - 1 => 2
			},
		},
	}

	// define unbonding delegation for first validator
	del2Ubd := []stakingtypes.UnbondingDelegation{
		{
			DelegatorAddress: delAddr2.String(),
			ValidatorAddress: valDstAddr1.String(),
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CompletionTime: timeNow.Add(3 * time.Hour),
					InitialBalance: math.NewInt(5), // 1 - 5 => 0
				},
			},
		}, {
			DelegatorAddress: delAddr2.String(),
			ValidatorAddress: valDstAddr2.String(),
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CompletionTime: timeNow.Add(8 * time.Hour),
					InitialBalance: math.NewInt(1), // 0 - 1 => 0
				},
				{
					CompletionTime: timeNow.Add(12 * time.Hour),
					InitialBalance: math.NewInt(10), // 100 - 10 => 90
				},
				{
					CompletionTime: timeNow.Add(25 * time.Hour),
					InitialBalance: math.NewInt(5), // 90 - 5 => 85
				},
			},
		},
	}

	allReds := append(del1Reds, del2Reds...)

	for _, red := range allReds {
		stakingKeeper.SetRedelegation(ctx, red)
	}

	// set ubds
	stakingKeeper.SetUnbondingDelegation(ctx, del1Ubd)
	stakingKeeper.SetUnbondingDelegation(ctx, del2Ubd[0])
	stakingKeeper.SetUnbondingDelegation(ctx, del2Ubd[1])

	// set delegations
	stakingKeeper.SetDelegation(ctx, stakingtypes.Delegation{
		DelegatorAddress: delAddr1.String(),
		ValidatorAddress: valDstAddr3.String(),
		Shares:           sdk.NewDec(5),
	})
	stakingKeeper.SetDelegation(ctx, stakingtypes.Delegation{
		DelegatorAddress: delAddr2.String(),
		ValidatorAddress: valDstAddr2.String(),
		Shares:           sdk.NewDec(50),
	})
	stakingKeeper.SetDelegation(ctx, stakingtypes.Delegation{
		DelegatorAddress: delAddr2.String(),
		ValidatorAddress: valDstAddr3.String(),
		Shares:           sdk.NewDec(40),
	})

	stakingKeeper.SetUnbondingDelegation(ctx, del1Ubd)
	stakingKeeper.SetUnbondingDelegation(ctx, del2Ubd[0])
	stakingKeeper.SetUnbondingDelegation(ctx, del2Ubd[1])

	// set validators
	for i := 0; i < 4; i++ {
		stakingKeeper.SetValidator(ctx, stakingtypes.Validator{
			OperatorAddress: sdk.ValAddress([]byte("ValDstAddr" + strconv.Itoa(i+1))).String(),
			DelegatorShares: sdk.NewDecFromInt(sdk.NewInt(100)),
			LiquidShares:    sdk.NewDecFromInt(sdk.NewInt(100)),
			Tokens:          sdk.NewInt(100),
		})
	}

	/*
		delegators redelegated shares
		after unbonding delegation completions:

		 									val1 	val2 	val3 	val4
			del1 redelegation remaining		 0       0       7      20
			del2 redelegation remaining		 0 		  85	 22 	  0

											val1 	val2 	val3 	val4
			del1 delegation shares			 0		  0 	  5       0
			del2 delegation shares			 0		  50	  40      0

			expected redelegations after migrations:

											val1 	val2 	val3 	val4
			del1 delegation shares			 0		  0 	  2       0
			del2 delegation shares			 0		  50	  22      0

	*/

	// case 1 de exists so all redelegations should be deleted
	err := v17.MigrateRedelegations(ctx, *stakingKeeper)

	resDel1Reds := stakingKeeper.GetRedelegations(ctx, delAddr1, uint16(10000))
	require.Len(t, resDel1Reds, 1)
	require.Len(t, resDel1Reds[0].Entries, 1)
	require.Equal(t, del1Reds[0].ValidatorDstAddress, resDel1Reds[0].ValidatorDstAddress)
	require.Equal(t, sdk.NewDec(5), del1Reds[0].ValidatorDstAddress, resDel1Reds[0].Entries[0].SharesDst)

	require.Equal(t, sdk.NewDec(5), stakingKeeper.GetRedelegations(ctx, delAddr1, uint16(10000)))
	require.Empty(t, stakingKeeper.GetRedelegations(ctx, delAddr2, uint16(10000)))

	require.NoError(t, err)
}

func TestComputeRemainingRedelegatedSharesAfterUnbondings(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	delAddr := sdk.AccAddress([]byte("delAddr"))
	validatorDstAddress := sdk.ValAddress([]byte("ValDstAddr"))
	timeNow := time.Now()

	reds := []stakingtypes.Redelegation{
		{
			DelegatorAddress:    delAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow,
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					CompletionTime: timeNow.Add(5 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr1")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(10 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr0")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []stakingtypes.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(15 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					CompletionTime: timeNow.Add(20 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(1)),
				},
				{
					CompletionTime: timeNow.Add(30 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
	}

	ubd := stakingtypes.UnbondingDelegation{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: validatorDstAddress.String(),
		Entries: []stakingtypes.UnbondingDelegationEntry{
			{
				CompletionTime: timeNow.Add(3 * time.Hour),
				InitialBalance: math.NewInt(5), // 5 - 5 => 0
			},
			{
				CompletionTime: timeNow.Add(8 * time.Hour),
				InitialBalance: math.NewInt(1), // 5 - 1 => 4
			},
			{
				CompletionTime: timeNow.Add(12 * time.Hour),
				InitialBalance: math.NewInt(10), // 4 + 2 - 10 => 0
			},
			{
				CompletionTime: timeNow.Add(25 * time.Hour),
				InitialBalance: math.NewInt(5), // 5 + 1 - 5 + 2 => 3
			},
		},
	}

	// expect an error when validator isn't set
	_, err := v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr.String(),
		validatorDstAddress,
		ubd,
		reds,
	)
	require.Error(t, err)

	// set validator
	stakingKeeper.SetValidator(ctx, stakingtypes.Validator{
		OperatorAddress: validatorDstAddress.String(),
		DelegatorShares: sdk.NewDecFromInt(sdk.NewInt(100)),
		Tokens:          sdk.NewInt(100),
	})

	// expect an error when the passed delegator address doesn't match the one in the redelegations
	_, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		sdk.AccAddress([]byte("wrongDelAddr")).String(),
		validatorDstAddress,
		ubd,
		reds,
	)
	require.Error(t, err)

	// expect an error when the passed validator address doesn't match the one in the redelegations
	_, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr.String(),
		sdk.ValAddress([]byte("wrongValDstAddr")),
		ubd,
		reds,
	)
	require.Error(t, err)

	// expect no error when no redelegations is passed
	res, err := v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr.String(),
		validatorDstAddress,
		ubd,
		[]stakingtypes.Redelegation{},
	)
	require.NoError(t, err)
	require.Equal(t, sdk.ZeroDec(), res)

	// expect no error when no unbonding delegations exist
	res, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr.String(),
		validatorDstAddress,
		stakingtypes.UnbondingDelegation{},
		reds,
	)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDecFromInt(sdk.NewInt(20)), res)

	stakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// expect no error when no redelegations is passed
	res, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr.String(),
		validatorDstAddress,
		ubd,
		[]stakingtypes.Redelegation{},
	)
	require.NoError(t, err)
	require.Equal(t, sdk.ZeroDec(), res)

	// expect no error
	res, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr.String(),
		validatorDstAddress,
		ubd,
		reds,
	)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDecFromInt(sdk.NewInt(3)), res)
}
