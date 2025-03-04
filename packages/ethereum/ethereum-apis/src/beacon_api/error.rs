//! This module defines errors for `BeaconApiClient`.

use reqwest::StatusCode;
use serde::{Deserialize, Serialize};

/// The error type for the beacon api client.
#[derive(Debug, thiserror::Error)]
#[allow(missing_docs, clippy::module_name_repetitions)]
pub enum BeaconApiClientError {
    #[error("http error: {0}")]
    Http(#[from] reqwest::Error),

    #[error("json deserialization error: {0}")]
    Json(#[from] serde_json::Error),

    #[error("not found: {0}")]
    NotFound(#[from] NotFoundError),

    #[error("internal error: {0}")]
    Internal(#[from] InternalServerError),

    #[error("unknown error ({code}): {text}")]
    Other { code: StatusCode, text: String },
}

/// The not found error structure returned by the Beacon API.
#[derive(Debug, Serialize, Deserialize, thiserror::Error)]
#[error("{status_code} {error}: {message}")]
#[allow(missing_docs, clippy::module_name_repetitions)]
pub struct NotFoundError {
    #[serde(rename = "statusCode")]
    pub status_code: u64,
    pub error: String,
    pub message: String,
}

/// The internal server error returned by the Beacon API.
#[derive(Debug, Serialize, Deserialize, thiserror::Error)]
#[error("{status_code} {error}: {message}")]
#[allow(missing_docs, clippy::module_name_repetitions)]
pub struct InternalServerError {
    #[serde(rename = "statusCode")]
    pub status_code: u64,
    pub error: String,
    pub message: String,
}
