package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	// CancelOrderLifeSpan is the lifespan of order cancellation.
	CancelOrderLifeSpan int64 = 0

	// MinReserveCoinNum is the minimum number of reserve coins in each liquidity pool.
	MinReserveCoinNum uint32 = 2

	// MaxReserveCoinNum is the maximum number of reserve coins in each liquidity pool.
	MaxReserveCoinNum uint32 = 2

	// DefaultUnitBatchHeight is the default number of blocks in one batch. This param is used for scalability.
	DefaultUnitBatchHeight uint32 = 1

	// DefaultPoolTypeID is the default pool type id. The only supported pool type id is 1.
	DefaultPoolTypeID uint32 = 1

	// DefaultSwapTypeID is the default swap type id. The only supported swap type (instant swap) id is 1.
	DefaultSwapTypeID uint32 = 1

	// DefaultCircuitBreakerEnabled is the default circuit breaker status. This param is used for a contingency plan.
	DefaultCircuitBreakerEnabled = false
)

// Parameter store keys
var (
	KeyPoolTypes              = []byte("PoolTypes")
	KeyMinInitDepositAmount   = []byte("MinInitDepositAmount")
	KeyInitPoolCoinMintAmount = []byte("InitPoolCoinMintAmount")
	KeyMaxReserveCoinAmount   = []byte("MaxReserveCoinAmount")
	KeySwapFeeRate            = []byte("SwapFeeRate")
	KeyPoolCreationFee        = []byte("PoolCreationFee")
	KeyUnitBatchHeight        = []byte("UnitBatchHeight")
	KeyWithdrawFeeRate        = []byte("WithdrawFeeRate")
	KeyMaxOrderAmountRatio    = []byte("MaxOrderAmountRatio")
	KeyCircuitBreakerEnabled  = []byte("CircuitBreakerEnabled")
)

var (
	DefaultMinInitDepositAmount   = sdk.NewInt(1000000)
	DefaultInitPoolCoinMintAmount = sdk.NewInt(1000000)
	DefaultMaxReserveCoinAmount   = sdk.ZeroInt()
	DefaultSwapFeeRate            = sdk.NewDecWithPrec(3, 3) // "0.003000000000000000"
	DefaultWithdrawFeeRate        = sdk.ZeroDec()
	DefaultMaxOrderAmountRatio    = sdk.NewDecWithPrec(1, 1) // "0.100000000000000000"
	DefaultPoolCreationFee        = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(40000000)))
	DefaultPoolType               = PoolType{
		Id:                1,
		Name:              "StandardLiquidityPool",
		MinReserveCoinNum: MinReserveCoinNum,
		MaxReserveCoinNum: MaxReserveCoinNum,
		Description:       "Standard liquidity pool with pool price function X/Y, ESPM constraint, and two kinds of reserve coins",
	}
	DefaultPoolTypes = []PoolType{DefaultPoolType}

	MinOfferCoinAmount = sdk.NewInt(100)
)

var _ paramstypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns the default liquidity module parameters.
func DefaultParams() Params {
	return Params{
		PoolTypes:              DefaultPoolTypes,
		MinInitDepositAmount:   DefaultMinInitDepositAmount,
		InitPoolCoinMintAmount: DefaultInitPoolCoinMintAmount,
		MaxReserveCoinAmount:   DefaultMaxReserveCoinAmount,
		PoolCreationFee:        DefaultPoolCreationFee,
		SwapFeeRate:            DefaultSwapFeeRate,
		WithdrawFeeRate:        DefaultWithdrawFeeRate,
		MaxOrderAmountRatio:    DefaultMaxOrderAmountRatio,
		UnitBatchHeight:        DefaultUnitBatchHeight,
		CircuitBreakerEnabled:  DefaultCircuitBreakerEnabled,
	}
}

// ParamSetPairs implements paramstypes.ParamSet.
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyPoolTypes, &p.PoolTypes, validatePoolTypes),
		paramstypes.NewParamSetPair(KeyMinInitDepositAmount, &p.MinInitDepositAmount, validateMinInitDepositAmount),
		paramstypes.NewParamSetPair(KeyInitPoolCoinMintAmount, &p.InitPoolCoinMintAmount, validateInitPoolCoinMintAmount),
		paramstypes.NewParamSetPair(KeyMaxReserveCoinAmount, &p.MaxReserveCoinAmount, validateMaxReserveCoinAmount),
		paramstypes.NewParamSetPair(KeyPoolCreationFee, &p.PoolCreationFee, validatePoolCreationFee),
		paramstypes.NewParamSetPair(KeySwapFeeRate, &p.SwapFeeRate, validateSwapFeeRate),
		paramstypes.NewParamSetPair(KeyWithdrawFeeRate, &p.WithdrawFeeRate, validateWithdrawFeeRate),
		paramstypes.NewParamSetPair(KeyMaxOrderAmountRatio, &p.MaxOrderAmountRatio, validateMaxOrderAmountRatio),
		paramstypes.NewParamSetPair(KeyUnitBatchHeight, &p.UnitBatchHeight, validateUnitBatchHeight),
		paramstypes.NewParamSetPair(KeyCircuitBreakerEnabled, &p.CircuitBreakerEnabled, validateCircuitBreakerEnabled),
	}
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates parameters.
func (p Params) Validate() error {
	for _, v := range []struct {
		value     interface{}
		validator func(interface{}) error
	}{
		{p.PoolTypes, validatePoolTypes},
		{p.MinInitDepositAmount, validateMinInitDepositAmount},
		{p.InitPoolCoinMintAmount, validateInitPoolCoinMintAmount},
		{p.MaxReserveCoinAmount, validateMaxReserveCoinAmount},
		{p.PoolCreationFee, validatePoolCreationFee},
		{p.SwapFeeRate, validateSwapFeeRate},
		{p.WithdrawFeeRate, validateWithdrawFeeRate},
		{p.MaxOrderAmountRatio, validateMaxOrderAmountRatio},
		{p.UnitBatchHeight, validateUnitBatchHeight},
		{p.CircuitBreakerEnabled, validateCircuitBreakerEnabled},
	} {
		if err := v.validator(v.value); err != nil {
			return err
		}
	}
	return nil
}

func validatePoolTypes(i interface{}) error {
	v, ok := i.([]PoolType)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if len(v) == 0 {
		return fmt.Errorf("pool types must not be empty")
	}

	for i, p := range v {
		if int(p.Id) != i+1 {
			return fmt.Errorf("pool type ids must be sorted")
		}
		if p.MaxReserveCoinNum > MaxReserveCoinNum || MinReserveCoinNum > p.MinReserveCoinNum {
			return fmt.Errorf("min, max reserve coin num value of pool types are out of bounds")
		}
	}

	if len(v) > 1 || !v[0].Equal(DefaultPoolType) {
		return fmt.Errorf("the only supported pool type is 1")
	}

	return nil
}

func validateMinInitDepositAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("minimum initial deposit amount must not be nil")
	}

	if !v.IsPositive() {
		return fmt.Errorf("minimum initial deposit amount must be positive: %s", v)
	}

	return nil
}

func validateInitPoolCoinMintAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("initial pool coin mint amount must not be nil")
	}

	if !v.IsPositive() {
		return fmt.Errorf("initial pool coin mint amount must be positive: %s", v)
	}

	if v.LT(DefaultInitPoolCoinMintAmount) {
		return fmt.Errorf("initial pool coin mint amount must be greater than or equal to 1000000: %s", v)
	}

	return nil
}

func validateMaxReserveCoinAmount(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("max reserve coin amount must not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("max reserve coin amount must not be negative: %s", v)
	}

	return nil
}

func validateSwapFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("swap fee rate must not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("swap fee rate must not be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("swap fee rate too large: %s", v)
	}

	return nil
}

func validateWithdrawFeeRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("withdraw fee rate must not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("withdraw fee rate must not be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("withdraw fee rate too large: %s", v)
	}

	return nil
}

func validateMaxOrderAmountRatio(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("max order amount ratio must not be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("max order amount ratio must not be negative: %s", v)
	}

	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max order amount ratio too large: %s", v)
	}

	return nil
}

func validatePoolCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if err := v.Validate(); err != nil {
		return err
	}

	if v.Empty() {
		return fmt.Errorf("pool creation fee must not be empty")
	}

	return nil
}

func validateUnitBatchHeight(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("unit batch height must be positive: %d", v)
	}

	return nil
}

func validateCircuitBreakerEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
