package globalfee

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	store"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/gaia/v8/x/globalfee/types"
)

func TestGlobalMinimumChainFeeAnteHandler(t *testing.T) {
	specs := map[string]struct {
		setupStore func(ctx sdk.Context, s paramtypes.Subspace)
		next       sdk.AnteDecorator
		feeAmount  sdk.Coins
		gasLimit   sdk.Gas
		expErr     *sdkerrors.Error
	}{
		"single fee above min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(2))),
			gasLimit:  1,
		},
		"single fee below min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(2))),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1))),
			gasLimit:  1,
			expErr:    sdkerrors.ErrInsufficientFee,
		},
		"single fee equal min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.OneInt())),
			gasLimit:  1,
		},
		"multiple fees both above min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(2)), sdk.NewCoin("BLX", sdk.NewInt(2))),
			gasLimit:  1,
		},
		"multiple fees both below min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(2)), sdk.NewDecCoin("BLX", sdk.NewInt(2))),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1)), sdk.NewCoin("BLX", sdk.NewInt(1))),
			gasLimit:  1,
			expErr:    sdkerrors.ErrInsufficientFee,
		},
		"multiple fees both equal min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1)), sdk.NewCoin("BLX", sdk.NewInt(1))),
			gasLimit:  1,
		},
		"multiple fees one below min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(2)), sdk.NewDecCoin("BLX", sdk.NewInt(2))),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1)), sdk.NewCoin("BLX", sdk.NewInt(3))),
			gasLimit:  1,
		},
		"multiple fees one equal min": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(2)), sdk.NewDecCoin("BLX", sdk.NewInt(2))),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1)), sdk.NewCoin("BLX", sdk.NewInt(2))),
			gasLimit:  1,
		},
		"multiple fees one submitted": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1))),
			gasLimit:  1,
		},
		"multiple fees with non fee token added": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(1)), sdk.NewCoin("CLX", sdk.NewInt(2))),
			gasLimit:  1,
		},
		"multiple fees with only non fee token": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt()), sdk.NewDecCoin("BLX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("CLX", sdk.NewInt(2))),
			gasLimit:  1,
			expErr:    sdkerrors.ErrInsufficientFee,
		},
		"no min gas price set": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{})
			},
			gasLimit: 1,
		},
		"no param set": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
			},
			gasLimit: 1,
		},
		"no gas set": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt())),
				})
			},
			feeAmount: sdk.NewCoins(sdk.NewCoin("ALX", sdk.NewInt(2))),
			expErr:    sdkerrors.ErrInsufficientFee, // error message is a bit odd: "got: 2ALX required: 0ALX: insufficient fee"
		},
		"no fee set": {
			setupStore: func(ctx sdk.Context, s paramtypes.Subspace) {
				s.SetParamSet(ctx, &types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.OneInt())),
				})
			},
			gasLimit: 1,
			expErr:   sdkerrors.ErrInsufficientFee,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, encCfg, subspace := setupTestStore(t)
			spec.setupStore(ctx, subspace)

			txBuilder := encCfg.TxConfig.NewTxBuilder()
			txBuilder.SetFeeAmount(spec.feeAmount)
			txBuilder.SetGasLimit(spec.gasLimit)
			tx := txBuilder.GetTx()
			captured := &CapturingAnteHandler{}
			anteHandler := sdk.ChainAnteDecorators(
				NewGlobalMinimumChainFeeDecorator(subspace),
				captured,
			)
			_, gotErr := anteHandler(ctx, tx, false)
			require.True(t, spec.expErr.Is(gotErr), "exp : %s but got %#+v", spec.expErr, gotErr)
			if spec.expErr != nil {
				require.Empty(t, captured.txs)
				return
			}
			assert.Equal(t, []sdk.Tx{tx}, captured.txs)
		})
	}
}

func setupTestStore(t *testing.T) (sdk.Context, simappparams.EncodingConfig, paramtypes.Subspace) {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	encCfg := simapp.MakeTestEncodingConfig()
	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	ms.MountStoreWithDB(keyParams, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, storetypes.StoreTypeTransient, db)
	require.NoError(t, ms.LoadLatestVersion())

	paramsKeeper := paramskeeper.NewKeeper(encCfg.Codec, encCfg.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	subspace := paramsKeeper.Subspace(ModuleName).WithKeyTable(types.ParamKeyTable())
	return ctx, encCfg, subspace
}

type CapturingAnteHandler struct {
	txs []sdk.Tx
}

func (c *CapturingAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	c.txs = append(c.txs, tx)
	return next(ctx, tx, simulate)
}

func MockResultAnteHandler(result error) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, err
	}
}
