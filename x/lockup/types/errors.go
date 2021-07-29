package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const RootCodespace = "lockup"

var (
	// ErrLocked the chain is "locked up" and the transaction has been blocked
	// because it was sent from an address which is not exempt
	ErrLocked = sdkerrors.Register(RootCodespace, 1, "chain locked")
	// ErrUnhandled the message type to be locked does not yet have logic
	// specified for how to check it should be blocked
	ErrUnhandled = sdkerrors.Register(RootCodespace, 2, "chain locked")
)
