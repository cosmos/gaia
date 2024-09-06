package v20_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	providertypes "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"

	"github.com/cosmos/gaia/v20/app/helpers"
	v20 "github.com/cosmos/gaia/v20/app/upgrades/v20"
)

func TestSetICSConsumerMetadata(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	pk := gaiaApp.ProviderKeeper

	// Add consumer chains
	neutronConsumerId := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, neutronConsumerId, "neutron-1")
	pk.SetConsumerPhase(ctx, neutronConsumerId, providertypes.CONSUMER_PHASE_LAUNCHED)
	strideConsumerId := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, strideConsumerId, "stride-1")
	pk.SetConsumerPhase(ctx, strideConsumerId, providertypes.CONSUMER_PHASE_LAUNCHED)

	err := v20.SetICSConsumerMetadata(ctx, pk)
	require.NoError(t, err)

	metadata, err := pk.GetConsumerMetadata(ctx, neutronConsumerId)
	require.NoError(t, err)
	require.Equal(t, "Neutron", metadata.Name)
	expectedMetadataField := map[string]string{
		"phase":          "mainnet",
		"forge_json_url": "https://raw.githubusercontent.com/neutron-org/neutron/main/forge.json",
	}
	metadataField := map[string]string{}
	err = json.Unmarshal([]byte(metadata.Metadata), &metadataField)
	require.NoError(t, err)
	require.Equal(t, expectedMetadataField, metadataField)

	metadata, err = pk.GetConsumerMetadata(ctx, strideConsumerId)
	require.NoError(t, err)
	require.Equal(t, "Stride", metadata.Name)
	expectedMetadataField = map[string]string{
		"phase":          "mainnet",
		"forge_json_url": "https://raw.githubusercontent.com/Stride-Labs/stride/main/forge.json",
	}
	metadataField = map[string]string{}
	err = json.Unmarshal([]byte(metadata.Metadata), &metadataField)
	require.NoError(t, err)
	require.Equal(t, expectedMetadataField, metadataField)
}
