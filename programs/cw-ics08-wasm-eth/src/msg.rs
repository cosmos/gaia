//! The messages that are passed between the contract and the ibc-go module
#![allow(clippy::module_name_repetitions)]

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;
use ethereum_light_client::header::ActiveSyncCommittee;
use ethereum_types::consensus::light_client_header::LightClientUpdate;

/// The message to instantiate the contract
#[cw_serde]
pub struct InstantiateMsg {
    /// The client state
    pub client_state: Binary,
    /// The consensus state
    pub consensus_state: Binary,
    /// The checksum of this wasm code
    pub checksum: Binary,
}

/// The unused message to execute the contract
#[cw_serde]
pub enum ExecuteMsg {}

/// The sudo messages called by `ibc-go`
#[cw_serde]
pub enum SudoMsg {
    /// The message to verify membership
    VerifyMembership(VerifyMembershipMsg),
    /// The message to verify non-membership
    VerifyNonMembership(VerifyNonMembershipMsg),
    /// The message to update the client state
    UpdateState(UpdateStateMsg),
    /// The message to update the client state on misbehaviour
    UpdateStateOnMisbehaviour(UpdateStateOnMisbehaviourMsg),
    /// The message to verify the upgrade and update the state
    VerifyUpgradeAndUpdateState(VerifyUpgradeAndUpdateStateMsg),
    /// The message to migrate the client store
    MigrateClientStore(MigrateClientStoreMsg),
}

/// Verify membership message
#[cw_serde]
pub struct VerifyMembershipMsg {
    /// The proof height
    pub height: Height,
    /// The delay time period (unused)
    pub delay_time_period: u64,
    /// The delay block period (unused)
    pub delay_block_period: u64,
    /// The proof bytes
    pub proof: Binary,
    /// The path to the value
    pub merkle_path: MerklePath,
    /// The value to verify
    pub value: Binary,
}

/// Verify non-membership message
#[cw_serde]
pub struct VerifyNonMembershipMsg {
    /// The proof height
    pub height: Height,
    /// The delay time period (unused)
    pub delay_time_period: u64,
    /// The delay block period (unused)
    pub delay_block_period: u64,
    /// The proof bytes
    pub proof: Binary,
    /// The path to the empty value
    pub merkle_path: MerklePath,
}

/// Update state message
#[cw_serde]
pub struct UpdateStateMsg {
    /// The client message
    pub client_message: Binary,
}

/// Update state on misbehaviour message
#[cw_serde]
pub struct UpdateStateOnMisbehaviourMsg {
    /// The client message
    pub client_message: Binary,
}

/// Verify upgrade and update state message
#[cw_serde]
pub struct VerifyUpgradeAndUpdateStateMsg {
    /// The upgraded client state
    pub upgrade_client_state: Binary,
    /// The upgraded consensus state
    pub upgrade_consensus_state: Binary,
    /// The proof of the upgraded client state
    pub proof_upgrade_client: Binary,
    /// The proof of the upgraded consensus state
    pub proof_upgrade_consensus_state: Binary,
}

/// Migrate client store message
#[cw_serde]
pub struct MigrateClientStoreMsg {}

/// The misbehaviour message for the ethereum light client
#[cw_serde]
pub struct EthereumMisbehaviourMsg {
    /// The slot of the trusted consensus state
    pub trusted_slot: u64,
    /// The trusted sync committee, active or next
    pub sync_committee: ActiveSyncCommittee,
    /// The first light client update
    pub update_1: LightClientUpdate,
    /// The second conflicting light client update
    pub update_2: LightClientUpdate,
}

/// The query messages called by `ibc-go`
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// The message to verify the client message
    #[returns[()]]
    VerifyClientMessage(VerifyClientMessageMsg),

    /// The message to check for misbehaviour
    #[returns[CheckForMisbehaviourResult]]
    CheckForMisbehaviour(CheckForMisbehaviourMsg),

    /// The message to get the timestamp at height
    #[returns[TimestampAtHeightResult]]
    TimestampAtHeight(TimestampAtHeightMsg),

    /// The message to get the status
    #[returns[StatusResult]]
    Status(StatusMsg),
}

/// The message to verify the client message
#[cw_serde]
pub struct VerifyClientMessageMsg {
    /// The client message to verify
    pub client_message: Binary,
}

/// The message to check for misbehaviour
#[cw_serde]
pub struct CheckForMisbehaviourMsg {
    /// The client message to check
    pub client_message: Binary,
}

/// The message to get the timestamp at height
#[cw_serde]
pub struct TimestampAtHeightMsg {
    /// The height to get the timestamp at
    pub height: Height,
}

/// The status query message
#[cw_serde]
pub struct StatusMsg {}

/// Height of the ethereum chain
#[cw_serde]
pub struct Height {
    /// The revision that the client is currently on
    /// Always zero in the ethereum light client
    #[serde(default)]
    pub revision_number: u64,
    /// The execution height of ethereum chain
    #[serde(default)]
    pub revision_height: u64,
}

/// The result of updating the client state
#[cw_serde]
pub struct UpdateStateResult {
    /// The updated client state heights
    pub heights: Vec<Height>,
}

/// The merkle path
#[cw_serde]
pub struct MerklePath {
    /// The key path
    pub key_path: Vec<Binary>,
}

/// The response to the status query
#[cw_serde]
pub struct StatusResult {
    /// The status of the client
    pub status: String,
}

/// The response to the check for misbehaviour query
#[cw_serde]
pub struct CheckForMisbehaviourResult {
    /// Whether the client has found misbehaviour
    pub found_misbehaviour: bool,
}

/// The response to the timestamp at height query
#[cw_serde]
pub struct TimestampAtHeightResult {
    /// The timestamp at the height (in nanoseconds)
    pub timestamp: u64,
}
