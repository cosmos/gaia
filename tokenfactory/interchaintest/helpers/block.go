package helpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

func GetBlockData(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, height uint64) BlockData {
	var res BlockData
	ExecuteQuery(ctx, chain, []string{"query", "block", "--type=height", fmt.Sprintf("%d", height)}, &res)
	return res
}

// SDK v50, not sure what the sdk.(type) is
type BlockData struct {
	Header struct {
		Version struct {
			Block string `json:"block"`
			App   string `json:"app"`
		} `json:"version"`
		ChainID     string `json:"chain_id"`
		Height      string `json:"height"`
		Time        string `json:"time"`
		LastBlockID struct {
			Hash          string `json:"hash"`
			PartSetHeader struct {
				Total int    `json:"total"`
				Hash  string `json:"hash"`
			} `json:"part_set_header"`
		} `json:"last_block_id"`
		LastCommitHash     string `json:"last_commit_hash"`
		DataHash           string `json:"data_hash"`
		ValidatorsHash     string `json:"validators_hash"`
		NextValidatorsHash string `json:"next_validators_hash"`
		ConsensusHash      string `json:"consensus_hash"`
		AppHash            string `json:"app_hash"`
		LastResultsHash    string `json:"last_results_hash"`
		EvidenceHash       string `json:"evidence_hash"`
		ProposerAddress    string `json:"proposer_address"`
	} `json:"header"`
	Data struct {
		Txs []string `json:"txs"`
	} `json:"data"`
	Evidence struct {
		Evidence []any `json:"evidence"`
	} `json:"evidence"`
	LastCommit struct {
		Height  string `json:"height"`
		Round   int    `json:"round"`
		BlockID struct {
			Hash          string `json:"hash"`
			PartSetHeader struct {
				Total int    `json:"total"`
				Hash  string `json:"hash"`
			} `json:"part_set_header"`
		} `json:"block_id"`
		Signatures []struct {
			BlockIDFlag      string `json:"block_id_flag"`
			ValidatorAddress string `json:"validator_address"`
			Timestamp        string `json:"timestamp"`
			Signature        string `json:"signature"`
		} `json:"signatures"`
	} `json:"last_commit"`
}
