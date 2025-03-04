package ethereum

// The Ethereum types to export to JSON Schema and used to generate types in other languages
// The structure itself is not used for anything else than being able to generate a single
// schema
type GeneratedTypes struct {
	ClientState    ClientState    `json:"client_state"`
	ConsensusState ConsensusState `json:"consensus_state"`
	Header         Header         `json:"header"`
	StorageProof   StorageProof   `json:"storage_proof"`
	TestFixtures   TestFixtures   `json:"test_fixtures"`
}

// The ethereum client state
//
// # The client state at the time of the proof
//
// # The client state at the initial state
//
// The client state after the update
type ClientState struct {
	// The chain ID
	ChainID uint64 `json:"chain_id"`
	// The number of epochs per sync committee period
	EpochsPerSyncCommitteePeriod uint64 `json:"epochs_per_sync_committee_period"`
	// The fork parameters
	ForkParameters ForkParameters `json:"fork_parameters"`
	// The time of genesis (unix timestamp)
	GenesisTime uint64 `json:"genesis_time"`
	// The genesis validators root
	GenesisValidatorsRoot string `json:"genesis_validators_root"`
	// The storage slot of the IBC commitment in the Ethereum contract
	IbcCommitmentSlot string `json:"ibc_commitment_slot"`
	// The address of the IBC contract being tracked on Ethereum
	IbcContractAddress string `json:"ibc_contract_address"`
	// Whether the client is frozen
	IsFrozen bool `json:"is_frozen"`
	// The latest slot of this client
	LatestSlot uint64 `json:"latest_slot"`
	// The minimum number of participants in the sync committee
	MinSyncCommitteeParticipants uint64 `json:"min_sync_committee_participants"`
	// The slot duration in seconds
	SecondsPerSlot uint64 `json:"seconds_per_slot"`
	// The number of slots per epoch
	SlotsPerEpoch uint64 `json:"slots_per_epoch"`
}

// The fork parameters
type ForkParameters struct {
	// The altair fork
	Altair Fork `json:"altair"`
	// The bellatrix fork
	Bellatrix Fork `json:"bellatrix"`
	// The capella fork
	Capella Fork `json:"capella"`
	// The deneb fork
	Deneb Fork `json:"deneb"`
	// The genesis fork version
	GenesisForkVersion string `json:"genesis_fork_version"`
	// The genesis slot
	GenesisSlot uint64 `json:"genesis_slot"`
}

// The altair fork
//
// # The fork data
//
// # The bellatrix fork
//
// # The capella fork
//
// The deneb fork
type Fork struct {
	// The epoch at which this fork is activated
	Epoch uint64 `json:"epoch"`
	// The version of the fork
	Version string `json:"version"`
}

// The consensus state of the Ethereum light client
//
// # The consensus state at the time of the proof
//
// # The consensus state at the initial state
//
// The consensus state after the update
type ConsensusState struct {
	// aggregate public key of current sync committee
	CurrentSyncCommittee string `json:"current_sync_committee"`
	// aggregate public key of next sync committee
	NextSyncCommittee string `json:"next_sync_committee"`
	// The slot number
	Slot uint64 `json:"slot"`
	// The state merkle root
	StateRoot string `json:"state_root"`
	// The storage merkle root
	StorageRoot string `json:"storage_root"`
	// The unix timestamp at the time of the slot. It is calculated from the genesis time and
	// slots per.
	Timestamp uint64 `json:"timestamp"`
}

// The header of a light client update
type Header struct {
	// The account update
	AccountUpdate AccountUpdate `json:"account_update"`
	// The consensus update
	ConsensusUpdate LightClientUpdate `json:"consensus_update"`
	// The trusted sync committee
	TrustedSyncCommittee TrustedSyncCommittee `json:"trusted_sync_committee"`
}

// The account update
type AccountUpdate struct {
	// The account proof
	AccountProof AccountProof `json:"account_proof"`
}

// The account proof
type AccountProof struct {
	// The account proof
	Proof []string `json:"proof"`
	// The account storage root
	StorageRoot string `json:"storage_root"`
}

// The consensus update
//
// A light client update
type LightClientUpdate struct {
	// Header attested to by the sync committee
	AttestedHeader LightClientHeader `json:"attested_header"`
	// Branch of the finalized header
	FinalityBranch []string `json:"finality_branch"`
	// Finalized header corresponding to `attested_header.state_root`
	FinalizedHeader LightClientHeader `json:"finalized_header"`
	// Next sync committee corresponding to `attested_header.state_root`
	NextSyncCommittee *SyncCommittee `json:"next_sync_committee"`
	// The branch of the next sync committee
	NextSyncCommitteeBranch []string `json:"next_sync_committee_branch"`
	// Slot at which the aggregate signature was created (untrusted)
	SignatureSlot string `json:"signature_slot"`
	// Sync committee aggregate signature
	SyncAggregate SyncAggregate `json:"sync_aggregate"`
}

// Header attested to by the sync committee
//
// # The header of a light client
//
// Finalized header corresponding to `attested_header.state_root`
type LightClientHeader struct {
	// The beacon block header
	Beacon BeaconBlockHeader `json:"beacon"`
	// The execution payload header
	Execution ExecutionPayloadHeader `json:"execution"`
	// The execution branch
	ExecutionBranch []string `json:"execution_branch"`
}

// The beacon block header
type BeaconBlockHeader struct {
	// The tree hash merkle root of the `BeaconBlockBody` for the `BeaconBlock`
	BodyRoot string `json:"body_root"`
	// The signing merkle root of the parent `BeaconBlock`
	ParentRoot string `json:"parent_root"`
	// The index of validator in validator registry
	ProposerIndex string `json:"proposer_index"`
	// The slot to which this block corresponds
	Slot string `json:"slot"`
	// The tree hash merkle root of the `BeaconState` for the `BeaconBlock`
	StateRoot string `json:"state_root"`
}

// The execution payload header
//
// Header to track the execution block
type ExecutionPayloadHeader struct {
	// Block base fee per gas
	BaseFeePerGas string `json:"base_fee_per_gas"`
	// Blob gas used (new in Deneb)
	BlobGasUsed string `json:"blob_gas_used"`
	// The block hash
	BlockHash string `json:"block_hash"`
	// The block number of the execution payload
	BlockNumber string `json:"block_number"`
	// Excess blob gas (new in Deneb)
	ExcessBlobGas string `json:"excess_blob_gas"`
	// The extra data of the execution payload
	ExtraData string `json:"extra_data"`
	// Block fee recipient
	FeeRecipient string `json:"fee_recipient"`
	// Execution block gas limit
	GasLimit string `json:"gas_limit"`
	// Execution block gas used
	GasUsed string `json:"gas_used"`
	// The logs bloom filter
	LogsBloom string `json:"logs_bloom"`
	// The parent hash of the execution payload header
	ParentHash string `json:"parent_hash"`
	// The previous Randao value, used to compute the randomness on the execution layer.
	PrevRandao string `json:"prev_randao"`
	// The root of the receipts trie
	ReceiptsRoot string `json:"receipts_root"`
	// The state root
	StateRoot string `json:"state_root"`
	// The timestamp of the execution payload
	Timestamp string `json:"timestamp"`
	// SSZ hash tree root of the transaction list
	TransactionsRoot string `json:"transactions_root"`
	// Tree root of the withdrawals list
	WithdrawalsRoot string `json:"withdrawals_root"`
}

// The sync committee data
type SyncCommittee struct {
	// The aggregate public key of the sync committee
	AggregatePubkey string `json:"aggregate_pubkey"`
	// The public keys of the sync committee
	Pubkeys []string `json:"pubkeys"`
}

// Sync committee aggregate signature
//
// The sync committee aggregate
type SyncAggregate struct {
	// The bits representing the sync committee's participation.
	SyncCommitteeBits string `json:"sync_committee_bits"`
	// The aggregated signature of the sync committee.
	SyncCommitteeSignature string `json:"sync_committee_signature"`
}

// The trusted sync committee
type TrustedSyncCommittee struct {
	// The current sync committee
	SyncCommittee ActiveSyncCommittee `json:"sync_committee"`
	// The trusted slot
	TrustedSlot uint64 `json:"trusted_slot"`
}

// The current sync committee
//
// # The active sync committee
//
// The next sync committee
type ActiveSyncCommittee struct {
	Current *SyncCommittee `json:"Current,omitempty"`
	Next    *SyncCommittee `json:"Next,omitempty"`
}

// The key-value storage proof for a smart contract account
//
// The storage proof used to verify membership
type StorageProof struct {
	// The key of the storage
	Key string `json:"key"`
	// The proof of the storage
	Proof []string `json:"proof"`
	// The value of the storage
	Value string `json:"value"`
}

type TestFixtures struct {
	CommitmentProof CommitmentProof `json:"commitment_proof"`
	InitialState    InitialState    `json:"initial_state"`
	Step            Step            `json:"step"`
	StepsFixture    StepsFixture    `json:"steps_fixture"`
	UpdateClient    UpdateClient    `json:"update_client"`
}

// The proof used to verify membership
type CommitmentProof struct {
	// The client state at the time of the proof
	ClientState ClientState `json:"client_state"`
	// The consensus state at the time of the proof
	ConsensusState ConsensusState `json:"consensus_state"`
	// The IBC path sent to verify membership
	Path string `json:"path"`
	// The slot of the proof (ibc height)
	ProofSlot uint64 `json:"proof_slot"`
	// The storage proof used to verify membership
	StorageProof StorageProof `json:"storage_proof"`
}

// The initial state of the light client in the e2e tests
type InitialState struct {
	// The client state at the initial state
	ClientState ClientState `json:"client_state"`
	// The consensus state at the initial state
	ConsensusState ConsensusState `json:"consensus_state"`
}

// Step is a light client operation such as an initial state, commitment proof, or update
// client
type Step struct {
	// data is the operation data as a JSON object to be deserialized into the appropriate type
	Data interface{} `json:"data"`
	// name is the name of the operation, only used for documentation and easy of reading
	Name string `json:"name"`
}

// A test fixture with an ordered list of light client operations from the e2e test
type StepsFixture struct {
	// steps is a list of light client operations
	Steps []Step `json:"steps"`
}

// Operation to update the light client
type UpdateClient struct {
	// The client state after the update
	ClientState ClientState `json:"client_state"`
	// The consensus state after the update
	ConsensusState ConsensusState `json:"consensus_state"`
	// The headers used to update the light client, in order
	Updates []Header `json:"updates"`
}
