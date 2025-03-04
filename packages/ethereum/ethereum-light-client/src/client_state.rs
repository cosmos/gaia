//! This module defines [`ClientState`].

use alloy_primitives::{Address, B256, U256};
use ethereum_types::consensus::fork::ForkParameters;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// The ethereum client state
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug, Default)]
pub struct ClientState {
    /// The chain ID
    pub chain_id: u64,
    /// The genesis validators root
    #[schemars(with = "String")]
    pub genesis_validators_root: B256,
    /// The minimum number of participants in the sync committee
    pub min_sync_committee_participants: u64,
    /// The time of genesis (unix timestamp)
    pub genesis_time: u64,
    /// The fork parameters
    pub fork_parameters: ForkParameters,
    /// The slot duration in seconds
    pub seconds_per_slot: u64,
    /// The number of slots per epoch
    pub slots_per_epoch: u64,
    /// The number of epochs per sync committee period
    pub epochs_per_sync_committee_period: u64,
    /// The latest slot of this client
    pub latest_slot: u64,
    /// Whether the client is frozen
    pub is_frozen: bool,
    /// The address of the IBC contract being tracked on Ethereum
    #[schemars(with = "String")]
    pub ibc_contract_address: Address,
    /// The storage slot of the IBC commitment in the Ethereum contract
    #[schemars(with = "String")]
    pub ibc_commitment_slot: U256,
}
