package bindings

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	tokenfactorykeeper "github.com/cosmos/gaia/v23/x/tokenfactory/keeper"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

func RegisterCustomPlugins(
	bank bankkeeper.Keeper,
	tokenFactory *tokenfactorykeeper.Keeper,
) []wasmkeeper.Option {
	wasmQueryPlugin := NewQueryPlugin(bank, tokenFactory)

	queryPluginOpt := wasmkeeper.WithQueryPlugins(&wasmkeeper.QueryPlugins{
		Custom: CustomQuerier(wasmQueryPlugin),
	})
	messengerDecoratorOpt := wasmkeeper.WithMessageHandlerDecorator(
		CustomMessageDecorator(bank, tokenFactory),
	)

	return []wasmkeeper.Option{
		queryPluginOpt,
		messengerDecoratorOpt,
	}
}
