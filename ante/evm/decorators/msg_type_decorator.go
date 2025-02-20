package decorators

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/gaia/v23/evm"
)

type MsgTypeDecorator struct{}

func NewMsgTypeDecorator() *MsgTypeDecorator {
	return &MsgTypeDecorator{}
}

func (a MsgTypeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()

	// only allow a single message of type MessageEthereumTx
	if len(msgs) != 1 {
		return ctx, fmt.Errorf("expected 1 msg, got %d", len(msgs))
	}

	for _, msg := range msgs {
		if _, ok := msg.(*evmtypes.MsgEthereumTx); !ok {
			return ctx, nil
		}
	}

	return next(ctx, tx, simulate)
}
