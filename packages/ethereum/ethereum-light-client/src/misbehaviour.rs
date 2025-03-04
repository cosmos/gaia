//! This module provides [`verify_misbehavior`] function to check for misbehaviour

use ethereum_types::consensus::{
    light_client_header::LightClientUpdate,
    slot::{compute_slot_at_timestamp, GENESIS_SLOT},
};

use crate::{
    client_state::ClientState,
    consensus_state::TrustedConsensusState,
    error::EthereumIBCError,
    verify::{validate_light_client_update, BlsVerify},
};

/// Verifies if a consensus misbehaviour is valid by checking if the two conflicting light client updates are valid.
///
/// * `client_state`: The current client state.
/// * `trusted_consensus_state`: The trusted consensus state (previously verified and stored)
/// * `update_1`: The first light client update.
/// * `update_2`: The second light client update.
/// * `current_slot`: The slot number computed based on the current timestamp.
/// * `bls_verifier`: BLS verification implementation.
///
/// # Errors
/// Returns an error if the misbehaviour cannot be verified.
#[allow(clippy::module_name_repetitions, clippy::needless_pass_by_value)]
pub fn verify_misbehaviour<V: BlsVerify>(
    client_state: &ClientState,
    trusted_consensus_state: &TrustedConsensusState,
    update_1: &LightClientUpdate,
    update_2: &LightClientUpdate,
    current_timestamp: u64,
    bls_verifier: V,
) -> Result<(), EthereumIBCError> {
    // There is no point to check for misbehaviour when the headers are not for the same height
    let (slot_1, slot_2) = (
        update_1.finalized_header.beacon.slot,
        update_2.finalized_header.beacon.slot,
    );
    ensure!(
        slot_1 == slot_2,
        EthereumIBCError::MisbehaviourSlotMismatch(slot_1, slot_2)
    );

    let (state_root_1, state_root_2) = (
        update_1.attested_header.execution.state_root,
        update_2.attested_header.execution.state_root,
    );
    ensure!(
        state_root_1 != state_root_2,
        EthereumIBCError::MisbehaviourStorageRootsMatch(state_root_1)
    );

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
        trusted_consensus_state,
        update_1,
        current_slot,
        &bls_verifier,
    )?;

    validate_light_client_update::<V>(
        client_state,
        trusted_consensus_state,
        update_2,
        current_slot,
        &bls_verifier,
    )?;

    Ok(())
}
