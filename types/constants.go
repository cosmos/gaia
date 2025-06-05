package types

import (
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	UAtomDenom        string = "uatom"
	AtomDenom         string = "atom"
	DefaultEVMChainID        = uint64(4231)
)

var (
	// UAtomCoinInfo is the EvmCoinInfo representation of uatom
	UAtomCoinInfo = evmtypes.EvmCoinInfo{
		Denom:         UAtomDenom,
		ExtendedDenom: AtomDenom,
		DisplayDenom:  AtomDenom,
		Decimals:      evmtypes.SixDecimals,
	}
)
