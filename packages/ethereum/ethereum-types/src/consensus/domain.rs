//! This module defines types and functions related to the signature domain.

use alloy_primitives::{hex, FixedBytes, B256};
use serde::{Deserialize, Serialize};

use super::fork::{compute_fork_data_root, Version};

/// The signature domain type.
/// Defined in
/// <https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#domain-types>
#[derive(Serialize, Deserialize, PartialEq, Eq, Clone, Debug, Default)]
#[allow(clippy::module_name_repetitions)]
pub struct DomainType(pub [u8; 4]);

#[allow(missing_docs)]
impl DomainType {
    pub const BEACON_PROPOSER: Self = Self(hex!("00000000"));
    pub const BEACON_ATTESTER: Self = Self(hex!("01000000"));
    pub const RANDAO: Self = Self(hex!("02000000"));
    pub const DEPOSIT: Self = Self(hex!("03000000"));
    pub const VOLUNTARY_EXIT: Self = Self(hex!("04000000"));
    pub const SELECTION_PROOF: Self = Self(hex!("05000000"));
    pub const AGGREGATE_AND_PROOF: Self = Self(hex!("06000000"));
    pub const SYNC_COMMITTEE: Self = Self(hex!("07000000"));
    pub const SYNC_COMMITTEE_SELECTION_PROOF: Self = Self(hex!("08000000"));
    pub const CONTRIBUTION_AND_PROOF: Self = Self(hex!("09000000"));
    pub const BLS_TO_EXECUTION_CHANGE: Self = Self(hex!("0A000000"));
    pub const APPLICATION_MASK: Self = Self(hex!("00000001"));
}

/// Return the domain for the `domain_type` and `fork_version`.
///
/// [See in consensus-spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#compute_domain)
#[allow(clippy::module_name_repetitions, clippy::needless_pass_by_value)]
#[must_use]
pub fn compute_domain(
    domain_type: DomainType,
    fork_version: Option<Version>,
    genesis_validators_root: Option<B256>,
    genesis_fork_version: Version,
) -> B256 {
    let fork_version = fork_version.unwrap_or(genesis_fork_version);
    let genesis_validators_root = genesis_validators_root.unwrap_or_default();
    let fork_data_root = compute_fork_data_root(fork_version, genesis_validators_root);

    let mut domain = [0; 32];
    domain[..4].copy_from_slice(&domain_type.0);
    domain[4..].copy_from_slice(&fork_data_root[..28]);

    FixedBytes(domain)
}

#[cfg(test)]
mod test {
    use hex::FromHex;

    use super::*;

    const GENESIS_FORK_VERSION: Version = FixedBytes([0, 0, 0, 1]);
    const DENEB_FORK_VERSION: Version = FixedBytes([4, 0, 0, 1]);

    #[test]
    fn test_compute_domain() {
        let domain_type = DomainType::SYNC_COMMITTEE;
        let fork_version = DENEB_FORK_VERSION;
        let genesis_validators_root =
            B256::from_hex("d61ea484febacfae5298d52a2b581f3e305a51f3112a9241b968dccf019f7b11")
                .unwrap();
        let genesis_fork_version = GENESIS_FORK_VERSION;

        let domain = compute_domain(
            domain_type,
            Some(fork_version),
            Some(genesis_validators_root),
            genesis_fork_version,
        );

        // expected domain taken from running the same code in the union repo
        let expected =
            B256::from_hex("07000000eaa5664b85c5e9dc16d64ac6ee15cc92ec477990061b30024696db67")
                .unwrap();

        assert_eq!(domain, expected);
    }

    #[test]
    fn test_compute_domain_with_union_data() {
        // this test is essentially a copy of the union unit test for compute_domain
        let domain_type = DomainType([1, 2, 3, 4]);
        let current_version = Version::from([5, 6, 7, 8]);
        let genesis_validators_root = B256::new([1; 32]);
        let fork_data_root = compute_fork_data_root(current_version, genesis_validators_root);
        let genesis_version = Version::from([0, 0, 0, 0]);

        let mut domain = B256::default();
        domain.0[..4].copy_from_slice(&domain_type.0);
        domain.0[4..].copy_from_slice(&fork_data_root[..28]);

        // Uses the values instead of the default ones when `current_version` and
        // `genesis_validators_root` is provided.
        assert_eq!(
            domain,
            compute_domain(
                domain_type.clone(),
                Some(current_version),
                Some(genesis_validators_root),
                genesis_version,
            )
        );

        let fork_data_root = compute_fork_data_root(genesis_version, FixedBytes::default());
        let mut domain = B256::default();
        domain.0[..4].copy_from_slice(&domain_type.0);
        domain.0[4..].copy_from_slice(&fork_data_root[..28]);

        // Uses default values when version and validators root is None
        assert_eq!(
            domain,
            compute_domain(domain_type, None, None, genesis_version)
        );
    }
}
