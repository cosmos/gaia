//! The crate that contains the types and utilities for `sp1-ics07-tendermint-update-client`
//! program.
#![deny(missing_docs, clippy::nursery, clippy::pedantic, warnings)]

pub mod types;

use std::{str::FromStr, time::Duration};

use ibc_client_tendermint::{
    client_state::verify_header,
    types::{ConsensusState, Header, TENDERMINT_CLIENT_TYPE},
};
use ibc_core_host_types::identifiers::{ChainId, ClientId};
use ibc_eureka_solidity_types::msgs::{
    IICS07TendermintMsgs::ClientState, IUpdateClientMsgs::UpdateClientOutput,
};

use tendermint_light_client_verifier::{options::Options, ProdVerifier};

/// The main function of the program without the zkVM wrapper.
#[allow(clippy::missing_panics_doc)]
#[must_use]
pub fn update_client(
    client_state: ClientState,
    trusted_consensus_state: ConsensusState,
    proposed_header: Header,
    time: u64,
) -> UpdateClientOutput {
    let client_id = ClientId::new(TENDERMINT_CLIENT_TYPE, 0).unwrap();
    let chain_id = ChainId::from_str(&client_state.chainId).unwrap();
    let options = Options {
        trust_threshold: client_state.trustLevel.clone().into(),
        trusting_period: Duration::from_secs(client_state.trustingPeriod.into()),
        clock_drift: Duration::default(),
    };

    let ctx = types::validation::ClientValidationCtx::new(time, &trusted_consensus_state);

    verify_header::<_, sha2::Sha256>(
        &ctx,
        &proposed_header,
        &client_id,
        &chain_id,
        &options,
        &ProdVerifier::default(),
    )
    .unwrap();

    let trusted_height = proposed_header.trusted_height.try_into().unwrap();
    let new_height = proposed_header.height().try_into().unwrap();
    let new_consensus_state = ConsensusState::from(proposed_header);

    UpdateClientOutput {
        clientState: client_state,
        trustedConsensusState: trusted_consensus_state.into(),
        newConsensusState: new_consensus_state.into(),
        time,
        trustedHeight: trusted_height,
        newHeight: new_height,
    }
}
