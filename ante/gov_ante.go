package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// initial deposit must be greater than or equal to 10% of the minimum deposit
var minInitialDepositFraction = sdk.NewDecWithPrec(10, 2)

type GovPreventSpamDecorator struct {
	govKeeper *govkeeper.Keeper
	cdc       codec.BinaryCodec
}

func NewGovPreventSpamDecorator(cdc codec.BinaryCodec, govKeeper *govkeeper.Keeper) GovPreventSpamDecorator {
	return GovPreventSpamDecorator{
		govKeeper: govKeeper,
		cdc:       cdc,
	}
}

func (g GovPreventSpamDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// run checks only on CheckTx or simulate
	if !ctx.IsCheckTx() || simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if err = g.ValidateGovMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// validateGovMsgs checks if the InitialDeposit amounts are greater than the minimum initial deposit amount
func (g GovPreventSpamDecorator) ValidateGovMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	validMsg := func(m sdk.Msg) error {
		if msg, ok := m.(*govtypes.MsgSubmitProposal); ok {
			// prevent messages with insufficient initial deposit amount
			depositParams := g.govKeeper.GetDepositParams(ctx)
			minInitialDeposit := g.calcMinInitialDeposit(depositParams.MinDeposit)
			if msg.InitialDeposit.IsAllLT(minInitialDeposit) {
				return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient initial deposit amount - required: %v", minInitialDeposit)
			}
		}

		return nil
	}

	validAuthz := func(execMsg *authz.MsgExec) error {
		for _, v := range execMsg.Msgs {
			var innerMsg sdk.Msg
			if err := g.cdc.UnpackAny(v, &innerMsg); err != nil {
				return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
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

func (g GovPreventSpamDecorator) calcMinInitialDeposit(minDeposit sdk.Coins) (minInitialDeposit sdk.Coins) {
	for _, coin := range minDeposit {
		minInitialCoins := minInitialDepositFraction.MulInt(coin.Amount).RoundInt()
		minInitialDeposit = minInitialDeposit.Add(sdk.NewCoin(coin.Denom, minInitialCoins))
	}
	return
}
