package types

import (
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	UAtomDenom string = "uatom"
	AtomDenom  string = "atom"
)

var (
	// UAtomCoinInfo is the EvmCoinInfo representation of uatom
	UAtomCoinInfo = evmtypes.EvmCoinInfo{
		Denom:        UAtomDenom,
		DisplayDenom: AtomDenom,
		Decimals:     evmtypes.SixDecimals,
	}
)
