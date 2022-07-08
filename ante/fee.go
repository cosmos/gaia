package ante

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"

	"github.com/cosmos/gaia/v8/x/globalfee"
	"github.com/cosmos/gaia/v8/x/globalfee/types"
)

const maxBypassMinFeeMsgGasUsage = uint64(200_000)

// FeeWithBypassDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config) and global fee.
//
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true. If fee is high enough or not
// CheckTx, then call next AnteHandler.
//
// CONTRACT: Tx must implement FeeTx to use BypassMinFeeDecorator

var _ sdk.AnteDecorator = BypassMinFeeDecorator{}

type BypassMinFeeDecorator struct {
	BypassMinFeeMsgTypes []string
	GlobalMinFee         globalfee.ParamSource
}

const defaultZeroGlobalFeeDenom = "uatom"

func DefaultZeroGlobalFee() []sdk.DecCoin {
	return []sdk.DecCoin{sdk.NewDecCoinFromDec(defaultZeroGlobalFeeDenom, sdk.NewDec(0))}
}

func NewBypassMinFeeDecorator(bypassMsgTypes []string, paramSpace paramtypes.Subspace) BypassMinFeeDecorator {
	if !paramSpace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	return BypassMinFeeDecorator{
		BypassMinFeeMsgTypes: bypassMsgTypes,
		GlobalMinFee:         paramSpace,
	}
}

func (mfd BypassMinFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	feeCoins := feeTx.GetFee().Sort()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	//todo if need to check gas > uint64(len(msgs))*maxBypassMinFeeMsgGasUsage) here ?

	// Only check for minimum fees and global fee if the execution mode is CheckTx and the tx does
	// not contain operator configured bypass messages. If the tx does contain
	// operator configured bypass messages only, it's total gas must be less than
	// or equal to a constant, otherwise minimum fees and global fees are checked to prevent spam.

	containsOnlyBypassMinFeeMsgs := mfd.bypassMinFeeMsgs(msgs)
	doesNotExceedMaxGasUsage := gas <= uint64(len(msgs))*maxBypassMinFeeMsgGasUsage
	allowedToBypassMinFee := containsOnlyBypassMinFeeMsgs && doesNotExceedMaxGasUsage

	if ctx.IsCheckTx() && !simulate && !allowedToBypassMinFee {
		// check global fees
		if mfd.GlobalMinFee.Has(ctx, types.ParamStoreKeyMinGasPrices) {
			//requiredGlobalFees is sorted
			requiredGlobalFees := mfd.getGlobalFee(ctx, feeTx)
			if !DenomsSubsetOf(feeCoins, requiredGlobalFees) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fees %s is not a subset of required global fees %s", feeCoins, requiredGlobalFees)
			}
			// At least feeCoin amount must be greater than or equal to one of the requiredGlobalFees
			if !IsAnyGTEIncludingZero(feeCoins, requiredGlobalFees) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees for global fee; got: %s required: %s", feeCoins, requiredGlobalFees)
			}
		}

		// passed globalfee check, this means the fee denom is in the globalfee denom,  check min gas price
		minGasPrices := ctx.MinGasPrices()
		// if not all coins are zero, check fee with min_gas_price
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdk.NewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}
			// 1stake is not a subset of 0stake or 0photon or 1photon, not skip the min gas price check. this case can still use the DenomsSubsetOf() from sdk
			// empty set is subset of empty set for sdk DenomsSubsetOf
			// 0 stake is a subset of a 1stake, but the custom IsAnyGTEIncludingZero will be true as well.
			if !feeCoins.DenomsSubsetOf(requiredFees.Sort()) {
				return next(ctx, tx, simulate)
			}
			// requiredFees here is ensured not all zero, when check min_gas_price, fee might be zero. if min_gas_price=0stake,1photon, and feecoins is 0stake, it should not return err. so use IsAnyGTEIncludingZero() rather than IsAnyGTEIncludingZero() from sdk.
			if !IsAnyGTEIncludingZero(feeCoins, requiredFees) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	// when the tx is bypass msg type, still need to check the denom is not random denom
	// this is to prevent the situation that a bypass msg carries random fee denoms
	if ctx.IsCheckTx() && !simulate && allowedToBypassMinFee && mfd.GlobalMinFee.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		requiredFees := mfd.getGlobalFee(ctx, feeTx)
		if !DenomsSubsetOf(feeCoins, requiredFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fees denom is wrong; got: %s required: %s", feeCoins, requiredFees)
		}
	}

	return next(ctx, tx, simulate)
}

//coins must be sorted
func (mfd BypassMinFeeDecorator) getGlobalFee(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	var globalMinGasPrices sdk.DecCoins
	mfd.GlobalMinFee.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)

	// global fee is empty set, set global fee to 0uatom
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices = DefaultZeroGlobalFee()
	}
	requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(feeTx.GetGas()))
	for i, gp := range globalMinGasPrices {
		fee := gp.Amount.Mul(glDec)
		amount := fee.Ceil().RoundInt()
		requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, amount)
	}

	return requiredGlobalFees.Sort()
}

func (mfd BypassMinFeeDecorator) bypassMinFeeMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), mfd.BypassMinFeeMsgTypes) {
			continue
		}

		return false
	}

	return true
}

//utils function: GetTxPriority, DenomsSubsetOf, IsAnyGTEIncludingZero

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the fee
// provided in a transaction.
func GetTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		if c.Amount.IsInt64() {
			p = c.Amount.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}

// overwrite DenomsSubsetOf from sdk, to allow zero amt coins. e.g. 1stake is DenomsSubsetOf 0stake.
// DenomsSubsetOf returns true if coins's denoms is subset of coinsB's denoms.
// if coins is empty set, empty set is any sets' subset
func DenomsSubsetOf(coins, coinsB sdk.Coins) bool {
	for _, coin := range coins {
		// validate denom
		err := sdk.ValidateDenom(coin.Denom)
		if err != nil {
			panic(err)
		}

		if ok, _ := coinsB.Find(coin.Denom); !ok {
			return false
		}
	}

	return true
}

// overwrite the IsAnyGTEIncludingZero from sdk to allow zero coins.
// IsAnyGTEIncludingZero returns true if coins contain at least one denom that is present at a greater or equal amount in coinsB; it returns false otherwise.
// if CoinsB is emptyset, no coins sets are IsAnyGTEIncludingZero coinsB unless coins is also empty set.
// NOTE: IsAnyGTEIncludingZero operates under the invariant that both coin sets are sorted by denoms.
func IsAnyGTEIncludingZero(coins, coinsB sdk.Coins) bool {
	// no set is empty set's subset except empty set
	// this is different from sdk, sdk return false for coinsB empty
	if len(coinsB) == 0 && len(coins) == 0 {
		return true
	}
	if len(coinsB) == 0 {
		return false
	}
	// if feecoins empty, and globalfee has one denom of amt zero. feecoins equals to that 0denom.
	if len(coins) == 0 {
		// sdk.NewCoins will return non-zero coins
		coins = sdk.Coins{sdk.NewInt64Coin(coinsB.GetDenomByIndex(0), 0)}
		for i, coinB := range coinsB {
			if coinB.Amount.Equal(sdk.ZeroInt()) {
				coins = sdk.Coins{sdk.NewInt64Coin(coinsB.GetDenomByIndex(i), 0)}
			}
		}
	}

	for _, coin := range coins {
		amt := coinsB.AmountOf(coin.Denom)
		if coin.Amount.GTE(amt) {
			return true
		}
	}

	return false
}
