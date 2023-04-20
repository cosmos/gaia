package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// Simulation parameter constants
const (
	LiquidityPoolTypes     = "liquidity_pool_types"
	MinInitDepositAmount   = "min_init_deposit_amount"
	InitPoolCoinMintAmount = "init_pool_coin_mint_amount"
	MaxReserveCoinAmount   = "max_reserve_coin_amount"
	PoolCreationFee        = "pool_creation_fee"
	SwapFeeRate            = "swap_fee_rate"
	WithdrawFeeRate        = "withdraw_fee_rate"
	MaxOrderAmountRatio    = "max_order_amount_ratio"
	UnitBatchHeight        = "unit_batch_height"
)

// GenLiquidityPoolTypes return default PoolType temporarily, It will be randomized in the liquidity v2
func GenLiquidityPoolTypes(r *rand.Rand) (liquidityPoolTypes []types.PoolType) {
	return types.DefaultPoolTypes
}

// GenMinInitDepositAmount randomized MinInitDepositAmount
func GenMinInitDepositAmount(r *rand.Rand) sdk.Int {
	return sdk.NewInt(int64(simulation.RandIntBetween(r, int(types.DefaultMinInitDepositAmount.Int64()), 1e7)))
}

// GenInitPoolCoinMintAmount randomized InitPoolCoinMintAmount
func GenInitPoolCoinMintAmount(r *rand.Rand) sdk.Int {
	return sdk.NewInt(int64(simulation.RandIntBetween(r, int(types.DefaultInitPoolCoinMintAmount.Int64()), 1e8)))
}

// GenMaxReserveCoinAmount randomized MaxReserveCoinAmount
func GenMaxReserveCoinAmount(r *rand.Rand) sdk.Int {
	return sdk.NewInt(int64(simulation.RandIntBetween(r, int(types.DefaultMaxReserveCoinAmount.Int64()), 1e13)))
}

// GenPoolCreationFee randomized PoolCreationFee
// list of 1 to 4 coins with an amount greater than 1
func GenPoolCreationFee(r *rand.Rand) sdk.Coins {
	var coins sdk.Coins
	var denoms []string

	count := simulation.RandIntBetween(r, 1, 4)
	for i := 0; i < count; i++ {
		randomDenom := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 4, 6))
		denoms = append(denoms, strings.ToLower(randomDenom))
	}

	sortedDenoms := types.SortDenoms(denoms)

	for i := 0; i < count; i++ {
		randomCoin := sdk.NewCoin(sortedDenoms[i], sdk.NewInt(int64(simulation.RandIntBetween(r, 1e6, 1e7))))
		coins = append(coins, randomCoin)
	}

	return coins
}

// GenSwapFeeRate randomized SwapFeeRate ranging from 0.00001 to 1
func GenSwapFeeRate(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 1, 1e5)), 5)
}

// GenWithdrawFeeRate randomized WithdrawFeeRate ranging from 0.00001 to 1
func GenWithdrawFeeRate(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 1, 1e5)), 5)
}

// GenMaxOrderAmountRatio randomized MaxOrderAmountRatio ranging from 0.00001 to 1
func GenMaxOrderAmountRatio(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 1, 1e5)), 5)
}

// GenUnitBatchHeight randomized UnitBatchHeight ranging from 1 to 20
func GenUnitBatchHeight(r *rand.Rand) uint32 {
	return uint32(simulation.RandIntBetween(r, int(types.DefaultUnitBatchHeight), 20))
}

// RandomizedGenState generates a random GenesisState for liquidity
func RandomizedGenState(simState *module.SimulationState) {
	var liquidityPoolTypes []types.PoolType
	simState.AppParams.GetOrGenerate(
		simState.Cdc, LiquidityPoolTypes, &liquidityPoolTypes, simState.Rand,
		func(r *rand.Rand) { liquidityPoolTypes = GenLiquidityPoolTypes(r) },
	)

	var minInitDepositAmount sdk.Int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MinInitDepositAmount, &minInitDepositAmount, simState.Rand,
		func(r *rand.Rand) { minInitDepositAmount = GenMinInitDepositAmount(r) },
	)

	var initPoolCoinMintAmount sdk.Int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InitPoolCoinMintAmount, &initPoolCoinMintAmount, simState.Rand,
		func(r *rand.Rand) { initPoolCoinMintAmount = GenInitPoolCoinMintAmount(r) },
	)

	var maxReserveCoinAmount sdk.Int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxReserveCoinAmount, &maxReserveCoinAmount, simState.Rand,
		func(r *rand.Rand) { maxReserveCoinAmount = GenMaxReserveCoinAmount(r) },
	)

	var poolCreationFee sdk.Coins
	simState.AppParams.GetOrGenerate(
		simState.Cdc, PoolCreationFee, &poolCreationFee, simState.Rand,
		func(r *rand.Rand) { poolCreationFee = GenPoolCreationFee(r) },
	)

	var swapFeeRate sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SwapFeeRate, &swapFeeRate, simState.Rand,
		func(r *rand.Rand) { swapFeeRate = GenSwapFeeRate(r) },
	)

	var withdrawFeeRate sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, WithdrawFeeRate, &withdrawFeeRate, simState.Rand,
		func(r *rand.Rand) { withdrawFeeRate = GenWithdrawFeeRate(r) },
	)

	var maxOrderAmountRatio sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxOrderAmountRatio, &maxOrderAmountRatio, simState.Rand,
		func(r *rand.Rand) { maxOrderAmountRatio = GenMaxOrderAmountRatio(r) },
	)

	var unitBatchHeight uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, UnitBatchHeight, &unitBatchHeight, simState.Rand,
		func(r *rand.Rand) { unitBatchHeight = GenUnitBatchHeight(r) },
	)

	liquidityGenesis := types.GenesisState{
		Params: types.Params{
			PoolTypes:              liquidityPoolTypes,
			MinInitDepositAmount:   minInitDepositAmount,
			InitPoolCoinMintAmount: initPoolCoinMintAmount,
			MaxReserveCoinAmount:   maxReserveCoinAmount,
			PoolCreationFee:        poolCreationFee,
			SwapFeeRate:            swapFeeRate,
			WithdrawFeeRate:        withdrawFeeRate,
			MaxOrderAmountRatio:    maxOrderAmountRatio,
			UnitBatchHeight:        unitBatchHeight,
		},
		PoolRecords: []types.PoolRecord{},
	}

	bz, _ := json.MarshalIndent(&liquidityGenesis, "", " ")
	fmt.Printf("Selected randomly generated liquidity parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&liquidityGenesis)
}
