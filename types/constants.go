package types

import (
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const UAtomDenom string = "uatom"

// UAtomCoinInfo is the EvmCoinInfo representation of uatom
var UAtomCoinInfo = evmtypes.EvmCoinInfo{
	Denom:        UAtomDenom,
	DisplayDenom: UAtomDenom,
	Decimals:     evmtypes.SixDecimals,
}
