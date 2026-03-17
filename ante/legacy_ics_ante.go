package ante

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gaiaerrors "github.com/cosmos/gaia/v28/types/errors"
	ics "github.com/cosmos/gaia/v28/x/legacy/ics"
)

// RejectLegacyICSDecorator rejects transactions that contain deprecated ICS
// provider messages. These type URLs are registered solely for historical query
// decoding; new transactions must not use them.
type RejectLegacyICSDecorator struct{}

func NewRejectLegacyICSDecorator() RejectLegacyICSDecorator {
	return RejectLegacyICSDecorator{}
}

// isLegacyICSMsg returns true if msg is one of the ICS provider stub types that
// were removed from this build. A type switch is used rather than sdk.MsgTypeURL
// for clarity and to avoid any reliance on the gogo proto registry at call time.
func isLegacyICSMsg(msg sdk.Msg) bool {
	switch msg.(type) {
	case *ics.MsgAssignConsumerKey,
		*ics.MsgConsumerAddition,
		*ics.MsgConsumerRemoval,
		*ics.MsgConsumerModification,
		*ics.MsgCreateConsumer,
		*ics.MsgUpdateConsumer,
		*ics.MsgRemoveConsumer,
		*ics.MsgChangeRewardDenoms,
		*ics.MsgUpdateParams,
		*ics.MsgSubmitConsumerMisbehaviour,
		*ics.MsgSubmitConsumerDoubleVoting,
		*ics.MsgOptIn,
		*ics.MsgOptOut,
		*ics.MsgSetConsumerCommissionRate:
		return true
	}
	return false
}

func (d RejectLegacyICSDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	if simulate {
		// Simulation is a read-only dry-run; the decorator does not reject
		// deprecated messages in this mode. Rejection only occurs on broadcast.
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		if isLegacyICSMsg(msg) {
			return ctx, errorsmod.Wrapf(gaiaerrors.ErrDeprecatedMessage,
				"legacy ICS message type %T is no longer accepted", msg)
		}
	}
	return next(ctx, tx, simulate)
}
