//! This module defines [`compute_signing_root`].

use alloy_primitives::B256;
use serde::{Deserialize, Serialize};
use tree_hash::TreeHash;
use tree_hash_derive::TreeHash;

#[derive(Serialize, Deserialize, PartialEq, Clone, Debug, Default, TreeHash)]
struct SigningData {
    pub object_root: B256,
    pub domain: B256,
}

/// Return the signing root for the corresponding signing data
///
/// [See in consensus-spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#compute_signing_root)
pub fn compute_signing_root<T: TreeHash>(ssz_object: &T, domain: B256) -> B256 {
    SigningData {
        object_root: ssz_object.tree_hash_root(),
        domain,
    }
    .tree_hash_root()
}
