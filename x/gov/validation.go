package gov

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gaiaerrors "github.com/cosmos/gaia/v27/types/errors"
)

var (
	minStakedTokens       = math.LegacyNewDec(1000000) // 1_000_000 uatom (or 1 atom)
	maxDelegationsChecked = 100                        // number of delegation to check for the minStakedTokens
)

// SetMinStakedTokens sets the minimum amount of staked tokens required to vote
// Should only be used in testing
func SetMinStakedTokens(tokens math.LegacyDec) {
	minStakedTokens = tokens
}

// ValidateVoterStake checks if an address has sufficient stake to vote.
// This function validates the governance rule that voters must have a minimum
// amount of staked tokens. It is used by both the MsgServer (for all vote messages)
// and ante handlers (for transaction pre-validation).
func ValidateVoterStake(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, voter sdk.AccAddress) error {
	if minStakedTokens.IsZero() {
		return nil
	}

	enoughStake := false
	delegationCount := 0
	stakedTokens := math.LegacyNewDec(0)
	err := stakingKeeper.IterateDelegatorDelegations(ctx, voter, func(delegation stakingtypes.Delegation) bool {
		validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, err := stakingKeeper.GetValidator(ctx, validatorAddr)
		if err == nil {
			shares := delegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			stakedTokens = stakedTokens.Add(tokens)
			if stakedTokens.GTE(minStakedTokens) {
				enoughStake = true
				return true // break the iteration
			}
		}
		delegationCount++
		// break the iteration if maxDelegationsChecked were already checked
		return delegationCount >= maxDelegationsChecked
	})
	if err != nil {
		return err
	}

	if !enoughStake {
		return errorsmod.Wrapf(gaiaerrors.ErrInsufficientStake, "insufficient stake for voting - min required %v", minStakedTokens)
	}

	return nil
}
