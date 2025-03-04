package ethereum

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"

	ethereumtypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/ethereum"
)

type Spec struct {
	SecondsPerSlot               time.Duration `json:"SECONDS_PER_SLOT"`
	SlotsPerEpoch                uint64        `json:"SLOTS_PER_EPOCH"`
	EpochsPerSyncCommitteePeriod uint64        `json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD"`

	// Fork Parameters
	GenesisForkVersion   phase0.Version `json:"GENESIS_FORK_VERSION"`
	GenesisSlot          uint64         `json:"GENESIS_SLOT"`
	AltairForkVersion    phase0.Version `json:"ALTAIR_FORK_VERSION"`
	AltairForkEpoch      uint64         `json:"ALTAIR_FORK_EPOCH"`
	BellatrixForkVersion phase0.Version `json:"BELLATRIX_FORK_VERSION"`
	BellatrixForkEpoch   uint64         `json:"BELLATRIX_FORK_EPOCH"`
	CapellaForkVersion   phase0.Version `json:"CAPELLA_FORK_VERSION"`
	CapellaForkEpoch     uint64         `json:"CAPELLA_FORK_EPOCH"`
	DenebForkVersion     phase0.Version `json:"DENEB_FORK_VERSION"`
	DenebForkEpoch       uint64         `json:"DENEB_FORK_EPOCH"`
}

type Bootstrap struct {
	Data struct {
		Header               BootstrapHeader `json:"header"`
		CurrentSyncCommittee SyncCommittee   `json:"current_sync_committee"`
	} `json:"data"`
}

type BootstrapHeader struct {
	Beacon    BeaconJSON    `json:"beacon"`
	Execution ExecutionJSON `json:"execution"`
}

type SyncCommittee struct {
	Pubkeys         []string `json:"pubkeys"`
	AggregatePubkey string   `json:"aggregate_pubkey"`
}

type LightClientUpdatesResponse []LightClientUpdateJSON

type BeaconJSON struct {
	Slot          uint64 `json:"slot,string"`
	ProposerIndex uint64 `json:"proposer_index,string"`
	ParentRoot    string `json:"parent_root"`
	StateRoot     string `json:"state_root"`
	BodyRoot      string `json:"body_root"`
}

type ExecutionJSON struct {
	ParentHash       string `json:"parent_hash"`
	FeeRecipient     string `json:"fee_recipient"`
	StateRoot        string `json:"state_root"`
	ReceiptsRoot     string `json:"receipts_root"`
	LogsBloom        string `json:"logs_bloom"`
	PrevRandao       string `json:"prev_randao"`
	BlockNumber      uint64 `json:"block_number,string"`
	GasLimit         uint64 `json:"gas_limit,string"`
	GasUsed          uint64 `json:"gas_used,string"`
	Timestamp        uint64 `json:"timestamp,string"`
	ExtraData        string `json:"extra_data"`
	BaseFeePerGas    uint64 `json:"base_fee_per_gas,string"`
	BlockHash        string `json:"block_hash"`
	TransactionsRoot string `json:"transactions_root"`
	WithdrawalsRoot  string `json:"withdrawals_root"`
	BlobGasUsed      uint64 `json:"blob_gas_used,string"`
	ExcessBlobGas    uint64 `json:"excess_blob_gas,string"`
}

type FinalityUpdateJSONResponse struct {
	Version string                          `json:"version"`
	Data    ethereumtypes.LightClientUpdate `json:"data"`
}

type BeaconBlocksResponseJSON struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			Body          struct {
				RandaoReveal string `json:"randao_reveal"`
				Eth1Data     struct {
					DepositRoot  string `json:"deposit_root"`
					DepositCount string `json:"deposit_count"`
					BlockHash    string `json:"block_hash"`
				} `json:"eth1_data"`
				Graffiti          string `json:"graffiti"`
				ProposerSlashings []any  `json:"proposer_slashings"`
				AttesterSlashings []any  `json:"attester_slashings"`
				Attestations      []any  `json:"attestations"`
				Deposits          []any  `json:"deposits"`
				VoluntaryExits    []any  `json:"voluntary_exits"`
				SyncAggregate     struct {
					SyncCommitteeBits      string `json:"sync_committee_bits"`
					SyncCommitteeSignature string `json:"sync_committee_signature"`
				} `json:"sync_aggregate"`
				ExecutionPayload struct {
					ParentHash    string `json:"parent_hash"`
					FeeRecipient  string `json:"fee_recipient"`
					StateRoot     string `json:"state_root"`
					ReceiptsRoot  string `json:"receipts_root"`
					LogsBloom     string `json:"logs_bloom"`
					PrevRandao    string `json:"prev_randao"`
					BlockNumber   uint64 `json:"block_number,string"`
					GasLimit      string `json:"gas_limit"`
					GasUsed       string `json:"gas_used"`
					Timestamp     string `json:"timestamp"`
					ExtraData     string `json:"extra_data"`
					BaseFeePerGas string `json:"base_fee_per_gas"`
					BlockHash     string `json:"block_hash"`
					Transactions  []any  `json:"transactions"`
					Withdrawals   []any  `json:"withdrawals"`
					BlobGasUsed   string `json:"blob_gas_used"`
					ExcessBlobGas string `json:"excess_blob_gas"`
				} `json:"execution_payload"`
				BlsToExecutionChanges []any `json:"bls_to_execution_changes"`
				BlobKzgCommitments    []any `json:"blob_kzg_commitments"`
			} `json:"body"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"data"`
}

type LightClientUpdateJSON struct {
	Data ethereumtypes.LightClientUpdate `json:"data"`
}
