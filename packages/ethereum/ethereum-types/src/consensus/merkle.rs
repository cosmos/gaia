//! This module defines constants related to merkle trees in the Ethereum consensus.

// https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/light-client/sync-protocol.md#constants
// REVIEW: Is it possible to implement get_generalized_index in const rust?

// https://github.com/ethereum/consensus-specs/blob/dev/ssz/merkle-proofs.md
/// `get_generalized_index(BeaconState, "finalized_checkpoint", "root")`
pub const FINALIZED_ROOT_INDEX: u64 = 105;
/// `get_generalized_index(BeaconState, "current_sync_committee")`
pub const CURRENT_SYNC_COMMITTEE_INDEX: u64 = 54;
/// `get_generalized_index(BeaconState, "next_sync_committee")`
pub const NEXT_SYNC_COMMITTEE_INDEX: u64 = 55;
/// `get_generalized_index(BeaconBlockBody, "execution_payload")`
pub const EXECUTION_PAYLOAD_INDEX: u64 = 25;

// Branch depths for different merkle trees related to ethereum consensus

/// The depth of the merkle tree for execution payloads.
pub const EXECUTION_BRANCH_DEPTH: usize = floorlog2(EXECUTION_PAYLOAD_INDEX);
/// The depth of the merkle tree for the next sync committee.
pub const NEXT_SYNC_COMMITTEE_BRANCH_DEPTH: usize = floorlog2(NEXT_SYNC_COMMITTEE_INDEX);
/// The depth of the merkle tree for the finalized root.
pub const FINALITY_BRANCH_DEPTH: usize = floorlog2(FINALIZED_ROOT_INDEX);

/// Values that are constant across all configurations.
/// <https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/light-client/sync-protocol.md#get_subtree_index>
#[must_use]
pub const fn get_subtree_index(idx: u64) -> u64 {
    idx % 2_u64.pow(idx.ilog2())
}

/// Convenience function safely to call [`u64::ilog2`] and convert the result into a usize.
#[cfg(any(target_pointer_width = "32", target_pointer_width = "64"))]
#[must_use]
const fn floorlog2(n: u64) -> usize {
    // conversion is safe since usize is either 32 or 64 bits as per cfg above
    n.ilog2() as usize
}
