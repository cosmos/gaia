package ante

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gaiaerrors "github.com/cosmos/gaia/v23/types/errors"
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

type GovVoteDecorator struct {
	stakingKeeper *stakingkeeper.Keeper
	cdc           codec.BinaryCodec
}

func NewGovVoteDecorator(cdc codec.BinaryCodec, stakingKeeper *stakingkeeper.Keeper) GovVoteDecorator {
	return GovVoteDecorator{
		stakingKeeper: stakingKeeper,
		cdc:           cdc,
	}
}

func (g GovVoteDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// do not run check during simulations
	if simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if err = g.ValidateVoteMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ValidateVoteMsgs checks if a voter has enough stake to vote
func (g GovVoteDecorator) ValidateVoteMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	validMsg := func(m sdk.Msg) error {
		var accAddr sdk.AccAddress
		var err error

		switch msg := m.(type) {
		case *govv1beta1.MsgVote:
			accAddr, err = sdk.AccAddressFromBech32(msg.Voter)
			if err != nil {
				return err
			}
		case *govv1.MsgVote:
			accAddr, err = sdk.AccAddressFromBech32(msg.Voter)
			if err != nil {
				return err
			}
		default:
			// not a vote message - nothing to validate
			return nil
		}

		if minStakedTokens.IsZero() {
			return nil
		}

		enoughStake := false
		delegationCount := 0
		stakedTokens := math.LegacyNewDec(0)
		err = g.stakingKeeper.IterateDelegatorDelegations(ctx, accAddr, func(delegation stakingtypes.Delegation) bool {
			validatorAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
			if err != nil {
				panic(err) // shouldn't happen
			}
			validator, err := g.stakingKeeper.GetValidator(ctx, validatorAddr)
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

	validAuthz := func(execMsg *authz.MsgExec) error {
		for _, v := range execMsg.Msgs {
			var innerMsg sdk.Msg
			if err := g.cdc.UnpackAny(v, &innerMsg); err != nil {
				return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}
			if err := validMsg(innerMsg); err != nil {
				return err
			}
		}

		return nil
	}

	for _, m := range msgs {
		if msg, ok := m.(*authz.MsgExec); ok {
			if err := validAuthz(msg); err != nil {
				return err
			}
			continue
		}

		// validate normal msgs
		if err := validMsg(m); err != nil {
			return err
		}
	}
	return nil
}
