package relayer

import (
	"os"
	"text/template"
)

// CosmosToCosmosConfigInfo is a struct that holds the configuration information for the Cosmos to Cosmos config template
type CosmosToCosmosConfigInfo struct {
	// Chain A chain identifier
	ChainAID string
	// Chain B chain identifier
	ChainBID string
	// ChainA Tendermint RPC URL
	ChainATmRPC string
	// ChainB Tendermint RPC URL
	ChainBTmRPC string
	// ChainA Submitter address
	ChainAUser string
	// ChainB Submitter address
	ChainBUser string
}

// GenerateCosmosToCosmosConfigFile generates a cosmos to cosmos config file from the template.
func (c *CosmosToCosmosConfigInfo) GenerateCosmosToCosmosConfigFile(path string) error {
	tmpl, err := template.ParseFiles("e2e/interchaintestv8/relayer/cosmos_to_cosmos_config.tmpl")
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
