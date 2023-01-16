package lockup

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/althea-net/althea-chain/x/lockup/keeper"
	"github.com/althea-net/althea-chain/x/lockup/types"
)

// WrappedAnteHandle An AnteDecorator used to wrap any AnteHandler for decorator chaining
type WrappedAnteHandler struct {
	anteHandler sdk.AnteHandler
}

// AnteHandle calls wad.anteHandler and then the next one in the chain
// This is necessary to use the Cosmos SDK's NewAnteHandler() output with a LockupAnteHandler
func (wad WrappedAnteHandler) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx, simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	modCtx, ok := wad.anteHandler(ctx, tx, simulate)
	if ok != nil {
		return modCtx, err
	}
	return next(modCtx, tx, simulate)
}

// WrappedLockupAnteHandler wraps a LockupAnteHandler around the input AnteHandler
func NewWrappedLockupAnteHandler(
	anteHandler sdk.AnteHandler,
	lockupKeeper keeper.Keeper,
	cdc codec.Codec,
) sdk.AnteHandler {
	wrapped := WrappedAnteHandler{anteHandler} // Must wrap to use in ChainAnteDecorators
	lad := NewLockupAnteDecorator(lockupKeeper, cdc)

	// Produces an AnteHandler which runs wrapped, then lad
	// Note: this is important as the default SetUpContextDecorator must be the
	// outermost one (see cosmos-sdk/x/auth/ante.NewAnteHandler())
	return sdk.ChainAnteDecorators(wrapped, lad)
}

// LockAnteDecorator Ensures that any transaction under a locked chain originates from a LockExempt address
type LockAnteDecorator struct {
	lockupKeeper keeper.Keeper
	cdc          codec.Codec
}

// AnteHandle Ensures that any transaction containing locked message types under a locked chain originates from a
// LockExempt address
func (lad LockAnteDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	if lad.lockupKeeper.GetChainLocked(ctx) {
		for _, msg := range tx.GetMsgs() {
			switch msg := msg.(type) {
			case *authz.MsgExec:
				// Check that all of the inner Msgs are not locked types
				for _, m := range msg.Msgs {
					var inner sdk.Msg
					err := lad.cdc.UnpackAny(m, &inner)
					if err != nil {
						return ctx,
							sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "unable to unpack authz msgexec message: %v", err)
					}
					if err := lad.isAcceptable(ctx, inner); err != nil {
						return ctx, err
					}
				}
			default:
				// Check that this msg is not one of the locked types
				if err := lad.isAcceptable(ctx, msg); err != nil {
					return ctx, err
				}
			}

		}
	}

	return next(ctx, tx, simulate)
}

// isAcceptable checks if the given msg is acceptable, returns an error if not
func (lad LockAnteDecorator) isAcceptable(ctx sdk.Context, msg sdk.Msg) error {
	lockedMsgTypesSet := lad.lockupKeeper.GetLockedMessageTypesSet(ctx)
	exemptSet := lad.lockupKeeper.GetLockExemptAddressesSet(ctx)
	if _, typePresent := lockedMsgTypesSet[sdk.MsgTypeURL(msg)]; typePresent {
		if allow, err := allowMessage(msg, exemptSet); !allow {
			return sdkerrors.Wrap(err, fmt.Sprintf("Transaction blocked because of message %v", msg))
		}
	}

	return nil
}

// NewAnteHandler returns an AnteHandler that ensures any transaction under a locked chain
// originates from a LockExempt address
func NewLockupAnteHandler(lockupKeeper keeper.Keeper, cdc codec.Codec) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(NewLockupAnteDecorator(lockupKeeper, cdc))
}

// NewLockupAnteDecorator initializes a LockupAnteDecorator for locking messages
// based on the settings stored in lockupKeeper
func NewLockupAnteDecorator(lockupKeeper keeper.Keeper, cdc codec.Codec) LockAnteDecorator {
	return LockAnteDecorator{lockupKeeper, cdc}
}

// allowMessage checks that an input `msg` was sent by only addresses in `exemptSet`
// returns true if `msg` is either permissible or not a type of message this module blocks
func allowMessage(msg sdk.Msg, exemptSet map[string]struct{}) (bool, error) {
	switch sdk.MsgTypeURL(msg) {
	case sdk.MsgTypeURL(&banktypes.MsgSend{}):
		msgSend := msg.(*banktypes.MsgSend)
		if _, present := exemptSet[msgSend.FromAddress]; !present {
			// Message sent from a non-exempt address while the chain is locked up, returning error
			return false, sdkerrors.Wrap(types.ErrLocked,
				"The chain is locked, only exempt addresses may be the FromAddress in a Send message")
		}
		return true, nil
	case sdk.MsgTypeURL(&banktypes.MsgMultiSend{}):
		msgMultiSend := msg.(*banktypes.MsgMultiSend)
		for _, input := range msgMultiSend.Inputs {
			if _, present := exemptSet[input.Address]; !present {
				// Multi-send Message sent with a non-exempt input address while the chain is locked up, returning error
				return false, sdkerrors.Wrap(types.ErrLocked,
					"The chain is locked, only exempt addresses may be inputs in a MultiSend message")
			}
		}
		return true, nil
	default:
		return false, sdkerrors.Wrap(types.ErrUnhandled,
			fmt.Sprintf("Message type %v does not have a case in allowMessage, unable to handle messages like this",
				sdk.MsgTypeURL(msg),
			),
		)
	}
}
