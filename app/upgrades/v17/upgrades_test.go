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

	// create two dummy validators addresses
	delAddr1 := sdk.AccAddress([]byte("delAddr1"))
	delAddr2 := sdk.AccAddress([]byte("delAddr2"))

	// create 8 dummy validators addresses to use
	// in dummy redelegations
	valSrcAddr1 := sdk.ValAddress([]byte("ValSrcAddr1"))
	valSrcAddr2 := sdk.ValAddress([]byte("ValSrcAddr2"))
	valSrcAddr3 := sdk.ValAddress([]byte("ValSrcAddr3"))
	valSrcAddr4 := sdk.ValAddress([]byte("ValSrcAddr4"))

	valDstAddr1 := sdk.ValAddress([]byte("ValDstAddr1"))
	valDstAddr2 := sdk.ValAddress([]byte("ValDstAddr2"))
	valDstAddr3 := sdk.ValAddress([]byte("ValDstAddr3"))
	valDstAddr4 := sdk.ValAddress([]byte("ValDstAddr4"))

	timeNow := time.Now()

	// define 1 delegation for delegator 1 (delAddr1)
	del1Delegation := stakingtypes.Delegation{
		DelegatorAddress: delAddr1.String(),
		ValidatorAddress: valDstAddr3.String(),
		Shares:           sdk.NewDec(5),
	}

	// define 2 delegations for delegator 2 (delAddr2)
	del2Delegations := []stakingtypes.Delegation{
		{
			DelegatorAddress: delAddr2.String(),
			ValidatorAddress: valDstAddr2.String(),
			Shares:           sdk.NewDec(50),
		},
		{
			DelegatorAddress: delAddr2.String(),
			ValidatorAddress: valDstAddr3.String(),
			Shares:           sdk.NewDec(40),
		},
	}

	// define 4 redelegations from delegator 1
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

	// define 1 unbonding delegation for delegator 1
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

	// define 3 redelegations for delegator 2
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

	// define 2 unbonding delegations for delegator 2
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
		},
		{
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

	// set delegations
	stakingKeeper.SetDelegation(ctx, del1Delegation)
	stakingKeeper.SetDelegation(ctx, del2Delegations[0])
	stakingKeeper.SetDelegation(ctx, del2Delegations[1])

	// set redelegations
	allReds := append(del1Reds, del2Reds...)
	for _, red := range allReds {
		stakingKeeper.SetRedelegation(ctx, red)
	}

	// set unbonding delegations
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
		delegation shares:
					valDst1		valDst2		valDst3 	valDst4
			del1	      0	   		  0	   		  5	          0
			del2	   	  0	  		 50	  		 40	          0


		shares redelegated to validators:
					valDst1		valDst2		valDst3 	valDst4
			del1	      0	      	  0	  		 17	  		 20
			del2	      3	    	100	  		 22	   		  0

		unbonding delegation shares:
					valDst1		valDst2		valDst3 	valDst4
			del1	      0	          0	  		 10	   		  0
			del2	      5	         16	  		 22	   		  0


		calculations to determine the remaining redelegated shares
		after unbonding delegations completes. Each redelegation (+) and unbonding delegation (-) entries
		are sorted by completion time in ascending order.

		(del1, val3)
		+5 -5 +5 -4 +2 -1 +5 => 7

		(del1, val4)
		+20 => 20

		(del2, val1)
		+1 -5 +2 => 2

		(del2, val2)
		+100 -10 -5 => 85

		(del2, val3)
		+22 => 22

		expected remaining shares after migrations:
		(del1, val3)
		=> 5

		(del1, val4)
		=> 0

		(del2, val1)
		=> 0

		(del2, val2)
		=> 50

		(del2, val3)
		=> 22

	*/

	// assert that all redelegations are persisted before the upgrade
	_, found := stakingKeeper.GetRedelegation(ctx, delAddr1, valSrcAddr3, valDstAddr3)
	require.True(t, found)
	_, found = stakingKeeper.GetRedelegation(ctx, delAddr1, valSrcAddr4, valDstAddr4)
	require.True(t, found)
	_, found = stakingKeeper.GetRedelegation(ctx, delAddr2, valSrcAddr1, valDstAddr1)
	require.True(t, found)
	_, found = stakingKeeper.GetRedelegation(ctx, delAddr2, valSrcAddr2, valDstAddr2)
	require.True(t, found)
	_, found = stakingKeeper.GetRedelegation(ctx, delAddr2, valSrcAddr3, valDstAddr3)
	require.True(t, found)

	// case 1 de exists so all redelegations should be deleted
	err := v17.MigrateRedelegations(ctx, *stakingKeeper)
	require.NoError(t, err)

	// redelegation to valDst4 from delegator 1 should have been deleted
	_, found = stakingKeeper.GetRedelegation(ctx, delAddr1, valSrcAddr4, valDstAddr4)
	require.False(t, found)

	// check remaining redelegation shares for delegator 1 after unbonding delegations
	del1Reds = stakingKeeper.GetRedelegations(ctx, delAddr1, uint16(10000))
	// note that delegator 1 have 3 redelegations from different source to valDst3
	require.Len(t, del1Reds, 3)

	resDel1Ubd, found := stakingKeeper.GetUnbondingDelegation(ctx, delAddr1, valDstAddr3)
	require.True(t, found)
	resDel1Reds, err := v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr1.String(),
		valDstAddr3,
		resDel1Ubd,
		del1Reds,
	)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(5), resDel1Reds)

	// check remaining redelegation shares for delegator 2 after unbonding delegations
	resDel2Ubd, found := stakingKeeper.GetUnbondingDelegation(ctx, delAddr2, valDstAddr1)
	require.True(t, found)
	del2RedValDst1, found := stakingKeeper.GetRedelegation(ctx, delAddr2, valSrcAddr1, valDstAddr1)
	require.True(t, found)

	resDel2Reds, err := v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr2.String(),
		valDstAddr1,
		resDel2Ubd,
		[]stakingtypes.Redelegation{del2RedValDst1},
	)
	require.NoError(t, err)
	require.Equal(t, sdk.ZeroDec(), resDel2Reds)

	resDel2Ubd2, found := stakingKeeper.GetUnbondingDelegation(ctx, delAddr2, valDstAddr2)
	require.True(t, found)
	del2RedValDst2, found := stakingKeeper.GetRedelegation(ctx, delAddr2, valSrcAddr2, valDstAddr2)
	require.True(t, found)

	resDel2Reds, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr2.String(),
		valDstAddr2,
		resDel2Ubd2,
		[]stakingtypes.Redelegation{del2RedValDst2},
	)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(50), resDel2Reds)

	del2RedValDst3, found := stakingKeeper.GetRedelegation(ctx, delAddr2, valSrcAddr3, valDstAddr3)
	require.True(t, found)

	resDel2Reds, err = v17.ComputeRemainingRedelegatedSharesAfterUnbondings(
		*stakingKeeper,
		ctx,
		delAddr2.String(),
		valDstAddr3,
		stakingtypes.UnbondingDelegation{},
		[]stakingtypes.Redelegation{del2RedValDst3},
	)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(22), resDel2Reds)
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
