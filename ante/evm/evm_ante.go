package evm

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewAnteHandler() sdk.AnteHandler {
	panic("EVM Ante handler not set")
	return sdk.ChainAnteDecorators()
}
