//! Contains types and traits for `verify_header` validation within the program.

use ibc_client_tendermint::{
    client_state::ClientState as ClientStateWrapper,
    consensus_state::ConsensusState as ConsensusStateWrapper, types::ConsensusState,
};
use ibc_core_client::context::{ClientValidationContext, ExtClientValidationContext};
use ibc_core_host_types::error::HostError;
use ibc_primitives::Timestamp;

/// The client validation context.
pub struct ClientValidationCtx<'a> {
    /// Current time in seconds.
    now: u64,
    trusted_consensus_state: &'a ConsensusState,
}

impl<'a> ClientValidationCtx<'a> {
    /// Create a new instance of the client validation context.
    #[must_use]
    pub const fn new(now: u64, trusted_consensus_state: &'a ConsensusState) -> Self {
        Self {
            now,
            trusted_consensus_state,
        }
    }
}

impl ClientValidationContext for ClientValidationCtx<'_> {
    type ClientStateRef = ClientStateWrapper;
    type ConsensusStateRef = ConsensusStateWrapper;

    fn consensus_state(
        &self,
        _client_cons_state_path: &ibc_core_host_types::path::ClientConsensusStatePath,
    ) -> Result<Self::ConsensusStateRef, HostError> {
        // This is the trusted consensus state, whether or not it corresponds to the
        // consensus state path will be checked in solidity.
        Ok(self.trusted_consensus_state.clone().into())
    }

    fn client_state(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
    ) -> Result<Self::ClientStateRef, HostError> {
        // not needed by the `verify_header` function
        unimplemented!()
    }

    fn client_update_meta(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
        _height: &ibc_core_client::types::Height,
    ) -> Result<(Timestamp, ibc_core_client::types::Height), HostError> {
        // not needed by the `verify_header` function
        unimplemented!()
    }
}

impl ExtClientValidationContext for ClientValidationCtx<'_> {
    fn host_timestamp(&self) -> Result<Timestamp, HostError> {
        Ok(Timestamp::from_nanoseconds(self.now * 1_000_000_000))
    }

    fn host_height(&self) -> Result<ibc_core_client::types::Height, HostError> {
        // not needed by the `verify_header` function
        unimplemented!()
    }

    fn consensus_state_heights(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
    ) -> Result<Vec<ibc_core_client::types::Height>, HostError> {
        // not needed by the `verify_header` function
        unimplemented!()
    }

    fn next_consensus_state(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
        _height: &ibc_core_client::types::Height,
    ) -> Result<Option<Self::ConsensusStateRef>, HostError> {
        // not needed by the `verify_header` function
        unimplemented!()
    }

    fn prev_consensus_state(
        &self,
        _client_id: &ibc_core_host_types::identifiers::ClientId,
        _height: &ibc_core_client::types::Height,
    ) -> Result<Option<Self::ConsensusStateRef>, HostError> {
        // not needed by the `verify_header` function
        unimplemented!()
    }
}
