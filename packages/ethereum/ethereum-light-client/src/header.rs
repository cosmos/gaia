//! This module defines types related to the Header we use for the Ethereum light client

use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use ethereum_types::{
    consensus::{light_client_header::LightClientUpdate, sync_committee::SyncCommittee},
    execution::account_proof::AccountProof,
};

/// The header of a light client update
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct Header {
    /// The trusted sync committee
    pub trusted_sync_committee: TrustedSyncCommittee,
    /// The consensus update
    pub consensus_update: LightClientUpdate,
    /// The account update
    pub account_update: AccountUpdate,
}

/// The account update
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug, Default)]
pub struct AccountUpdate {
    /// The account proof
    pub account_proof: AccountProof,
}

/// The trusted sync committee
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct TrustedSyncCommittee {
    /// The trusted slot
    pub trusted_slot: u64,
    /// The current sync committee
    pub sync_committee: ActiveSyncCommittee,
}

/// The active sync committee
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub enum ActiveSyncCommittee {
    /// The current sync committee
    Current(SyncCommittee),
    /// The next sync committee
    Next(SyncCommittee),
}
