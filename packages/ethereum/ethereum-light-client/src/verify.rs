//! This module contains functions to verify the header of the light client.

use alloy_primitives::B256;

use ethereum_trie_db::trie_db::verify_account_storage_root;
use ethereum_types::consensus::{
    bls::{BlsPublicKey, BlsSignature},
    domain::{compute_domain, DomainType},
    light_client_header::{LightClientHeader, LightClientUpdate},
    merkle::{
        get_subtree_index, EXECUTION_BRANCH_DEPTH, EXECUTION_PAYLOAD_INDEX, FINALITY_BRANCH_DEPTH,
        FINALIZED_ROOT_INDEX, NEXT_SYNC_COMMITTEE_BRANCH_DEPTH, NEXT_SYNC_COMMITTEE_INDEX,
    },
    signing_data::compute_signing_root,
    slot::{compute_epoch_at_slot, compute_slot_at_timestamp, GENESIS_SLOT},
    sync_committee::compute_sync_committee_period_at_slot,
};
use tree_hash::TreeHash;

use crate::{
    client_state::ClientState,
    consensus_state::{ConsensusState, TrustedConsensusState},
    error::EthereumIBCError,
    header::Header,
    trie::validate_merkle_branch,
};

/// The BLS verifier trait.
#[allow(clippy::module_name_repetitions)]
pub trait BlsVerify {
    /// The error type for the BLS verifier.
    type Error: std::fmt::Display;

    /// Verify a BLS signature.
    /// # Errors
    /// Returns an error if the signature cannot be verified.
    fn fast_aggregate_verify(
        &self,
        public_keys: &[BlsPublicKey],
        msg: B256,
        signature: BlsSignature,
    ) -> Result<(), Self::Error>;
}

/// Verifies the header of the light client.
/// # Errors
/// Returns an error if the header cannot be verified.
#[allow(clippy::module_name_repetitions, clippy::needless_pass_by_value)]
pub fn verify_header<V: BlsVerify>(
    consensus_state: &ConsensusState,
    client_state: &ClientState,
    current_timestamp: u64,
    header: &Header,
    bls_verifier: V,
) -> Result<(), EthereumIBCError> {
    let trusted_consensus_state = TrustedConsensusState {
        state: consensus_state.clone(),
        sync_committee: header.trusted_sync_committee.sync_committee.clone(),
    };

    // Ethereum consensus-spec says that we should use the slot at the current timestamp.
    let current_slot = compute_slot_at_timestamp(
        client_state.genesis_time,
        client_state.seconds_per_slot,
        current_timestamp,
    )
    .ok_or(EthereumIBCError::FailedToComputeSlotAtTimestamp {
        timestamp: current_timestamp,
        genesis: client_state.genesis_time,
        seconds_per_slot: client_state.seconds_per_slot,
        genesis_slot: GENESIS_SLOT,
    })?;

    validate_light_client_update::<V>(
        client_state,
        &trusted_consensus_state,
        &header.consensus_update,
        current_slot,
        &bls_verifier,
    )?;

    // check whether at least 2/3 of the sync committee signed
    ensure!(
        header
            .consensus_update
            .sync_aggregate
            .validate_signature_supermajority(),
        EthereumIBCError::NotEnoughSignatures
    );

    let proof_data = header.account_update.account_proof.clone();

    verify_account_storage_root(
        header.consensus_update.attested_header.execution.state_root,
        client_state.ibc_contract_address,
        &proof_data.proof,
        proof_data.storage_root,
    )
    .map_err(|err| EthereumIBCError::VerifyStorageProof(err.to_string()))
}

/// Verifies if the light client `update` is valid.
///
/// * `client_state`: The current client state.
/// * `trusted_consensus_state`: The trusted consensus state (previously verified and stored)
/// * `update`: The update to be verified.
/// * `current_slot`: The slot number computed based on the current timestamp.
/// * `bls_verifier`: BLS verification implementation.
///
/// ## Important Notes
/// * This verification does not assume that the updated header is greater (in terms of slot) than the
///   light client state. When the updated header is in the next signature period, the light client uses
///   the next sync committee to verify the signature, then it saves the next sync committee as the current
///   sync committee. However, it's not mandatory for light clients to expect the next sync committee to be given
///   during these updates. So if it's not given, the light client still can validate updates until the next signature
///   period arrives. In a situation like this, the update can be any header within the same signature period. And
///   this function only allows a non-existent next sync committee to be set in that case. It doesn't allow a sync committee
///   to be changed or removed.
///
/// [See in consensus-spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/altair/light-client/sync-protocol.md#validate_light_client_update)
/// # Errors
/// Returns an error if the update cannot be verified.
/// # Panics
/// If the minimum sync committee participants is not a valid usize.
#[allow(clippy::too_many_lines, clippy::needless_pass_by_value)]
pub fn validate_light_client_update<V: BlsVerify>(
    client_state: &ClientState,
    trusted_consensus_state: &TrustedConsensusState,
    update: &LightClientUpdate,
    current_slot: u64,
    bls_verifier: &V,
) -> Result<(), EthereumIBCError> {
    // Verify sync committee has sufficient participants
    ensure!(
        update
            .sync_aggregate
            .has_sufficient_participants(client_state.min_sync_committee_participants),
        EthereumIBCError::InsufficientSyncCommitteeParticipants(
            update.sync_aggregate.num_sync_committe_participants(),
        )
    );

    is_valid_light_client_header(client_state, &update.attested_header)?;

    // Verify update does not skip a sync committee period
    let update_attested_slot = update.attested_header.beacon.slot;
    let update_finalized_slot = update.finalized_header.beacon.slot;

    ensure!(
        update_finalized_slot != GENESIS_SLOT,
        EthereumIBCError::FinalizedSlotIsGenesis
    );

    ensure!(
        current_slot >= update.signature_slot,
        EthereumIBCError::UpdateMoreRecentThanCurrentSlot {
            current_slot,
            update_signature_slot: update.signature_slot
        }
    );

    ensure!(
        update.signature_slot > update_attested_slot
            && update_attested_slot >= update_finalized_slot,
        EthereumIBCError::InvalidSlots {
            update_signature_slot: update.signature_slot,
            update_attested_slot,
            update_finalized_slot,
        }
    );

    // Let's say N is the signature period of the header we store, we can only do updates with
    // the following settings:
    // 1. stored_period = N, signature_period = N:
    //     - the light client must have the `current_sync_committee` and use it to verify the new header.
    // 2. stored_period = N, signature_period = N + 1:
    //     - the light client must have the `next_sync_committee` and use it to verify the new header.
    let stored_period = compute_sync_committee_period_at_slot(
        client_state.slots_per_epoch,
        client_state.epochs_per_sync_committee_period,
        trusted_consensus_state.finalized_slot(),
    );
    let signature_period = compute_sync_committee_period_at_slot(
        client_state.slots_per_epoch,
        client_state.epochs_per_sync_committee_period,
        update.signature_slot,
    );

    if trusted_consensus_state.next_sync_committee().is_some() {
        ensure!(
            signature_period == stored_period || signature_period == stored_period + 1,
            EthereumIBCError::InvalidSignaturePeriodWhenNextSyncCommitteeExists {
                signature_period,
                stored_period,
            }
        );
    } else {
        ensure!(
            signature_period == stored_period,
            EthereumIBCError::InvalidSignaturePeriodWhenNextSyncCommitteeDoesNotExist {
                signature_period,
                stored_period,
            }
        );
    }

    // Verify update is relevant
    let update_attested_period = compute_sync_committee_period_at_slot(
        client_state.slots_per_epoch,
        client_state.epochs_per_sync_committee_period,
        update_attested_slot,
    );

    // There are two options to do a light client update:
    // 1. We are updating the header with a newer one.
    // 2. We haven't set the next sync committee yet and we can use any attested header within the same
    // signature period to set the next sync committee. This means that the stored header could be larger.
    // The light client implementation needs to take care of it.
    ensure!(
        update_attested_slot > trusted_consensus_state.finalized_slot()
            || (update_attested_period == stored_period
                && update.next_sync_committee.is_some()
                && trusted_consensus_state.next_sync_committee().is_none()),
        EthereumIBCError::IrrelevantUpdate {
            update_attested_slot,
            trusted_finalized_slot: trusted_consensus_state.finalized_slot(),
            update_attested_period,
            stored_period,
            update_sync_committee_is_set: update.next_sync_committee.is_some(),
            trusted_next_sync_committee_is_set: trusted_consensus_state
                .next_sync_committee()
                .is_some(),
        }
    );

    // Verify that the `finality_branch`, if present, confirms `finalized_header`
    // to match the finalized checkpoint root saved in the state of `attested_header`.
    is_valid_light_client_header(client_state, &update.finalized_header)?;
    let finalized_root = update.finalized_header.beacon.tree_hash_root();

    // This confirms that the `finalized_header` is really finalized.
    validate_merkle_branch(
        finalized_root,
        update.finality_branch.into(),
        FINALITY_BRANCH_DEPTH,
        get_subtree_index(FINALIZED_ROOT_INDEX),
        update.attested_header.beacon.state_root,
    )
    .map_err(|e| EthereumIBCError::ValidateFinalizedHeaderFailed(Box::new(e)))?;

    // Verify that if the update contains the next sync committee, and the signature periods do match,
    // next sync committees match too.
    if let (Some(next_sync_committee), Some(stored_next_sync_committee)) = (
        &update.next_sync_committee,
        trusted_consensus_state.next_sync_committee(),
    ) {
        if update_attested_period == stored_period {
            ensure!(
                next_sync_committee == stored_next_sync_committee,
                EthereumIBCError::NextSyncCommitteeMismatch {
                    expected: stored_next_sync_committee.aggregate_pubkey,
                    found: next_sync_committee.aggregate_pubkey,
                }
            );
        }
        // This validates the given next sync committee against the attested header's state root.
        validate_merkle_branch(
            next_sync_committee.tree_hash_root(),
            update.next_sync_committee_branch.unwrap_or_default().into(),
            NEXT_SYNC_COMMITTEE_BRANCH_DEPTH,
            get_subtree_index(NEXT_SYNC_COMMITTEE_INDEX),
            update.attested_header.beacon.state_root,
        )
        .map_err(|e| EthereumIBCError::ValidateNextSyncCommitteeFailed(Box::new(e)))?;
    }

    // Verify sync committee aggregate signature
    let sync_committee = if signature_period == stored_period {
        trusted_consensus_state
            .current_sync_committee()
            .ok_or(EthereumIBCError::ExpectedCurrentSyncCommittee)?
    } else {
        trusted_consensus_state
            .next_sync_committee()
            .ok_or(EthereumIBCError::ExpectedNextSyncCommittee)?
    };

    // It's not mandatory for all of the members of the sync committee to participate. So we are extracting the
    // public keys of the ones who participated.
    let participant_pubkeys = update
        .sync_aggregate
        .sync_committee_bits
        .iter()
        .flat_map(|byte| (0..8).rev().map(move |i| (byte & (1 << i)) != 0))
        .zip(sync_committee.pubkeys.iter())
        .filter_map(|(included, pubkey)| included.then_some(*pubkey))
        .collect::<Vec<_>>();

    let fork_version_slot = std::cmp::max(update.signature_slot, 1) - 1;
    let fork_version = client_state
        .fork_parameters
        .compute_fork_version(compute_epoch_at_slot(
            client_state.slots_per_epoch,
            fork_version_slot,
        ));

    let domain = compute_domain(
        DomainType::SYNC_COMMITTEE,
        Some(fork_version),
        Some(client_state.genesis_validators_root),
        client_state.fork_parameters.genesis_fork_version,
    );
    let signing_root = compute_signing_root(&update.attested_header.beacon, domain);

    bls_verifier
        .fast_aggregate_verify(
            &participant_pubkeys,
            signing_root,
            update.sync_aggregate.sync_committee_signature,
        )
        .map_err(|err| EthereumIBCError::FastAggregateVerify(err.to_string()))?;

    Ok(())
}

/// Validates a light client header.
///
/// [See in consensus-spec](https://github.com/ethereum/consensus-specs/blob/dev/specs/deneb/light-client/sync-protocol.md#modified-is_valid_light_client_header)
/// # Errors
/// Returns an error if the header cannot be validated.
pub fn is_valid_light_client_header(
    client_state: &ClientState,
    header: &LightClientHeader,
) -> Result<(), EthereumIBCError> {
    let epoch = compute_epoch_at_slot(client_state.slots_per_epoch, header.beacon.slot);

    if epoch < client_state.fork_parameters.deneb.epoch {
        ensure!(
            header.execution.blob_gas_used == 0 && header.execution.excess_blob_gas == 0,
            EthereumIBCError::MustBeDeneb
        );
    }

    ensure!(
        epoch >= client_state.fork_parameters.capella.epoch,
        EthereumIBCError::InvalidChainVersion
    );

    validate_merkle_branch(
        get_lc_execution_root(client_state, header)?,
        header.execution_branch.into(),
        EXECUTION_BRANCH_DEPTH,
        get_subtree_index(EXECUTION_PAYLOAD_INDEX),
        header.beacon.body_root,
    )
}

/// Returns the execution root of the light client header.
/// # Errors
/// Returns an error if the epoch is less than the Deneb epoch.
pub fn get_lc_execution_root(
    client_state: &ClientState,
    header: &LightClientHeader,
) -> Result<B256, EthereumIBCError> {
    let epoch = compute_epoch_at_slot(client_state.slots_per_epoch, header.beacon.slot);

    ensure!(
        epoch >= client_state.fork_parameters.deneb.epoch,
        EthereumIBCError::MustBeDeneb
    );

    Ok(header.execution.tree_hash_root())
}

#[cfg(test)]
mod test {
    use crate::test_utils::{
        bls_verifier::{fast_aggregate_verify, BlsError},
        fixtures::{self, InitialState, UpdateClient},
    };

    use super::*;

    struct TestBlsVerifier;

    impl BlsVerify for TestBlsVerifier {
        type Error = BlsError;

        fn fast_aggregate_verify(
            &self,
            public_keys: &[BlsPublicKey],
            msg: B256,
            signature: BlsSignature,
        ) -> Result<(), BlsError> {
            fast_aggregate_verify(public_keys, msg, signature)
        }
    }

    #[test]
    fn test_verify_header() {
        let bls_verifier = TestBlsVerifier;

        let fixture: fixtures::StepsFixture =
            fixtures::load("TestICS20TransferNativeCosmosCoinsToEthereumAndBack_Groth16");

        let initial_state: InitialState = fixture.get_data_at_step(0);

        let client_state = initial_state.client_state;
        let consensus_state = initial_state.consensus_state;

        let update_state: UpdateClient = fixture.get_data_at_step(1);

        let header = update_state.updates[0].clone();
        verify_header(
            &consensus_state,
            &client_state,
            header.consensus_update.attested_header.execution.timestamp + 1000,
            &header,
            bls_verifier,
        )
        .unwrap();
    }
}
