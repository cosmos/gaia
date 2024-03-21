package gaia

import (
	"cosmossdk.io/math"

	signerextraction "github.com/skip-mev/block-sdk/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/block/base"
	defaultlane "github.com/skip-mev/block-sdk/lanes/base"
)

// Creating the lanes for the block sdk.
func CreateLanes(app *GaiaApp) *base.BaseLane {
	// Create the signer extractor. This is used to extract the expected signers from
	// a transaction. Each lane can have a different signer extractor if needed.
	signerAdapter := signerextraction.NewDefaultAdapter()

	// Create the configurations for each lane. These configurations determine how many
	// transactions the lane can store, the maximum block space the lane can consume, and
	// the signer extractor used to extract the expected signers from a transaction.

	// Create a default configuration that accepts 1000 transactions and consumes 100% of the
	// block space, since the default lane is currently the only one in our app
	defaultConfig := base.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.txConfig.TxEncoder(),
		TxDecoder:       app.txConfig.TxDecoder(),
		MaxBlockSpace:   math.LegacyOneDec(),
		SignerExtractor: signerAdapter,
		MaxTxs:          1000,
	}

	// Create the match handlers for each lane. These match handlers determine whether or not
	// a transaction belongs in the lane.

	// Create the final match handler for the default lane.
	defaultMatchHandler := base.DefaultMatchHandler()

	// Create the lanes.
	defaultLane := defaultlane.NewDefaultLane(
		defaultConfig,
		defaultMatchHandler,
	)

	return defaultLane
}
