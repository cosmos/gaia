package gov_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v27/app/helpers"
	gaiagov "github.com/cosmos/gaia/v27/x/gov"
)

// TestMsgServerVoteValidation tests that the decorated MsgServer
// properly validates voter stake for Vote messages.
func TestMsgServerVoteValidation(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	// Create the decorated MsgServer
	msgServer := gaiagov.NewMsgServerImpl(gaiaApp.GovKeeper, gaiaApp.StakingKeeper)

	// Get validator
	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	valAddr1, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	require.NoError(t, err)
	valAddr1 = sdk.ValAddress(valAddr1)

	// Get delegator (this account was created during setup)
	addr, err := gaiaApp.AccountKeeper.Accounts.Indexes.Number.MatchExact(ctx, 0)
	require.NoError(t, err)
	delegator, err := sdk.AccAddressFromBech32(addr.String())
	require.NoError(t, err)

	tests := []struct {
		name       string
		bondAmt    math.Int
		expectPass bool
	}{
		{
			name:       "vote with 0 atom - should fail",
			bondAmt:    math.ZeroInt(),
			expectPass: false,
		},
		{
			name:       "vote with 0.5 atom - should fail",
			bondAmt:    math.NewInt(500000),
			expectPass: false,
		},
		{
			name:       "vote with 1 atom - should pass validation (may fail on proposal)",
			bondAmt:    math.NewInt(1000000),
			expectPass: true, // Will pass stake validation but may fail on proposal not found
		},
		{
			name:       "vote with 10 atom - should pass validation (may fail on proposal)",
			bondAmt:    math.NewInt(10000000),
			expectPass: true, // Will pass stake validation but may fail on proposal not found
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Unbond all tokens for this delegator
			delegations, err := stakingKeeper.GetAllDelegatorDelegations(ctx, delegator)
			require.NoError(t, err)
			for _, del := range delegations {
				valAddr, err := sdk.ValAddressFromBech32(del.GetValidatorAddr())
				require.NoError(t, err)
				_, _, err = stakingKeeper.Undelegate(ctx, delegator, valAddr, del.GetShares())
				require.NoError(t, err)
			}

			// Delegate tokens
			if !tc.bondAmt.IsZero() {
				val, err := stakingKeeper.GetValidator(ctx, valAddr1)
				require.NoError(t, err)
				_, err = stakingKeeper.Delegate(ctx, delegator, tc.bondAmt, stakingtypes.Unbonded, val, true)
				require.NoError(t, err)
			}

			// Create vote message
			msg := govv1.NewMsgVote(
				delegator,
				1, // Non-existent proposal
				govv1.VoteOption_VOTE_OPTION_YES,
				"test-vote",
			)

			// Execute via decorated MsgServer
			_, err = msgServer.Vote(ctx, msg)

			if tc.expectPass {
				// For sufficient stake, the validation passes but may fail on "proposal not found"
				// We check that it's NOT the insufficient stake error
				if err != nil {
					require.NotContains(t, err.Error(), "insufficient stake")
				}
			} else {
				// For insufficient stake, should get the stake error
				require.Error(t, err)
				require.Contains(t, err.Error(), "insufficient stake")
			}
		})
	}
}

// TestMsgServerVoteWeightedValidation tests that the decorated MsgServer
// properly validates voter stake for VoteWeighted messages.
func TestMsgServerVoteWeightedValidation(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	// Create the decorated MsgServer
	msgServer := gaiagov.NewMsgServerImpl(gaiaApp.GovKeeper, gaiaApp.StakingKeeper)

	// Get validator
	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	valAddr1, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	require.NoError(t, err)
	valAddr1 = sdk.ValAddress(valAddr1)

	// Get delegator (this account was created during setup)
	addr, err := gaiaApp.AccountKeeper.Accounts.Indexes.Number.MatchExact(ctx, 0)
	require.NoError(t, err)
	delegator, err := sdk.AccAddressFromBech32(addr.String())
	require.NoError(t, err)

	tests := []struct {
		name       string
		bondAmt    math.Int
		expectPass bool
	}{
		{
			name:       "weighted vote with 0 atom - should fail",
			bondAmt:    math.ZeroInt(),
			expectPass: false,
		},
		{
			name:       "weighted vote with 0.5 atom - should fail",
			bondAmt:    math.NewInt(500000),
			expectPass: false,
		},
		{
			name:       "weighted vote with 1 atom - should pass validation",
			bondAmt:    math.NewInt(1000000),
			expectPass: true,
		},
		{
			name:       "weighted vote with 10 atom - should pass validation",
			bondAmt:    math.NewInt(10000000),
			expectPass: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Unbond all tokens for this delegator
			delegations, err := stakingKeeper.GetAllDelegatorDelegations(ctx, delegator)
			require.NoError(t, err)
			for _, del := range delegations {
				valAddr, err := sdk.ValAddressFromBech32(del.GetValidatorAddr())
				require.NoError(t, err)
				_, _, err = stakingKeeper.Undelegate(ctx, delegator, valAddr, del.GetShares())
				require.NoError(t, err)
			}

			// Delegate tokens
			if !tc.bondAmt.IsZero() {
				val, err := stakingKeeper.GetValidator(ctx, valAddr1)
				require.NoError(t, err)
				_, err = stakingKeeper.Delegate(ctx, delegator, tc.bondAmt, stakingtypes.Unbonded, val, true)
				require.NoError(t, err)
			}

			// Create weighted vote message
			msg := govv1.NewMsgVoteWeighted(
				delegator,
				1, // Non-existent proposal
				govv1.NewNonSplitVoteOption(govv1.VoteOption_VOTE_OPTION_YES),
				"test-weighted-vote",
			)

			// Execute via decorated MsgServer
			_, err = msgServer.VoteWeighted(ctx, msg)

			if tc.expectPass {
				// For sufficient stake, the validation passes but may fail on "proposal not found"
				// We check that it's NOT the insufficient stake error
				if err != nil {
					require.NotContains(t, err.Error(), "insufficient stake")
				}
			} else {
				// For insufficient stake, should get the stake error
				require.Error(t, err)
				require.Contains(t, err.Error(), "insufficient stake")
			}
		})
	}
}

// TestMsgServerPassthroughMethods tests that non-vote methods
// are properly passed through to the underlying MsgServer.
func TestMsgServerPassthroughMethods(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	// Create the decorated MsgServer
	msgServer := gaiagov.NewMsgServerImpl(gaiaApp.GovKeeper, gaiaApp.StakingKeeper)

	// Get a test address
	pk := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pk.Address())

	// Test Deposit - should pass through to underlying implementation
	// (will fail because proposal doesn't exist, but that proves passthrough works)
	depositMsg := &govv1.MsgDeposit{
		ProposalId: 1,
		Depositor:  addr.String(),
		Amount:     sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(1000000))),
	}
	_, err := msgServer.Deposit(ctx, depositMsg)
	// Should fail with proposal not found, not with any stake-related error
	require.Error(t, err)
	require.NotContains(t, err.Error(), "insufficient stake")
}
