package wasmbinding

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/codec/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	tokenfactorykeeper "github.com/cosmos/gaia/v23/x/tokenfactory/keeper"

	tokenfactorybindings "github.com/cosmos/gaia/v23/x/tokenfactory/bindings"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tokenfactorybindingstypes "github.com/cosmos/gaia/v23/x/tokenfactory/bindings/types"
)

func CustomMessageDecorator(bank bankkeeper.Keeper, tokenFactory *tokenfactorykeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	tokenfactoryCustomMessenger := tokenfactorybindings.NewCustomMessenger(nil, bank, tokenFactory)

	return func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &CustomMessenger{
			wrapped:                     old,
			tokenfactoryCustomMessenger: tokenfactoryCustomMessenger,
		}
	}
}

var emptyMsgResp = [][]*types.Any{}

type CustomMessenger struct {
	wrapped                     wasmkeeper.Messenger
	tokenfactoryCustomMessenger *tokenfactorybindings.CustomMessenger
}

var _ wasmkeeper.Messenger = (*CustomMessenger)(nil)

func (m *CustomMessenger) DispatchMsg(ctx sdk.Context, contractAddr sdk.AccAddress, contractIBCPortID string, msg wasmvmtypes.CosmosMsg) (events []sdk.Event, data [][]byte, msgResponses [][]*types.Any, err error) {
	if msg.Custom != nil {
		var contractMsg tokenfactorybindingstypes.TokenFactoryMsg
		if err := json.Unmarshal(msg.Custom, &contractMsg); err != nil {
			if contractMsg.CreateDenom != nil {
				return m.tokenfactoryCustomMessenger.CreateDenom(ctx, contractAddr, contractMsg.CreateDenom)
			}
			if contractMsg.MintTokens != nil {
				return m.tokenfactoryCustomMessenger.MintTokens(ctx, contractAddr, contractMsg.MintTokens)
			}
			if contractMsg.ChangeAdmin != nil {
				return m.tokenfactoryCustomMessenger.ChangeAdmin(ctx, contractAddr, contractMsg.ChangeAdmin)
			}
			if contractMsg.BurnTokens != nil {
				return m.tokenfactoryCustomMessenger.BurnTokens(ctx, contractAddr, contractMsg.BurnTokens)
			}
			if contractMsg.SetMetadata != nil {
				return m.tokenfactoryCustomMessenger.SetMetadata(ctx, contractAddr, contractMsg.SetMetadata)
			}
			if contractMsg.ForceTransfer != nil {
				return m.tokenfactoryCustomMessenger.ForceTransfer(ctx, contractAddr, contractMsg.ForceTransfer)
			}
		}
	}
	return m.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}
