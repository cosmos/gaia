package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type BankKeeper interface {
	// Methods imported from bank should be defined here
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)

	HasSupply(ctx context.Context, denom string) bool
	IterateTotalSupply(ctx context.Context, cb func(sdk.Coin) bool)

	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	BlockedAddr(addr sdk.AccAddress) bool
}

type AccountKeeper interface {
	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for community pool interactions.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
