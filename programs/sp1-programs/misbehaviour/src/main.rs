//! A program that verifies a misbehaviour evidence.

#![deny(missing_docs, clippy::nursery, clippy::pedantic, warnings)]
#![allow(clippy::no_mangle_with_rust_abi)]
// These two lines are necessary for the program to properly compile.
//
// Under the hood, we wrap your main function with some extra code so that it behaves properly
// inside the zkVM.
#![no_main]
sp1_zkvm::entrypoint!(main);

use alloy_sol_types::SolValue;
use ibc_client_tendermint::types::Misbehaviour;
use ibc_eureka_solidity_types::msgs::IICS07TendermintMsgs::{
    ClientState as SolClientState, ConsensusState as SolConsensusState,
};
use ibc_proto::{ibc::lightclients::tendermint::v1::Misbehaviour as RawMisbehaviour, Protobuf};
use sp1_ics07_tendermint_misbehaviour::check_for_misbehaviour;

/// The main function of the program.
///
/// # Panics
/// Panics if the verification fails.
pub fn main() {
    let encoded_1 = sp1_zkvm::io::read_vec();
    let encoded_2 = sp1_zkvm::io::read_vec();
    let encoded_3 = sp1_zkvm::io::read_vec();
    let encoded_4 = sp1_zkvm::io::read_vec();
    let encoded_5 = sp1_zkvm::io::read_vec();

    // input 1: client state
    let client_state = SolClientState::abi_decode(&encoded_1, true).unwrap();
    // input 2: the misbehaviour evidence
    let misbehaviour = <Misbehaviour as Protobuf<RawMisbehaviour>>::decode_vec(&encoded_2).unwrap();
    // input 3: header 1 trusted consensus state
    let trusted_consensus_state_1 = SolConsensusState::abi_decode(&encoded_3, true)
        .unwrap()
        .into();
    // input 4: header 2 trusted consensus state
    let trusted_consensus_state_2 = SolConsensusState::abi_decode(&encoded_4, true)
        .unwrap()
        .into();
    // input 5: time
    let time = u64::from_le_bytes(encoded_5.try_into().unwrap());

    let output = check_for_misbehaviour(
        client_state,
        &misbehaviour,
        trusted_consensus_state_1,
        trusted_consensus_state_2,
        time,
    );

    sp1_zkvm::io::commit_slice(&output.abi_encode());
}
