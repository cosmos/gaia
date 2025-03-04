//! This module defines types related to forks in Ethereum.

use alloy_primitives::{aliases::B32, B256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use tree_hash::TreeHash;
use tree_hash_derive::TreeHash;

/// Type alias Etheruem Version which is a fixed 4 byte array
pub type Version = B32;

/// The fork data
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug, Default)]
pub struct Fork {
    /// The version of the fork
    #[schemars(with = "String")]
    pub version: Version,
    /// The epoch at which this fork is activated
    pub epoch: u64,
}

/// The fork data
#[derive(Serialize, Deserialize, PartialEq, Clone, Debug, Default, TreeHash)]
struct ForkData {
    /// The current version
    pub current_version: Version,
    /// The genesis validators root
    pub genesis_validators_root: B256,
}

/// The fork parameters
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug, Default)]
#[allow(clippy::module_name_repetitions)]
pub struct ForkParameters {
    /// The genesis fork version
    #[schemars(with = "String")]
    pub genesis_fork_version: Version,
    /// The genesis slot
    pub genesis_slot: u64,
    /// The altair fork
    pub altair: Fork,
    /// The bellatrix fork
    pub bellatrix: Fork,
    /// The capella fork
    pub capella: Fork,
    /// The deneb fork
    pub deneb: Fork,
}

impl ForkParameters {
    /// Returns the fork version based on the `epoch`.
    /// NOTE: This implementation is based on capella.
    ///
    /// [See in consensus-spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/capella/fork.md#modified-compute_fork_version)
    #[must_use]
    pub const fn compute_fork_version(&self, epoch: u64) -> Version {
        match epoch {
            _ if epoch >= self.deneb.epoch => self.deneb.version,
            _ if epoch >= self.capella.epoch => self.capella.version,
            _ if epoch >= self.bellatrix.epoch => self.bellatrix.version,
            _ if epoch >= self.altair.epoch => self.altair.version,
            _ => self.genesis_fork_version,
        }
    }
}

/// Return the 32-byte fork data root for the `current_version` and `genesis_validators_root`.
/// This is used primarily in signature domains to avoid collisions across forks/chains.
///
/// [See in consensus-spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#compute_fork_data_root)
#[must_use]
pub fn compute_fork_data_root(current_version: Version, genesis_validators_root: B256) -> B256 {
    let fork_data = ForkData {
        current_version,
        genesis_validators_root,
    };

    fork_data.tree_hash_root()
}
