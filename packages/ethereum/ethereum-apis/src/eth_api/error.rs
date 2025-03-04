//! This module defines errors for `EthApiClient`.

use alloy::transports::TransportError;

#[derive(Debug, thiserror::Error)]
#[allow(missing_docs, clippy::module_name_repetitions)]
pub enum EthGetProofError {
    #[error("provider error: {0}")]
    ProviderError(#[from] TransportError),

    #[error("parse error trying to parse {0}, {1}")]
    ParseError(String, String),
}
