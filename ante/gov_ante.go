package ante

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var MiniumInitialDepositRate = sdk.NewDecWithPrec(10, 2)

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

func (gpsd GovPreventSpamDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	if err = gpsd.checkSpamSubmitProposalMsg(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (gpsd GovPreventSpamDecorator) checkSpamSubmitProposalMsg(ctx sdk.Context, msgs []sdk.Msg) error {
	validMsg := func(m sdk.Msg) error {
		if msg, ok := m.(*govtypes.MsgSubmitProposal); ok {
			// prevent spam gov msg
			depositParams := gpsd.govKeeper.GetDepositParams(ctx)
			miniumInitialDeposit := gpsd.calcMiniumInitialDeposit(depositParams.MinDeposit)
			if msg.InitialDeposit.IsAllLT(miniumInitialDeposit) {
				return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "not enough initial deposit. required: %v", miniumInitialDeposit)
			}
		}

		return nil
	}

	validAuthz := func(execMsg *authz.MsgExec) error {
		for _, v := range execMsg.Msgs {
			var innerMsg sdk.Msg
			if err := gpsd.cdc.UnpackAny(v, &innerMsg); err != nil {
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

func (gpsd GovPreventSpamDecorator) calcMiniumInitialDeposit(minDeposit sdk.Coins) (miniumInitialDeposit sdk.Coins) {
	for _, coin := range minDeposit {
		miniumInitialCoin := MiniumInitialDepositRate.MulInt(coin.Amount).RoundInt()
		miniumInitialDeposit = miniumInitialDeposit.Add(sdk.NewCoin(coin.Denom, miniumInitialCoin))
	}

	return
}
