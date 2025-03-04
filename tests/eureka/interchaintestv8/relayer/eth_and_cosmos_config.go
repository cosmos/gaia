package relayer

import (
	"os"
	"text/template"
)

// EthCosmosConfigInfo is a struct that holds the configuration information for the Eth to Cosmos config template
type EthCosmosConfigInfo struct {
	// Ethereum chain identifier
	EthChainID string
	// Cosmos chain identifier
	CosmosChainID string
	// Tendermint RPC URL
	TmRPC string
	// ICS26 Router address
	ICS26Address string
	// Ethereum RPC URL
	EthRPC string
	// Ethereum Beacon API URL
	BeaconAPI string
	// SP1 config, "mock" or "env"
	SP1Config string
	// Signer address cosmos
	SignerAddress string
	// Whether we use the mock client in Cosmos
	MockWasmClient bool
}

// GenerateEthCosmosConfigFile generates an eth to cosmos config file from the template.
func (c *EthCosmosConfigInfo) GenerateEthCosmosConfigFile(path string) error {
	tmpl, err := template.ParseFiles("e2e/interchaintestv8/relayer/config.tmpl")
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()
	return tmpl.Execute(f, c)
}

// DefaultRelayerGRPCAddress returns the default gRPC address for the relayer.
func DefaultRelayerGRPCAddress() string {
	return "127.0.0.1:3000"
}
