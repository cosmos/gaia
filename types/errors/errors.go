package errors

import (
	errorsmod "cosmossdk.io/errors"
)

const codespace = "gaia"

var (
	// ErrTxDecode is returned if we cannot parse a transaction
	ErrTxDecode = errorsmod.Register(codespace, 1, "tx parse error")
	// ErrUnauthorized is used whenever a request without sufficient
	// authorization is handled.
	ErrUnauthorized = errorsmod.Register(codespace, 2, "unauthorized")

	// ErrInsufficientFunds is used when the account cannot pay requested amount.
	ErrInsufficientFunds = errorsmod.Register(codespace, 3, "insufficient funds")

	// ErrInsufficientFunds is used when the account cannot pay requested amount.
	ErrInsufficientFee = errorsmod.Register(codespace, 4, "insufficient fee")

	// ErrInvalidCoins is used when sdk.Coins are invalid.
	ErrInvalidCoins = errorsmod.Register(codespace, 5, "invalid coins")

	// ErrInvalidType defines an error an invalid type.
	ErrInvalidType = errorsmod.Register(codespace, 6, "invalid type")

	// ErrLogic defines an internal logic error, e.g. an invariant or assertion
	// that is violated. It is a programmer error, not a user-facing error.
	ErrLogic = errorsmod.Register(codespace, 7, "internal logic error")

	// ErrNotFound defines an error when requested entity doesn't exist in the state.
	ErrNotFound = errorsmod.Register(codespace, 8, "not found")

	// ErrInsufficientStake is used when the account has insufficient staked tokens.
	ErrInsufficientStake = errorsmod.Register(codespace, 9, "insufficient stake")

	// ErrInvalidExpeditedProposal is used when an expedite proposal is submitted for an unsupported proposal type.
	ErrInvalidExpeditedProposal = errorsmod.Register(codespace, 10, "unsupported expedited proposal type")
)
