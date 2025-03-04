//! This module contains the custom `CosmWasm` query for the Ethereum light client

use alloy_primitives::B256;
use cosmwasm_std::{Binary, CustomQuery, QuerierWrapper, QueryRequest};
use ethereum_light_client::verify::BlsVerify;
use ethereum_types::consensus::bls::{BlsPublicKey, BlsSignature};
use thiserror::Error;

/// The custom query for the Ethereum light client
/// This is used to verify BLS signatures in `CosmosSDK`
#[derive(serde::Serialize, serde::Deserialize, Clone)]
#[serde(rename_all = "snake_case")]
#[allow(clippy::module_name_repetitions)]
pub enum EthereumCustomQuery {
    /// Verify a BLS signature
    AggregateVerify {
        /// The public keys to verify the signature
        public_keys: Vec<Binary>,
        /// The message to verify
        message: Binary,
        /// The signature to verify
        signature: Binary,
    },
    /// Aggregate public keys
    Aggregate {
        /// The public keys to aggregate
        public_keys: Vec<Binary>,
    },
}

impl CustomQuery for EthereumCustomQuery {}

/// The BLS verifier via [`EthereumCustomQuery`]
pub struct BlsVerifier<'a> {
    /// The `CosmWasm` querier
    pub querier: QuerierWrapper<'a, EthereumCustomQuery>,
}

/// The error type for the BLS verifier
#[derive(Error, Debug)]
#[allow(missing_docs)]
pub enum BlsVerifierError {
    #[error("fast aggregate verify error: {0}")]
    FastAggregateVerify(String),

    #[error("signature cannot be verified (public_keys: {public_keys:?}, msg: {msg}, signature: {signature})", msg = hex::encode(.msg))]
    InvalidSignature {
        /// The public keys used to verify the signature
        public_keys: Vec<BlsPublicKey>,
        /// The message that was signed
        msg: B256,
        /// The signature that was verified
        signature: BlsSignature,
    },
}

impl BlsVerify for BlsVerifier<'_> {
    type Error = BlsVerifierError;

    fn fast_aggregate_verify(
        &self,
        public_keys: &[BlsPublicKey],
        msg: B256,
        signature: BlsSignature,
    ) -> Result<(), Self::Error> {
        let binary_public_keys: Vec<Binary> = public_keys
            .iter()
            .map(|p| Binary::from(p.to_vec()))
            .collect();

        let request: QueryRequest<EthereumCustomQuery> =
            QueryRequest::Custom(EthereumCustomQuery::AggregateVerify {
                public_keys: binary_public_keys,
                message: Binary::from(msg.to_vec()),
                signature: Binary::from(signature.to_vec()),
            });

        let is_valid: bool = self
            .querier
            .query(&request)
            .map_err(|e| BlsVerifierError::FastAggregateVerify(e.to_string()))?;

        if !is_valid {
            return Err(BlsVerifierError::InvalidSignature {
                public_keys: public_keys.to_vec(),
                msg,
                signature,
            });
        }

        Ok(())
    }
}
