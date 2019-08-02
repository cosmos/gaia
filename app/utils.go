//nolint
package app

import (
	"io"

	dbm "github.com/tendermint/tm-db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// DONTCOVER

// NewGaiaAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewGaiaAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (gapp *GaiaApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewGaiaApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keys[baseapp.MainStoreKey], gapp.keys[staking.StoreKey], gapp.stakingKeeper
}
