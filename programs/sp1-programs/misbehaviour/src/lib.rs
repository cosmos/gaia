//! The crate that contains the types and utilities for `sp1-ics07-tendermint-update-client`
//! program.
#![deny(missing_docs, clippy::nursery, clippy::pedantic, warnings)]

pub mod types;

use ibc_client_tendermint::client_state::{
    check_for_misbehaviour_on_misbehavior, verify_misbehaviour,
};
use ibc_client_tendermint::types::{ConsensusState, Misbehaviour, TENDERMINT_CLIENT_TYPE};
use ibc_core_host_types::identifiers::{ChainId, ClientId};
use ibc_eureka_solidity_types::msgs::{
    IICS07TendermintMsgs::ClientState, IMisbehaviourMsgs::MisbehaviourOutput,
};
use std::collections::HashMap;
use std::time::Duration;
use tendermint_light_client_verifier::options::Options;
use tendermint_light_client_verifier::ProdVerifier;

/// The main function of the program without the zkVM wrapper.
#[allow(clippy::missing_panics_doc)]
#[must_use]
pub fn check_for_misbehaviour(
    client_state: ClientState,
    misbehaviour: &Misbehaviour,
    trusted_consensus_state_1: ConsensusState,
    trusted_consensus_state_2: ConsensusState,
    time: u64,
) -> MisbehaviourOutput {
    let client_id = ClientId::new(TENDERMINT_CLIENT_TYPE, 0).unwrap();
    assert_eq!(
        client_state.chainId,
        misbehaviour
            .header1()
            .signed_header
            .header
            .chain_id
            .to_string()
    );

    // Insert the two trusted consensus states into the trusted consensus state map that exists in the ClientValidationContext that is expected by verifyMisbehaviour
    // Since we are mocking the existence of prior trusted consensus states, we are only filling in the two consensus states that are passed in into the map
    let trusted_consensus_state_map = HashMap::from([
        (
            misbehaviour.header1().trusted_height.revision_height(),
            &trusted_consensus_state_1,
        ),
        (
            misbehaviour.header2().trusted_height.revision_height(),
            &trusted_consensus_state_2,
        ),
    ]);
    let ctx =
        types::validation::MisbehaviourValidationContext::new(time, trusted_consensus_state_map);

    let options = Options {
        trust_threshold: client_state.trustLevel.clone().into(),
        trusting_period: Duration::from_secs(client_state.trustingPeriod.into()),
        clock_drift: Duration::default(),
    };

    // Call into ibc-rs verify_misbehaviour function to verify that both headers are valid given their respective trusted consensus states
    verify_misbehaviour::<_, sha2::Sha256>(
        &ctx,
        misbehaviour,
        &client_id,
        &ChainId::new(&client_state.chainId).unwrap(),
        &options,
        &ProdVerifier::default(),
    )
    .unwrap();

    // Call into ibc-rs check_for_misbehaviour_on_misbehaviour method to ensure that the misbehaviour is valid
    // i.e. the headers are same height but different commits, or headers are not monotonically increasing in time
    let is_misbehaviour =
        check_for_misbehaviour_on_misbehavior(misbehaviour.header1(), misbehaviour.header2())
            .unwrap();
    assert!(is_misbehaviour, "Misbehaviour is not detected");

    // The prover takes in the trusted headers as an input but does not maintain its own internal state
    // Thus, the verifier must ensure that the trusted headers that were used in the proof are trusted consensus
    // states stored in its own internal state before it can accept the misbehaviour proof as valid.
    MisbehaviourOutput {
        clientState: client_state,
        trustedHeight1: misbehaviour.header1().trusted_height.try_into().unwrap(),
        trustedHeight2: misbehaviour.header2().trusted_height.try_into().unwrap(),
        trustedConsensusState1: trusted_consensus_state_1.into(),
        trustedConsensusState2: trusted_consensus_state_2.into(),
        time,
    }
}
