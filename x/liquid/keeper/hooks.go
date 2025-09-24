package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v24/x/liquid/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper

	//CONTRACT: assumes serial calling of hooks. (no parallel msg/tx/block processing)
	// while staking(delegate) -> hooks are called in either of these orders:
	// BeforeDelegationCreated -> AfterDelegationModified
	// BeforeDelegationModified -> AfterDelegationModified
	// while unstaking (undelegate) -> hooks are called either of these orders:
	// BeforeDelegationModified -> BeforeDelegationRemoved
	// BeforeDelegationModified -> AfterDelegationModified

	predelegation *stakingtypes.Delegation
}

var _ stakingtypes.StakingHooks = Hooks{}

// Create new liquid hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{
		k:             k,
		predelegation: nil,
	}
}

// initialize liquid validator record
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	val, err := h.k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}
	lVal := types.NewLiquidValidator(val.GetOperator())
	return h.k.SetLiquidValidator(ctx, lVal)
}

func (h Hooks) AfterValidatorRemoved(ctx context.Context, _ sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return h.k.RemoveLiquidValidator(ctx, valAddr)
}

func (h Hooks) BeforeDelegationCreated(_ context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if h.k.DelegatorIsLiquidStaker(delAddr) {
		if h.predelegation != nil {
			return types.ErrPreHookIsNotNil
		}

		h.predelegation = &stakingtypes.Delegation{
			DelegatorAddress: delAddr.String(),
			ValidatorAddress: valAddr.String(),
			Shares:           sdkmath.LegacyZeroDec(),
		}
	}
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if h.k.DelegatorIsLiquidStaker(delAddr) {
		if h.predelegation != nil {
			return types.ErrPreHookIsNotNil
		}

		predel, err := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
		if err != nil {
			return err
		}
		h.predelegation = &predel
	}
	return nil
}

func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if h.k.DelegatorIsLiquidStaker(delAddr) {
		if h.predelegation == nil {
			return types.ErrPreHookIsNil
		}
		del, err := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
		if err != nil {
			return err
		}
		if del.Shares.GT(h.predelegation.Shares) {
			// is bonding
			diffShares := del.Shares.Sub(h.predelegation.Shares)
			validator, err := h.k.stakingKeeper.GetValidator(ctx, valAddr)
			if err != nil {
				return err
			}
			diffTokens := validator.TokensFromSharesTruncated(del.Shares).TruncateInt().
				Sub(validator.TokensFromSharesTruncated(h.predelegation.Shares).TruncateInt())
			if err := h.k.SafelyIncreaseTotalLiquidStakedTokens(ctx, diffTokens, true); err != nil {
				return err
			}
			_, err = h.k.SafelyIncreaseValidatorLiquidShares(ctx, valAddr, diffShares, true)
			if err != nil {
				return err
			}
		} else if del.Shares.LT(h.predelegation.Shares) {
			// is unbonding
			diffShares := h.predelegation.Shares.Sub(del.Shares)
			validator, err := h.k.stakingKeeper.GetValidator(ctx, valAddr)
			if err != nil {
				return err
			}
			diffTokens := validator.TokensFromSharesTruncated(h.predelegation.Shares).TruncateInt().
				Sub(validator.TokensFromSharesTruncated(del.Shares).TruncateInt())
			if err := h.k.DecreaseTotalLiquidStakedTokens(ctx, diffTokens); err != nil {
				return err
			}
			_, err = h.k.DecreaseValidatorLiquidShares(ctx, valAddr, diffShares)
			if err != nil {
				return err
			}
		} else {
			return types.ErrInvalidHookInvocation
		}

		// reset prehook
		h.predelegation = nil
	}
	return nil
}

func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	// fraction = tokens_to_burn / validator.Tokens
	validator, err := h.k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}
	liquidVal, err := h.k.GetLiquidValidator(ctx, valAddr)
	if err != nil {
		return err
	}
	initialLiquidTokens := validator.TokensFromShares(liquidVal.LiquidShares).TruncateInt()
	slashedLiquidTokens := fraction.Mul(sdkmath.LegacyNewDecFromInt(initialLiquidTokens))

	decrease := slashedLiquidTokens.TruncateInt()
	if err := h.k.DecreaseTotalLiquidStakedTokens(ctx, decrease); err != nil {
		// This only error's if the total liquid staked tokens underflows
		// which would indicate there's a corrupted state where the validator has
		// liquid tokens that are not accounted for in the global total
		panic(err)
	}
	return nil
}

func (h Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	if h.k.DelegatorIsLiquidStaker(delAddr) {
		if h.predelegation == nil {
			return types.ErrPreHookIsNil
		}
		// is unbonding.
		validator, err := h.k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return err
		}
		tokens := validator.TokensFromSharesTruncated(h.predelegation.Shares).TruncateInt()
		if err := h.k.DecreaseTotalLiquidStakedTokens(ctx, tokens); err != nil {
			return err
		}
		_, err = h.k.DecreaseValidatorLiquidShares(ctx, valAddr, h.predelegation.Shares)
		if err != nil {
			return err
		}

		// reset prehook
		h.predelegation = nil
	}
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}

func (h Hooks) BeforeTokenizeShareRecordRemoved(_ context.Context, _ uint64) error {
	return nil
}

////
//// trimmed imports
//import (
//"context"
//"crypto/sha256"
//"encoding/binary"
//"fmt"
//
//sdk "github.com/cosmos/cosmos-sdk/types"
//"github.com/cosmos/cosmos-sdk/codec"
//stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
//)

// HookKind constants
//const (
//	HookCreate        = "create"
//	HookSharesMod     = "shares_modified"
//	HookRemove        = "remove"
//	HookRedelegateOut = "redelegate_out" // optional
//)
//
//// DelegationEvent is the pre-state snapshot pushed on Before*
//type DelegationEvent struct {
//	EventID            uint64
//	HookKind           string
//	Delegator          sdk.AccAddress
//	Validator          sdk.ValAddress
//	PreShares          sdk.Dec
//	PreValidatorTokens sdk.Dec // validator tokens pre-change (to compute tokens delta safely)
//	PreValidatorShares sdk.Dec // validator delegator shares pre-change
//}
//
//// Keeper (partial) -- you must provide cdc, tKey, stakingKeeper, and any cap params
////type Keeper struct {
////	cdc           codec.BinaryCodec
////	tKey          sdk.TransientStoreKey
////	stakingKeeper StakingKeeper
////	// per-validator cap config (example)
////	validatorCap sdk.Int // maximum tokens allowed per validator per tx (for example)
////}
//
////// StakingKeeper interface subset
////type StakingKeeper interface {
////	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (stakingtypes.Delegation, bool)
////	GetValidator(ctx sdk.Context, valAddr sdk.ValAddress) (stakingtypes.Validator, bool)
////}
//
//// prefixes
//var (
//	prefixStack   = []byte{0x01} // stack per ns||delegator||validator
//	prefixEvent   = []byte{0x02} // store event by ns||eventID
//	prefixCounter = []byte{0x03} // per-ns counter for eventID
//	prefixAccVal  = []byte{0x04} // per-ns per-validator accumulator (sdk.Int bytes)
//	prefixExecCtr = []byte{0x05} // per-block exec counter fallback
//)
//
//// txNsBytes get namespace for current execution (tx-scoped when possible)
//func (k Keeper) txNsBytes(ctx sdk.Context) []byte {
//	txBytes := ctx.TxBytes()
//	if len(txBytes) > 0 {
//		h := sha256.Sum256(txBytes)
//		return h[:]
//	}
//	// fallback: use block height + block time + a per-block transient exec counter
//	store := ctx.TransientStore(k.tKey)
//	bh := ctx.BlockHeight()
//	bt := ctx.BlockTime().UnixNano()
//
//	// exec counter per block to distinguish multiple begin/end calls across modules
//	ctrKey := append(prefixExecCtr, []byte(fmt.Sprintf("|%d|", bh))...)
//	bz := store.Get(ctrKey)
//	var ctr uint64
//	if len(bz) >= 8 {
//		ctr = binary.BigEndian.Uint64(bz)
//	}
//	ctr++
//	nb := make([]byte, 8)
//	binary.BigEndian.PutUint64(nb, ctr)
//	store.Set(ctrKey, nb)
//
//	keyStr := fmt.Sprintf("bh:%d|bt:%d|ctr:%d", bh, bt, ctr)
//	h := sha256.Sum256([]byte(keyStr))
//	return h[:]
//}
//
//// nextEventID increments per-ns counter and returns id
//func (k Keeper) nextEventID(ctx sdk.Context) uint64 {
//	store := ctx.TransientStore(k.tKey)
//	ns := k.txNsBytes(ctx)
//	key := append(prefixCounter, ns...)
//	bz := store.Get(key)
//	var ctr uint64
//	if len(bz) >= 8 {
//		ctr = binary.BigEndian.Uint64(bz)
//	}
//	ctr++
//	nb := make([]byte, 8)
//	binary.BigEndian.PutUint64(nb, ctr)
//	store.Set(key, nb)
//	return ctr
//}
//
//func stackKey(ns []byte, del sdk.AccAddress, val sdk.ValAddress) []byte {
//	k := make([]byte, 0, 1+len(ns)+1+len(del)+1+len(val))
//	k = append(k, prefixStack...)
//	k = append(k, ns...)
//	k = append(k, 0x00)
//	k = append(k, del.Bytes()...)
//	k = append(k, 0x00)
//	k = append(k, val.Bytes()...)
//	return k
//}
//func eventKey(ns []byte, eventID uint64) []byte {
//	k := make([]byte, 0, 1+len(ns)+8)
//	k = append(k, prefixEvent...)
//	k = append(k, ns...)
//	bz := make([]byte, 8)
//	binary.BigEndian.PutUint64(bz, eventID)
//	k = append(k, bz...)
//	return k
//}
//func accValKey(ns []byte, val sdk.ValAddress) []byte {
//	k := make([]byte, 0, 1+len(ns)+len(val))
//	k = append(k, prefixAccVal...)
//	k = append(k, ns...)
//	k = append(k, 0x00)
//	k = append(k, val.Bytes()...)
//	return k
//}
//
//// pushEvent: push eventID onto stack for (del,val) (LIFO behavior: append at end)
//func (k Keeper) pushEvent(ctx sdk.Context, del sdk.AccAddress, val sdk.ValAddress, event DelegationEvent) {
//	store := ctx.TransientStore(k.tKey)
//	ns := k.txNsBytes(ctx)
//	// store event by id
//	ekey := eventKey(ns, event.EventID)
//	bz := k.cdc.MustMarshal(&event)
//	store.Set(ekey, bz)
//
//	// append id to stack
//	sk := stackKey(ns, del, val)
//	old := store.Get(sk)
//	idbz := make([]byte, 8)
//	binary.BigEndian.PutUint64(idbz, event.EventID)
//	new := append(old, idbz...)
//	store.Set(sk, new)
//}
//
//// popEvent: pop the topmost event id for (del,val). LIFO: read last 8 bytes.
//func (k Keeper) popEvent(ctx sdk.Context, del sdk.AccAddress, val sdk.ValAddress) (DelegationEvent, bool) {
//	var ev DelegationEvent
//	store := ctx.TransientStore(k.tKey)
//	ns := k.txNsBytes(ctx)
//	sk := stackKey(ns, del, val)
//	old := store.Get(sk)
//	if len(old) < 8 {
//		return ev, false
//	}
//	// take last 8 bytes
//	lastIdx := len(old) - 8
//	id := binary.BigEndian.Uint64(old[lastIdx:])
//	// trim
//	new := old[:lastIdx]
//	if len(new) == 0 {
//		store.Delete(sk)
//	} else {
//		store.Set(sk, new)
//	}
//	// load event
//	ekey := eventKey(ns, id)
//	bz := store.Get(ekey)
//	if len(bz) == 0 {
//		return ev, false
//	}
//	k.cdc.MustUnmarshal(bz, &ev)
//	// delete event entry now or after processing
//	store.Delete(ekey)
//	return ev, true
//}
//
//// accValAdd: accumulate token delta to per-validator accumulator (sdk.Int)
//func (k Keeper) accValAdd(ctx sdk.Context, val sdk.ValAddress, delta sdk.Int) {
//	store := ctx.TransientStore(k.tKey)
//	ns := k.txNsBytes(ctx)
//	ak := accValKey(ns, val)
//	old := store.Get(ak)
//	var cur sdk.Int
//	if len(old) > 0 {
//		cur.Unmarshal([]byte(old))
//	} else {
//		cur = sdk.ZeroInt()
//	}
//	cur = cur.Add(delta)
//	bz, _ := cur.Marshal()
//	store.Set(ak, bz)
//}
//func (k Keeper) accValGet(ctx sdk.Context, val sdk.ValAddress) sdk.Int {
//	store := ctx.TransientStore(k.tKey)
//	ns := k.txNsBytes(ctx)
//	ak := accValKey(ns, val)
//	old := store.Get(ak)
//	var cur sdk.Int
//	if len(old) > 0 {
//		cur.Unmarshal([]byte(old))
//	} else {
//		cur = sdk.ZeroInt()
//	}
//	return cur
//}
//
//// --- Hooks implementation ---
//
//type Hooks struct {
//	k Keeper
//}
//
//var _ stakingtypes.StakingHooks = Hooks{}
//
//// BeforeDelegationCreated
//func (h Hooks) BeforeDelegationCreated(ctx context.Context, del sdk.AccAddress, val sdk.ValAddress) error {
//	sdkCtx := sdk.UnwrapSDKContext(ctx)
//
//	// take snapshot
//	preShares := sdk.ZeroDec()
//	if delg, found := h.k.stakingKeeper.GetDelegation(sdkCtx, del, val); found {
//		preShares = delg.Shares
//	}
//	preValTokens := sdk.ZeroDec()
//	preValShares := sdk.ZeroDec()
//	if v, ok := h.k.stakingKeeper.GetValidator(sdkCtx, val); ok {
//		preValTokens = v.Tokens.ToDec()
//		preValShares = v.DelegatorShares
//	}
//
//	id := h.k.nextEventID(sdkCtx)
//	ev := DelegationEvent{
//		EventID:            id,
//		HookKind:           HookCreate,
//		Delegator:          del,
//		Validator:          val,
//		PreShares:          preShares,
//		PreValidatorTokens: preValTokens,
//		PreValidatorShares: preValShares,
//	}
//	h.k.pushEvent(sdkCtx, del, val, ev)
//	return nil
//}
//
//// BeforeDelegationSharesModified
//func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, del sdk.AccAddress, val sdk.ValAddress) error {
//	sdkCtx := sdk.UnwrapSDKContext(ctx)
//	preShares := sdk.ZeroDec()
//	if delg, found := h.k.stakingKeeper.GetDelegation(sdkCtx, del, val); found {
//		preShares = delg.Shares
//	}
//	preValTokens := sdk.ZeroDec()
//	preValShares := sdk.ZeroDec()
//	if v, ok := h.k.stakingKeeper.GetValidator(sdkCtx, val); ok {
//		preValTokens = v.Tokens.ToDec()
//		preValShares = v.DelegatorShares
//	}
//	id := h.k.nextEventID(sdkCtx)
//	ev := DelegationEvent{
//		EventID:            id,
//		HookKind:           HookSharesMod,
//		Delegator:          del,
//		Validator:          val,
//		PreShares:          preShares,
//		PreValidatorTokens: preValTokens,
//		PreValidatorShares: preValShares,
//	}
//	h.k.pushEvent(sdkCtx, del, val, ev)
//	return nil
//}
//
//// BeforeDelegationRemoved
//func (h Hooks) BeforeDelegationRemoved(ctx context.Context, del sdk.AccAddress, val sdk.ValAddress) error {
//	sdkCtx := sdk.UnwrapSDKContext(ctx)
//	preShares := sdk.ZeroDec()
//	if delg, found := h.k.stakingKeeper.GetDelegation(sdkCtx, del, val); found {
//		preShares = delg.Shares
//	}
//	preValTokens := sdk.ZeroDec()
//	preValShares := sdk.ZeroDec()
//	if v, ok := h.k.stakingKeeper.GetValidator(sdkCtx, val); ok {
//		preValTokens = v.Tokens.ToDec()
//		preValShares = v.DelegatorShares
//	}
//	id := h.k.nextEventID(sdkCtx)
//	ev := DelegationEvent{
//		EventID:            id,
//		HookKind:           HookRemove,
//		Delegator:          del,
//		Validator:          val,
//		PreShares:          preShares,
//		PreValidatorTokens: preValTokens,
//		PreValidatorShares: preValShares,
//	}
//	h.k.pushEvent(sdkCtx, del, val, ev)
//	return nil
//}
//
//// AfterDelegationModified
//func (h Hooks) AfterDelegationModified(ctx context.Context, del sdk.AccAddress, val sdk.ValAddress) error {
//	sdkCtx := sdk.UnwrapSDKContext(ctx)
//
//	ev, ok := h.k.popEvent(sdkCtx, del, val)
//	if !ok {
//		// no matching before snapshot for this (del,val) — safe fallback: no-op
//		return nil
//	}
//
//	// load post-shares
//	postShares := sdk.ZeroDec()
//	if delg, found := h.k.stakingKeeper.GetDelegation(sdkCtx, del, val); found {
//		postShares = delg.Shares
//	}
//
//	// compute deltaShares
//	deltaShares := postShares.Sub(ev.PreShares)
//	if deltaShares.IsZero() {
//		return nil
//	}
//
//	// compute token delta using validator post-state (or pre-state)
//	valObj, ok := h.k.stakingKeeper.GetValidator(sdkCtx, val)
//	if !ok || valObj.DelegatorShares.IsZero() {
//		// can't compute -> safe no-op
//		return nil
//	}
//	// Use validator CURRENT tokens/shares (post-change) to derive token delta
//	// tokenDelta = deltaShares * validator.Tokens / validator.DelegatorShares
//	tokensDec := valObj.Tokens.ToDec()
//	tokenDeltaDec := deltaShares.Mul(tokensDec).Quo(valObj.DelegatorShares)
//	tokenDeltaInt := tokenDeltaDec.TruncateInt()
//
//	// accumulate per-validator (for cap checks)
//	h.k.accValAdd(sdkCtx, val, tokenDeltaInt)
//
//	// If you want immediate abort on cap exceed:
//	// total := h.k.accValGet(sdkCtx, val)
//	// if total.GT(h.k.validatorCap) {
//	//     return sdkerrors.Wrapf(..., "validator cap exceeded")
//	// }
//
//	// otherwise just continue and finalize later (EndBlocker or after all hooks)
//	return nil
//}
