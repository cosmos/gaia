//! Contains types and traits for `verify_misbehaviour` validation within the program.

use ibc_client_tendermint::{
    client_state::ClientState as ClientStateWrapper,
    consensus_state::ConsensusState as ConsensusStateWrapper, types::ConsensusState,
};
use ibc_core_client::context::{ClientValidationContext, ExtClientValidationContext};
use ibc_core_host_types::error::HostError;
use ibc_primitives::Timestamp;
use std::collections::HashMap;

/// The client validation context.
pub struct MisbehaviourValidationContext<'a> {
    /// Current time in seconds.
    time: u64,
    trusted_consensus_states: HashMap<u64, &'a ConsensusState>,
}

impl<'a> MisbehaviourValidationContext<'a> {
    /// Create a new instance of the client validation context.
    #[must_use]
    pub const fn new(
        time: u64,
        trusted_consensus_states: HashMap<u64, &'a ConsensusState>,
    ) -> Self {
        Self {
            time,
            trusted_consensus_states,
        }
    }
}

impl ClientValidationContext for MisbehaviourValidationContext<'_> {
    type ClientStateRef = ClientStateWrapper;
    type ConsensusStateRef = ConsensusStateWrapper;

    fn consensus_state(
        &self,
        client_cons_state_path: &ibc_core_host_types::path::ClientConsensusStatePath,
    ) -> Result<Self::ConsensusStateRef, HostError> {
        let height = client_cons_state_path.revision_height;
        let trusted_consensus_state = self.trusted_consensus_states[&height];

        Ok(trusted_consensus_state.clone().into())
    }

    fn client_state(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
    ) -> Result<Self::ClientStateRef, HostError> {
        // not needed by the `verify_misbehaviour` function
        unimplemented!()
    }

    fn client_update_meta(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
        _height: &ibc_core_client::types::Height,
    ) -> Result<(Timestamp, ibc_core_client::types::Height), HostError> {
        // not needed by the `verify_misbehaviour` function
        unimplemented!()
    }
}

impl ExtClientValidationContext for MisbehaviourValidationContext<'_> {
    fn host_timestamp(&self) -> Result<Timestamp, HostError> {
        Ok(Timestamp::from_nanoseconds(self.time * 1_000_000_000))
    }

    fn host_height(&self) -> Result<ibc_core_client::types::Height, HostError> {
        // not needed by the `verify_misbehaviour` function
        unimplemented!()
    }

    fn consensus_state_heights(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
    ) -> Result<Vec<ibc_core_client::types::Height>, HostError> {
        // not needed by the `verify_misbehaviour` function
        unimplemented!()
    }

    fn next_consensus_state(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
        _height: &ibc_core_client::types::Height,
    ) -> Result<Option<Self::ConsensusStateRef>, HostError> {
        // not needed by the `verify_misbehaviour` function
        unimplemented!()
    }

    fn prev_consensus_state(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
        _height: &ibc_core_client::types::Height,
    ) -> Result<Option<Self::ConsensusStateRef>, HostError> {
        // not needed by the `verify_misbehaviour` function
        unimplemented!()
    }
}
