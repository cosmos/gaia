package errors

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidChainID = errorsmod.Register(codespace, 11, "invalid chain ID")

	// ErrInvalidGasPrice returns an error if an invalid gas price is provided to the tx.
	ErrInvalidGasPrice = errorsmod.Register(codespace, 12, "invalid gas price")

	// ErrInvalidAmount returns an error if a tx contains an invalid amount.
	ErrInvalidAmount = errorsmod.Register(codespace, 13, "invalid transaction amount")

	// ErrInvalidGasFee returns an error if the tx gas fee is out of bound.
	ErrInvalidGasFee = errorsmod.Register(codespace, 14, "invalid gas fee")

	// ErrInvalidGasCap returns an error if the gas cap value is negative or invalid
	ErrInvalidGasCap = errorsmod.Register(codespace, 15, "invalid gas cap")

	ErrInvalidGasLimit = errorsmod.Register(codespace, 16, "gas limit must not be zero")

	ErrGasOverflow = errorsmod.Register(codespace, 17, "gas limit must be less than math.MaxInt64")
)
