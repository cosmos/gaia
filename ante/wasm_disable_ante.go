package ante

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gaiaerrors "github.com/cosmos/gaia/v28/types/errors"
)

// WasmDisableDecorator rejects MsgStoreCode, MsgInstantiateContract,
// MsgInstantiateContract2, MsgStoreAndInstantiateContract, and
// MsgStoreAndMigrateContract messages, including when wrapped (at any
// nesting depth) inside authz.MsgExec.
//
// This ante-layer check only sees messages that are top-level (or
// authz.MsgExec-wrapped) in a directly broadcast tx, so it's a cheap
// fail-fast rejection for the common case (rejects at CheckTx, before any
// gas is spent) but not airtight — e.g. Interchain Accounts dispatch
// packet-borne messages via baseapp.MsgServiceRouter directly, bypassing
// the ante chain entirely. x/wasm's msgServer (x/wasm/msg_server.go) blocks
// the same five messages at the MsgServer level instead, which baseapp
// always routes through regardless of dispatch path, and is the
// authoritative backstop.
type WasmDisableDecorator struct {
	cdc codec.BinaryCodec
}

func NewWasmDisableDecorator(cdc codec.BinaryCodec) WasmDisableDecorator {
	return WasmDisableDecorator{
		cdc: cdc,
	}
}

func (w WasmDisableDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// do not run check during simulations
	if simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if err = w.ValidateMsgs(msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ValidateMsgs rejects the tx if it contains a blocked wasm message at any
// nesting depth.
func (w WasmDisableDecorator) ValidateMsgs(msgs []sdk.Msg) error {
	for _, msg := range msgs {
		if err := w.validateMsgRecursive(msg, 0); err != nil {
			return err
		}
	}
	return nil
}

func (w WasmDisableDecorator) validateMsgRecursive(m sdk.Msg, iters int) error {
	if iters >= maxWrappedMessageDepth {
		return errorsmod.Wrap(gaiaerrors.ErrNestedMessageLimitExceeded, "too many wrapped sdk messages")
	}
	if msg, ok := m.(*authz.MsgExec); ok {
		for _, v := range msg.Msgs {
			var innerMsg sdk.Msg
			if err := w.cdc.UnpackAny(v, &innerMsg); err != nil {
				return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}
			if err := w.validateMsgRecursive(innerMsg, iters+1); err != nil {
				return err
			}
		}
		return nil
	}
	return w.validMsg(m)
}

func (w WasmDisableDecorator) validMsg(m sdk.Msg) error {
	switch m.(type) {
	case *wasmtypes.MsgStoreCode:
		return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgStoreCode is temporarily disabled")
	case *wasmtypes.MsgInstantiateContract:
		return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgInstantiateContract is temporarily disabled")
	case *wasmtypes.MsgInstantiateContract2:
		return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgInstantiateContract2 is temporarily disabled")
	case *wasmtypes.MsgStoreAndInstantiateContract:
		return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgStoreAndInstantiateContract is temporarily disabled")
	case *wasmtypes.MsgStoreAndMigrateContract:
		return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgStoreAndMigrateContract is temporarily disabled")
	default:
		// not a blocked message - nothing to validate
		return nil
	}
}
