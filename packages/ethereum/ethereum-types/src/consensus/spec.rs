//! This module defines types related to Spec.

use serde::{Deserialize, Serialize};

/// The spec type, returned from the beacon api.
#[derive(Serialize, Deserialize, PartialEq, Eq, Clone, Debug, Default)]
#[serde(rename_all = "SCREAMING_SNAKE_CASE")]
pub struct Spec {
    /// The number of seconds per slot.
    pub seconds_per_slot: u64,
    /// The number of slots per epoch.
    pub slots_per_epoch: u64,
    /// The number of epochs per sync committee period.
    pub epochs_per_sync_committee_period: u64,

    // Fork Parameters
    /// The genesis fork version.
    pub genesis_fork_version: String,
    /// The genesis slot.
    pub genesis_slot: u64,
    /// The altair fork version.
    pub altair_fork_version: String,
    /// The altair fork epoch.
    pub altair_fork_epoch: u64,
    /// The bellatrix fork version.
    pub bellatrix_fork_version: String,
    /// The bellatrix fork epoch.
    pub bellatrix_fork_epoch: u64,
    /// The capella fork version.
    pub capella_fork_version: String,
    /// The capella fork epoch.
    pub capella_fork_epoch: u64,
    /// The deneb fork version.
    pub deneb_fork_version: String,
    /// The deneb fork epoch.
    pub deneb_fork_epoch: u64,
}

impl Spec {
    /// Returns the number of slots in a sync committee period.
    #[must_use]
    pub const fn period(&self) -> u64 {
        self.epochs_per_sync_committee_period * self.slots_per_epoch
    }
}
