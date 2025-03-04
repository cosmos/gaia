//! Provides helpers for deriving other types from `LightBlock`.

use ibc_client_tendermint_types::ConsensusState;
use ibc_core_commitment_types::commitment::CommitmentRoot;
use ibc_core_host_types::{error::IdentifierError, identifiers::ChainId};
use ibc_eureka_solidity_types::msgs::{
    IICS02ClientMsgs::Height as SolHeight,
    IICS07TendermintMsgs::{ClientState, SupportedZkAlgorithm, TrustThreshold},
};
use ibc_proto::ibc::{core::client::v1::Height, lightclients::tendermint::v1::Header};
use std::str::FromStr;
use tendermint_light_client_verifier::types::LightBlock;

/// Extension trait for [`LightBlock`] that provides additional methods for converting to other
/// types.
#[allow(clippy::module_name_repetitions)]
pub trait LightBlockExt {
    /// Convert the [`LightBlock`] to a new solidity [`ClientState`].
    ///
    /// # Errors
    /// Returns an error if the chain identifier or height cannot be parsed.
    fn to_sol_client_state(
        &self,
        trust_level: TrustThreshold,
        unbonding_period: u32,
        trusting_period: u32,
        zk_algorithm: SupportedZkAlgorithm,
    ) -> anyhow::Result<ClientState>;
    /// Convert the [`LightBlock`] to a new [`ConsensusState`].
    #[must_use]
    fn to_consensus_state(&self) -> ConsensusState;
    /// Convert the [`LightBlock`] to a new [`Header`].
    ///
    /// # Panics
    /// Panics if the `trusted_height` is zero.
    #[must_use]
    fn into_header(self, trusted_light_block: &LightBlock) -> Header;
    /// Get the chain identifier from the [`LightBlock`].
    ///
    /// # Errors
    /// Returns an error if the chain identifier cannot be parsed.
    fn chain_id(&self) -> Result<ChainId, IdentifierError>;
}

impl LightBlockExt for LightBlock {
    fn to_sol_client_state(
        &self,
        trust_level: TrustThreshold,
        unbonding_period: u32,
        trusting_period: u32,
        zk_algorithm: SupportedZkAlgorithm,
    ) -> anyhow::Result<ClientState> {
        let chain_id = ChainId::from_str(self.signed_header.header.chain_id.as_str())?;
        Ok(ClientState {
            chainId: chain_id.to_string(),
            trustLevel: trust_level,
            latestHeight: SolHeight {
                revisionNumber: chain_id.revision_number().try_into()?,
                revisionHeight: self.height().value().try_into()?,
            },
            isFrozen: false,
            zkAlgorithm: zk_algorithm,
            unbondingPeriod: unbonding_period,
            trustingPeriod: trusting_period,
        })
    }

    fn to_consensus_state(&self) -> ConsensusState {
        ConsensusState {
            timestamp: self.signed_header.header.time,
            root: CommitmentRoot::from_bytes(self.signed_header.header.app_hash.as_bytes()),
            next_validators_hash: self.signed_header.header.next_validators_hash,
        }
    }

    fn into_header(self, trusted_light_block: &LightBlock) -> Header {
        let trusted_revision_number =
            ChainId::from_str(trusted_light_block.signed_header.header.chain_id.as_str())
                .unwrap()
                .revision_number();
        let trusted_block_height = trusted_light_block.height().value();
        Header {
            signed_header: Some(self.signed_header.into()),
            validator_set: Some(self.validators.into()),
            trusted_height: Some(Height {
                revision_number: trusted_revision_number,
                revision_height: trusted_block_height,
            }),
            trusted_validators: Some(trusted_light_block.next_validators.clone().into()),
        }
    }

    fn chain_id(&self) -> Result<ChainId, IdentifierError> {
        ChainId::from_str(self.signed_header.header.chain_id.as_str())
    }
}
