package gaia

// DONTCOVER

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/keeper"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// GaiaApp testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

func setup(withGenesis bool, invCheckPeriod uint) (*GaiaApp, GenesisState) {
	db := dbm.NewMemDB()
	encCdc := MakeTestEncodingConfig()
	app := NewGaiaApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, invCheckPeriod, encCdc, EmptyAppOptions{})
	if withGenesis {
		return app, NewDefaultGenesisState()
	}
	return app, GenesisState{}
}

// Setup initializes a new GaiaApp. A Nop logger is set in GaiaApp.
func Setup(isCheckTx bool) *GaiaApp {
	app, genesisState := setup(!isCheckTx, 5)
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

type GenerateAccountStrategy func(int) []sdk.AccAddress

// createIncrementalAccounts is a strategy used by addTestAddrs() in order to generated addresses in ascending order.
func createIncrementalAccounts(accNum int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (accNum + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") // base address string

		buffer.WriteString(numString) // adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addr, _ := TestAddr(buffer.String(), bech)

		addresses = append(addresses, addr)
		buffer.Reset()
	}

	return addresses
}

// AddRandomTestAddr creates new account with random address.
func AddRandomTestAddr(app *GaiaApp, ctx sdk.Context, initCoins sdk.Coins) sdk.AccAddress {
	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	SaveAccount(app, ctx, addr, initCoins)
	return addr
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrs(app *GaiaApp, ctx sdk.Context, accNum int, initCoins sdk.Coins) []sdk.AccAddress {
	testAddrs := createIncrementalAccounts(accNum)
	for _, addr := range testAddrs {
		if err := FundAccount(app, ctx, addr, initCoins); err != nil {
			panic(err)
		}
	}
	return testAddrs
}

// permission of minting, create a "faucet" account. (@fdymylja)
func FundAccount(app *GaiaApp, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}
	return app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

// AddTestAddrs constructs and returns accNum amount of accounts with an
// initial balance of accAmt in random order
func AddTestAddrsIncremental(app *GaiaApp, ctx sdk.Context, accNum int, accAmt sdk.Int) []sdk.AccAddress {
	return addTestAddrs(app, ctx, accNum, accAmt, createIncrementalAccounts)
}

func addTestAddrs(app *GaiaApp, ctx sdk.Context, accNum int, accAmt sdk.Int, strategy GenerateAccountStrategy) []sdk.AccAddress {
	testAddrs := strategy(accNum)

	initCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), accAmt))

	for _, addr := range testAddrs {
		if err := FundAccount(app, ctx, addr, initCoins); err != nil {
			panic(err)
		}
	}

	return testAddrs
}

// SaveAccount saves the provided account into the simapp with balance based on initCoins.
func SaveAccount(app *GaiaApp, ctx sdk.Context, addr sdk.AccAddress, initCoins sdk.Coins) {
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)
	if initCoins.IsAllPositive() {
		err := FundAccount(app, ctx, addr, initCoins)
		if err != nil {
			panic(err)
		}
	}
}

func SaveAccountWithFee(app *GaiaApp, ctx sdk.Context, addr sdk.AccAddress, initCoins sdk.Coins, offerCoin sdk.Coin) {
	SaveAccount(app, ctx, addr, initCoins)
	params := app.LiquidityKeeper.GetParams(ctx)
	offerCoinFee := types.GetOfferCoinFee(offerCoin, params.SwapFeeRate)
	err := FundAccount(app, ctx, addr, sdk.NewCoins(offerCoinFee))
	if err != nil {
		panic(err)
	}
}

func TestAddr(addr string, bech string) (sdk.AccAddress, error) {
	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		return nil, err
	}
	bechexpected := res.String()
	if bech != bechexpected {
		return nil, fmt.Errorf("bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(bechres, res) {
		return nil, err
	}

	return res, nil
}

// CreateTestInput returns a simapp with custom LiquidityKeeper to avoid
// messing with the hooks.
func CreateTestInput() (*GaiaApp, sdk.Context) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)
	keeper.BatchLogicInvariantCheckFlag = true

	app := Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	appCodec := app.AppCodec()

	app.LiquidityKeeper = keeper.NewKeeper(
		appCodec,
		app.GetKey(types.StoreKey),
		app.GetSubspace(types.ModuleName),
		app.BankKeeper,
		app.AccountKeeper,
		app.DistrKeeper,
	)

	return app, ctx
}

func GetRandPoolAmt(r *rand.Rand, minInitDepositAmt sdk.Int) (x, y sdk.Int) {
	x = GetRandRange(r, int(minInitDepositAmt.Int64()), 100000000000000).MulRaw(int64(math.Pow10(r.Intn(10))))
	y = GetRandRange(r, int(minInitDepositAmt.Int64()), 100000000000000).MulRaw(int64(math.Pow10(r.Intn(10))))
	return
}

func GetRandRange(r *rand.Rand, min, max int) sdk.Int {
	return sdk.NewInt(int64(r.Intn(max-min) + min))
}

func GetRandomSizeOrders(denomX, denomY string, x, y sdk.Int, r *rand.Rand, sizeXToY, sizeYToX int32) (xToY, yToX []*types.MsgSwapWithinBatch) {
	randomSizeXtoY := int(r.Int31n(sizeXToY))
	randomSizeYtoX := int(r.Int31n(sizeYToX))
	return GetRandomOrders(denomX, denomY, x, y, r, randomSizeXtoY, randomSizeYtoX)
}

func GetRandomOrders(denomX, denomY string, x, y sdk.Int, r *rand.Rand, sizeXToY, sizeYToX int) (xToY, yToX []*types.MsgSwapWithinBatch) {
	currentPrice := x.ToDec().Quo(y.ToDec())

	for len(xToY) < sizeXToY {
		orderPrice := currentPrice.Mul(sdk.NewDecFromIntWithPrec(GetRandRange(r, 991, 1009), 3))
		orderAmt := sdk.ZeroDec() //nolint:staticcheck
		if r.Intn(2) == 1 {
			orderAmt = x.ToDec().Mul(sdk.NewDecFromIntWithPrec(GetRandRange(r, 1, 100), 4))
		} else {
			orderAmt = sdk.NewDecFromIntWithPrec(GetRandRange(r, 1000, 10000), 0)
		}
		if orderAmt.Quo(orderPrice).TruncateInt().IsZero() {
			continue
		}
		orderCoin := sdk.NewCoin(denomX, orderAmt.Ceil().TruncateInt())

		xToY = append(xToY, &types.MsgSwapWithinBatch{
			OfferCoin:       orderCoin,
			DemandCoinDenom: denomY,
			OrderPrice:      orderPrice,
		})
	}

	for len(yToX) < sizeYToX {
		orderPrice := currentPrice.Mul(sdk.NewDecFromIntWithPrec(GetRandRange(r, 991, 1009), 3))
		orderAmt := sdk.ZeroDec() //nolint:staticcheck
		if r.Intn(2) == 1 {
			orderAmt = y.ToDec().Mul(sdk.NewDecFromIntWithPrec(GetRandRange(r, 1, 100), 4))
		} else {
			orderAmt = sdk.NewDecFromIntWithPrec(GetRandRange(r, 1000, 10000), 0)
		}
		if orderAmt.Mul(orderPrice).TruncateInt().IsZero() {
			continue
		}
		orderCoin := sdk.NewCoin(denomY, orderAmt.Ceil().TruncateInt())

		yToX = append(yToX, &types.MsgSwapWithinBatch{
			OfferCoin:       orderCoin,
			DemandCoinDenom: denomX,
			OrderPrice:      orderPrice,
		})
	}
	return xToY, yToX
}

func TestCreatePool(t *testing.T, simapp *GaiaApp, ctx sdk.Context, x, y sdk.Int, denomX, denomY string, addr sdk.AccAddress) uint64 {
	deposit := sdk.NewCoins(sdk.NewCoin(denomX, x), sdk.NewCoin(denomY, y))
	params := simapp.LiquidityKeeper.GetParams(ctx)
	// set accounts for creator, depositor, withdrawer, balance for deposit
	SaveAccount(simapp, ctx, addr, deposit.Add(params.PoolCreationFee...)) // pool creator
	depositX := simapp.BankKeeper.GetBalance(ctx, addr, denomX)
	depositY := simapp.BankKeeper.GetBalance(ctx, addr, denomY)
	depositBalance := sdk.NewCoins(depositX, depositY)
	require.Equal(t, deposit, depositBalance)

	// create Liquidity pool
	poolTypeID := types.DefaultPoolTypeID
	poolID := simapp.LiquidityKeeper.GetNextPoolID(ctx)
	msg := types.NewMsgCreatePool(addr, poolTypeID, depositBalance)
	_, err := simapp.LiquidityKeeper.CreatePool(ctx, msg)
	require.NoError(t, err)

	// verify created liquidity pool
	pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
	require.True(t, found)
	require.Equal(t, poolID, pool.Id)
	require.Equal(t, denomX, pool.ReserveCoinDenoms[0])
	require.Equal(t, denomY, pool.ReserveCoinDenoms[1])

	// verify minted pool coin
	poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
	creatorBalance := simapp.BankKeeper.GetBalance(ctx, addr, pool.PoolCoinDenom)
	require.Equal(t, poolCoin, creatorBalance.Amount)
	return poolID
}

func TestDepositPool(t *testing.T, simapp *GaiaApp, ctx sdk.Context, x, y sdk.Int, addrs []sdk.AccAddress, poolID uint64, withEndblock bool) {
	pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
	require.True(t, found)
	denomX, denomY := pool.ReserveCoinDenoms[0], pool.ReserveCoinDenoms[1]
	deposit := sdk.NewCoins(sdk.NewCoin(denomX, x), sdk.NewCoin(denomY, y))

	moduleAccAddress := simapp.AccountKeeper.GetModuleAddress(types.ModuleName)
	moduleAccEscrowAmtX := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, denomX)
	moduleAccEscrowAmtY := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, denomY)
	iterNum := len(addrs)
	for i := 0; i < iterNum; i++ {
		SaveAccount(simapp, ctx, addrs[i], deposit) // pool creator

		depositMsg := types.NewMsgDepositWithinBatch(addrs[i], poolID, deposit)
		_, err := simapp.LiquidityKeeper.DepositWithinBatch(ctx, depositMsg)
		require.NoError(t, err)

		depositorBalanceX := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.ReserveCoinDenoms[0])
		depositorBalanceY := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.ReserveCoinDenoms[1])
		require.Equal(t, denomX, depositorBalanceX.Denom)
		require.Equal(t, denomY, depositorBalanceY.Denom)

		// check escrow balance of module account
		moduleAccEscrowAmtX = moduleAccEscrowAmtX.Add(deposit[0])
		moduleAccEscrowAmtY = moduleAccEscrowAmtY.Add(deposit[1])
		moduleAccEscrowAmtXAfter := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, denomX)
		moduleAccEscrowAmtYAfter := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, denomY)
		require.Equal(t, moduleAccEscrowAmtX, moduleAccEscrowAmtXAfter)
		require.Equal(t, moduleAccEscrowAmtY, moduleAccEscrowAmtYAfter)
	}
	batch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)
	require.True(t, found)

	// endblock
	if withEndblock {
		liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)
		msgs := simapp.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, batch)
		for i := 0; i < iterNum; i++ {
			// verify minted pool coin
			poolCoin := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
			depositorPoolCoinBalance := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.PoolCoinDenom)
			require.NotEqual(t, sdk.ZeroInt(), depositorPoolCoinBalance)
			require.NotEqual(t, sdk.ZeroInt(), poolCoin)

			require.True(t, msgs[i].Executed)
			require.True(t, msgs[i].Succeeded)
			require.True(t, msgs[i].ToBeDeleted)

			// error balance after endblock
			depositorBalanceX := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.ReserveCoinDenoms[0])
			depositorBalanceY := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.ReserveCoinDenoms[1])
			require.Equal(t, denomX, depositorBalanceX.Denom)
			require.Equal(t, denomY, depositorBalanceY.Denom)
		}
	}
}

func TestWithdrawPool(t *testing.T, simapp *GaiaApp, ctx sdk.Context, poolCoinAmt sdk.Int, addrs []sdk.AccAddress, poolID uint64, withEndblock bool) {
	pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
	require.True(t, found)
	moduleAccAddress := simapp.AccountKeeper.GetModuleAddress(types.ModuleName)
	moduleAccEscrowAmtPool := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, pool.PoolCoinDenom)

	iterNum := len(addrs)
	for i := 0; i < iterNum; i++ {
		balancePoolCoin := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.PoolCoinDenom)
		require.True(t, balancePoolCoin.Amount.GTE(poolCoinAmt))

		withdrawCoin := sdk.NewCoin(pool.PoolCoinDenom, poolCoinAmt)
		withdrawMsg := types.NewMsgWithdrawWithinBatch(addrs[i], poolID, withdrawCoin)
		_, err := simapp.LiquidityKeeper.WithdrawWithinBatch(ctx, withdrawMsg)
		require.NoError(t, err)

		moduleAccEscrowAmtPoolAfter := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, pool.PoolCoinDenom)
		moduleAccEscrowAmtPool.Amount = moduleAccEscrowAmtPool.Amount.Add(withdrawMsg.PoolCoin.Amount)
		require.Equal(t, moduleAccEscrowAmtPool, moduleAccEscrowAmtPoolAfter)

		balancePoolCoinAfter := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.PoolCoinDenom)
		if !balancePoolCoin.Amount.Equal(withdrawCoin.Amount) {
			require.Equal(t, balancePoolCoin.Sub(withdrawCoin).Amount, balancePoolCoinAfter.Amount)
		}

	}

	if withEndblock {
		poolCoinBefore := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)

		// endblock
		liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

		batch, found := simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)
		require.True(t, found)

		// verify burned pool coin
		poolCoinAfter := simapp.LiquidityKeeper.GetPoolCoinTotalSupply(ctx, pool)
		fmt.Println(poolCoinAfter, poolCoinBefore)
		require.True(t, poolCoinAfter.LT(poolCoinBefore))

		for i := 0; i < iterNum; i++ {
			withdrawerBalanceX := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.ReserveCoinDenoms[0])
			withdrawerBalanceY := simapp.BankKeeper.GetBalance(ctx, addrs[i], pool.ReserveCoinDenoms[1])
			require.True(t, withdrawerBalanceX.IsPositive())
			require.True(t, withdrawerBalanceY.IsPositive())

			withdrawMsgs := simapp.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, batch)
			require.True(t, withdrawMsgs[i].Executed)
			require.True(t, withdrawMsgs[i].Succeeded)
			require.True(t, withdrawMsgs[i].ToBeDeleted)
		}
	}
}

func TestSwapPool(t *testing.T, simapp *GaiaApp, ctx sdk.Context, offerCoins []sdk.Coin, orderPrices []sdk.Dec,
	addrs []sdk.AccAddress, poolID uint64, withEndblock bool,
) ([]*types.SwapMsgState, types.PoolBatch) {
	if len(offerCoins) != len(orderPrices) || len(orderPrices) != len(addrs) {
		require.True(t, false)
	}

	pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
	require.True(t, found)

	moduleAccAddress := simapp.AccountKeeper.GetModuleAddress(types.ModuleName)

	var swapMsgStates []*types.SwapMsgState

	params := simapp.LiquidityKeeper.GetParams(ctx)

	iterNum := len(addrs)
	for i := 0; i < iterNum; i++ {
		moduleAccEscrowAmtPool := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, offerCoins[i].Denom)
		currentBalance := simapp.BankKeeper.GetBalance(ctx, addrs[i], offerCoins[i].Denom)
		if currentBalance.IsLT(offerCoins[i]) {
			SaveAccountWithFee(simapp, ctx, addrs[i], sdk.NewCoins(offerCoins[i]), offerCoins[i])
		}
		var demandCoinDenom string
		switch offerCoins[i].Denom {
		case pool.ReserveCoinDenoms[0]:
			demandCoinDenom = pool.ReserveCoinDenoms[1]
		case pool.ReserveCoinDenoms[1]:
			demandCoinDenom = pool.ReserveCoinDenoms[0]
		default:
			require.True(t, false)
		}

		swapMsg := types.NewMsgSwapWithinBatch(addrs[i], poolID, types.DefaultSwapTypeID, offerCoins[i], demandCoinDenom, orderPrices[i], params.SwapFeeRate)
		batchPoolSwapMsg, err := simapp.LiquidityKeeper.SwapWithinBatch(ctx, swapMsg, 0)
		require.NoError(t, err)

		swapMsgStates = append(swapMsgStates, batchPoolSwapMsg)
		moduleAccEscrowAmtPoolAfter := simapp.BankKeeper.GetBalance(ctx, moduleAccAddress, offerCoins[i].Denom)
		moduleAccEscrowAmtPool.Amount = moduleAccEscrowAmtPool.Amount.Add(offerCoins[i].Amount).Add(types.GetOfferCoinFee(offerCoins[i], params.SwapFeeRate).Amount)
		require.Equal(t, moduleAccEscrowAmtPool, moduleAccEscrowAmtPoolAfter)

	}
	batch, _ := simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)

	if withEndblock {
		// endblock
		liquidity.EndBlocker(ctx, simapp.LiquidityKeeper)

		batch, found = simapp.LiquidityKeeper.GetPoolBatch(ctx, poolID)
		require.True(t, found)
	}
	return swapMsgStates, batch
}

func GetSwapMsg(t *testing.T, simapp *GaiaApp, ctx sdk.Context, offerCoins []sdk.Coin, orderPrices []sdk.Dec,
	addrs []sdk.AccAddress, poolID uint64,
) []*types.MsgSwapWithinBatch {
	t.Helper()
	if len(offerCoins) != len(orderPrices) || len(orderPrices) != len(addrs) {
		require.True(t, false)
	}

	var msgs []*types.MsgSwapWithinBatch
	pool, found := simapp.LiquidityKeeper.GetPool(ctx, poolID)
	require.True(t, found)

	params := simapp.LiquidityKeeper.GetParams(ctx)

	iterNum := len(addrs)
	for i := 0; i < iterNum; i++ {
		currentBalance := simapp.BankKeeper.GetBalance(ctx, addrs[i], offerCoins[i].Denom)
		if currentBalance.IsLT(offerCoins[i]) {
			SaveAccountWithFee(simapp, ctx, addrs[i], sdk.NewCoins(offerCoins[i]), offerCoins[i])
		}
		var demandCoinDenom string
		switch offerCoins[i].Denom {
		case pool.ReserveCoinDenoms[0]:
			demandCoinDenom = pool.ReserveCoinDenoms[1]
		case pool.ReserveCoinDenoms[1]:
			demandCoinDenom = pool.ReserveCoinDenoms[0]
		default:
			require.True(t, false)
		}

		msgs = append(msgs, types.NewMsgSwapWithinBatch(addrs[i], poolID, types.DefaultSwapTypeID, offerCoins[i], demandCoinDenom, orderPrices[i], params.SwapFeeRate))
	}
	return msgs
}
