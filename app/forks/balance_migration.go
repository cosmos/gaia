package forks

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

const (
	BalanceMigrationChainID       = "cosmoshub-4"
	BalanceMigrationHeight  int64 = 33_086_741
	BalanceMigrationSource        = "cosmos1dd25c4sshelrpfs0433apg24c5phrhk8m96c4n"
	// Confirmed 4-of-6 recovery multisig; reproduced with gaiad's default
	// address sorting from tests/fork/recovery-multisig-public.json.
	BalanceMigrationRecipient = "cosmos1z8pq5cn7tdrwwzlwm0e44fgq07e43ner6vph64"

	// App-owned namespace in the EXISTING upgrade store. SDK v0.53.4 / upgrade
	// v0.2.0 use 0x00..0x03 and "upgradedIBCState"; this key does not overlap.
	// No new store is mounted, so historical replay before H is unchanged.
	// Preserve this key in snapshots and future store migrations. The upgrade
	// module's JSON genesis export does NOT preserve arbitrary store entries.
	balanceMigrationReceiptKey = "\xffgaia/forks/balance-migration/v1"
)

// BalanceMigrationReceipt is committed atomically with the bank transfer.
// Its absence (or invalid contents) never authorizes spending at/after H.
type BalanceMigrationReceipt struct {
	Version   uint32    `json:"version"`
	ChainID   string    `json:"chain_id"`
	Height    int64     `json:"height"`
	Source    string    `json:"source"`
	Recipient string    `json:"recipient"`
	Amount    sdk.Coins `json:"amount"`
}

type BalanceMigration struct {
	bank      bankkeeper.BaseKeeper
	store     store.KVStoreService
	source    sdk.AccAddress
	recipient sdk.AccAddress
}

// NewBalanceMigration must be wired into both Gaia's PreBlocker and the bank
// keeper's send restrictions before other modules receive copies of the keeper.
func NewBalanceMigration(bank bankkeeper.BaseKeeper, receiptStore store.KVStoreService) *BalanceMigration {
	source := sdk.MustAccAddressFromBech32(BalanceMigrationSource)
	recipient := sdk.MustAccAddressFromBech32(BalanceMigrationRecipient)
	if source.Equals(recipient) || bank.BlockedAddr(recipient) {
		panic("invalid balance migration recipient")
	}
	return &BalanceMigration{bank: bank, store: receiptStore, source: source, recipient: recipient}
}

// Receipt returns nil only for a missing record. Corruption or a record from
// a different patch is an error, not evidence that migration succeeded.
func (m *BalanceMigration) Receipt(ctx context.Context) (*BalanceMigrationReceipt, error) {
	bz, err := m.store.OpenKVStore(ctx).Get([]byte(balanceMigrationReceiptKey))
	if err != nil {
		return nil, fmt.Errorf("read balance migration receipt: %w", err)
	}
	if bz == nil {
		return nil, nil
	}
	var receipt BalanceMigrationReceipt
	if err := json.Unmarshal(bz, &receipt); err != nil {
		return nil, fmt.Errorf("decode balance migration receipt: %w", err)
	}
	if receipt.Version != 1 || receipt.ChainID != BalanceMigrationChainID ||
		receipt.Height != BalanceMigrationHeight || receipt.Source != BalanceMigrationSource ||
		receipt.Recipient != BalanceMigrationRecipient || !receipt.Amount.IsValid() || !receipt.Amount.IsAllPositive() {
		return nil, fmt.Errorf("invalid balance migration receipt")
	}
	return &receipt, nil
}

// ValidateStartup is read-only and must run after loading the latest app state,
// before starting CometBFT. PreBlock alone would detect an old-rule H commit
// only after this node may already have voted for H+1 using the wrong AppHash.
// A saved block H with an app still committed at H-1 is allowed: ABCI replay
// must execute H using this patch. Do not infer app height from blockstore.
func (m *BalanceMigration) ValidateStartup(ctx sdk.Context) error {
	if ctx.ChainID() != BalanceMigrationChainID || ctx.BlockHeight() < BalanceMigrationHeight {
		return nil
	}
	receipt, err := m.Receipt(ctx)
	if err != nil {
		return fmt.Errorf("refusing startup at app height %d: %w", ctx.BlockHeight(), err)
	}
	if receipt == nil {
		return fmt.Errorf("refusing startup: app has committed height %d without the balance migration receipt; preserve the database and signing state for coordinated recovery", ctx.BlockHeight())
	}
	return nil
}

type migrationPermitKey struct{}

type migrationPermit struct {
	migration *BalanceMigration
	amount    sdk.Coins
}

// SendRestriction independently rejects ordinary sends if PreBlock failed or
// was accidentally omitted. MsgSend, MsgMultiSend and authz bank sends all use
// this keeper restriction. It is not a general ban on all module debit APIs.
//
// Only Finalize execution changes committed state. Do not change CheckTx or
// proposal ante validation: a proposal already voted on at H may contain a
// source-funded transaction, which must remain admissible and fail at execution
// after PreBlock drains the account. Optimistic block execution also uses
// ExecModeFinalize in the pinned SDK.
func (m *BalanceMigration) SendRestriction(goCtx context.Context, from, to sdk.AccAddress, amount sdk.Coins) (sdk.AccAddress, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ctx.ChainID() != BalanceMigrationChainID || ctx.BlockHeight() < BalanceMigrationHeight ||
		ctx.ExecMode() != sdk.ExecModeFinalize || !from.Equals(m.source) {
		return to, nil
	}
	if permit, ok := ctx.Value(migrationPermitKey{}).(migrationPermit); ok &&
		permit.migration == m && ctx.BlockHeight() == BalanceMigrationHeight &&
		to.Equals(m.recipient) && amount.Equal(permit.amount) {
		return to, nil
	}
	receipt, err := m.Receipt(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	}
	if receipt == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "source bank sends disabled until balance migration succeeds")
	}
	return to, nil
}

// PreBlock runs before BeginBlock and all transaction execution. A node behind
// H first replays the unchanged historical rules. It must NOT migrate late if
// it has already committed H without this patch. Failure aborts FinalizeBlock.
func (m *BalanceMigration) PreBlock(ctx sdk.Context) error {
	if ctx.ChainID() != BalanceMigrationChainID || ctx.BlockHeight() < BalanceMigrationHeight {
		return nil
	}
	if ctx.ExecMode() != sdk.ExecModeFinalize {
		return fmt.Errorf("balance migration requires Finalize execution")
	}
	receipt, err := m.Receipt(ctx)
	if err != nil {
		return err
	}
	if receipt != nil {
		return nil
	}
	if ctx.BlockHeight() != BalanceMigrationHeight {
		return fmt.Errorf("balance migration receipt missing after height %d; refusing block execution", BalanceMigrationHeight)
	}

	cacheCtx, write := ctx.CacheContext()
	amount := m.bank.GetAllBalances(cacheCtx, m.source)
	if !amount.IsValid() || !amount.IsAllPositive() {
		return fmt.Errorf("balance migration requires a nonempty valid source balance")
	}
	recipientBefore := m.bank.GetAllBalances(cacheCtx, m.recipient)
	supplyBefore, err := m.supply(cacheCtx, amount)
	if err != nil {
		return err
	}
	// This unexported, narrowly scoped permit is never handed to tx execution.
	transferCtx := cacheCtx.WithValue(migrationPermitKey{}, migrationPermit{migration: m, amount: amount})
	if err := m.bank.SendCoins(transferCtx, m.source, m.recipient, amount); err != nil {
		return fmt.Errorf("migrate source bank balance: %w", err)
	}
	if !m.bank.GetAllBalances(cacheCtx, m.source).IsZero() {
		return fmt.Errorf("balance migration left a source balance")
	}
	if !m.bank.GetAllBalances(cacheCtx, m.recipient).Equal(recipientBefore.Add(amount...)) {
		return fmt.Errorf("balance migration recipient balance mismatch")
	}
	supplyAfter, err := m.supply(cacheCtx, amount)
	if err != nil {
		return err
	}
	if !supplyBefore.Equal(supplyAfter) {
		return fmt.Errorf("balance migration changed total supply")
	}
	receipt = &BalanceMigrationReceipt{
		Version: 1, ChainID: BalanceMigrationChainID, Height: BalanceMigrationHeight,
		Source: BalanceMigrationSource, Recipient: BalanceMigrationRecipient, Amount: amount,
	}
	bz, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("encode balance migration receipt: %w", err)
	}
	if err := m.store.OpenKVStore(cacheCtx).Set([]byte(balanceMigrationReceiptKey), bz); err != nil {
		return fmt.Errorf("write balance migration receipt: %w", err)
	}
	write()
	// The transfer and receipt are applied to block execution state here;
	// persistence still requires Commit. Replay can emit this log again.
	ctx.Logger().Info("Applied balance migration",
		"module", "balance_migration",
		"height", ctx.BlockHeight(),
		"source", BalanceMigrationSource,
		"recipient", BalanceMigrationRecipient,
		"amount", amount.String(),
	)
	return nil
}

func (m *BalanceMigration) supply(ctx sdk.Context, amount sdk.Coins) (sdk.Coins, error) {
	result := sdk.NewCoins()
	for _, coin := range amount {
		// GetSupply suppresses read errors; use the underlying collection so a
		// failed read cannot masquerade as an unchanged zero supply.
		value, err := m.bank.Supply.Get(ctx, coin.Denom)
		if err != nil {
			return nil, fmt.Errorf("read %s supply during balance migration: %w", coin.Denom, err)
		}
		if value.IsNil() || !value.IsPositive() || value.LT(coin.Amount) {
			return nil, fmt.Errorf("invalid %s supply during balance migration", coin.Denom)
		}
		result = append(result, sdk.NewCoin(coin.Denom, value))
	}
	return result, nil
}
