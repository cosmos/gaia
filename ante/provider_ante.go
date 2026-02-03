package ante

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	gaiaerrors "github.com/cosmos/gaia/v26/types/errors"
)

const maxWrappedMessageDepthProvider = 4

// ProviderDecorator prevents MsgCreateConsumer messages from being processed
type ProviderDecorator struct {
	cdc codec.BinaryCodec
}

func NewProviderDecorator(cdc codec.BinaryCodec) ProviderDecorator {
	return ProviderDecorator{
		cdc: cdc,
	}
}

func (p ProviderDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, m := range tx.GetMsgs() {
		if err := p.validateMsgRecursive(ctx, m, 0); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (p ProviderDecorator) validateMsgRecursive(ctx sdk.Context, m sdk.Msg, iters int) error {
	if iters >= maxWrappedMessageDepthProvider {
		return errorsmod.Wrap(gaiaerrors.ErrNestedMessageLimitExceeded, "too many wrapped sdk messages")
	}

	// Handle authz wrapped messages
	if msg, ok := m.(*authz.MsgExec); ok {
		for _, v := range msg.Msgs {
			var innerMsg sdk.Msg
			if err := p.cdc.UnpackAny(v, &innerMsg); err != nil {
				return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "cannot unmarshal authz exec msgs")
			}
			if err := p.validateMsgRecursive(ctx, innerMsg, iters+1); err != nil {
				return err
			}
		}
		return nil
	}

	return p.validMsg(m)
}

func (p ProviderDecorator) validMsg(m sdk.Msg) error {
	// Block MsgCreateConsumer
	if _, ok := m.(*providertypes.MsgCreateConsumer); ok {
		return errorsmod.Wrap(gaiaerrors.ErrUnauthorized, "MsgCreateConsumer is disabled")
	}

	return nil
}
