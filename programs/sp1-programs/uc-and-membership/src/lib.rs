//! The crate that contains the types and utilities for `sp1-ics07-tendermint-membership` program.
#![deny(missing_docs, clippy::nursery, clippy::pedantic, warnings)]

use ibc_client_tendermint_types::{ConsensusState, Header};
use ibc_core_commitment_types::merkle::MerkleProof;
use ibc_eureka_solidity_types::msgs::{
    IICS07TendermintMsgs::ClientState, IMembershipMsgs::KVPair,
    IUpdateClientAndMembershipMsgs::UcAndMembershipOutput,
};

/// The main function of the program without the zkVM wrapper.
#[allow(clippy::missing_panics_doc)]
#[must_use]
pub fn update_client_and_membership(
    client_state: ClientState,
    trusted_consensus_state: ConsensusState,
    proposed_header: Header,
    time: u64,
    request_iter: impl Iterator<Item = (KVPair, MerkleProof)>,
) -> UcAndMembershipOutput {
    let app_hash: [u8; 32] = proposed_header
        .signed_header
        .header()
        .app_hash
        .as_bytes()
        .try_into()
        .unwrap();

    let uc_output = sp1_ics07_tendermint_update_client::update_client(
        client_state,
        trusted_consensus_state,
        proposed_header,
        time,
    );

    let mem_output = sp1_ics07_tendermint_membership::membership(app_hash, request_iter);

    UcAndMembershipOutput {
        updateClientOutput: uc_output,
        kvPairs: mem_output.kvPairs,
    }
}
