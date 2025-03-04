//! This module implements the `BeaconApiClient` to interact with the Ethereum Beacon API.

use ethereum_types::consensus::{
    light_client_header::{LightClientFinalityUpdate, LightClientUpdate},
    spec::Spec,
};
use reqwest::{Client, StatusCode};
use serde::de::DeserializeOwned;
use tracing::debug;

use super::{
    error::{BeaconApiClientError, InternalServerError, NotFoundError},
    response::{Response, Version},
};

const SPEC_PATH: &str = "/eth/v1/config/spec";
const FINALITY_UPDATE_PATH: &str = "/eth/v1/beacon/light_client/finality_update";
const LIGHT_CLIENT_UPDATES_PATH: &str = "/eth/v1/beacon/light_client/updates";

/// The api client for interacting with the Beacon API
#[allow(clippy::module_name_repetitions)]
pub struct BeaconApiClient {
    client: Client,
    base_url: String,
}

impl BeaconApiClient {
    /// Create new `BeaconApiClient`
    #[must_use]
    pub fn new(base_url: String) -> Self {
        Self {
            client: Client::new(),
            base_url,
        }
    }

    /// Fetches the Beacon spec
    /// # Errors
    /// Returns an error if the request fails or the response is not successful deserialized
    pub async fn spec(&self) -> Result<Response<Spec>, BeaconApiClientError> {
        self.get_json(SPEC_PATH).await
    }

    /// Fetches the latest Beacon light client finality update
    /// # Errors
    /// Returns an error if the request fails or the response is not successful deserialized
    pub async fn finality_update(
        &self,
    ) -> Result<Response<LightClientFinalityUpdate, Version>, BeaconApiClientError> {
        self.get_json(FINALITY_UPDATE_PATH).await
    }

    /// Fetches Beacon light client updates starting from a given period
    /// # Errors
    /// Returns an error if the request fails or the response is not successful deserialized
    pub async fn light_client_updates(
        &self,
        start_period: u64,
        count: u64,
    ) -> Result<Vec<Response<LightClientUpdate>>, BeaconApiClientError> {
        self.get_json(&format!(
            "{LIGHT_CLIENT_UPDATES_PATH}?start_period={start_period}&count={count}"
        ))
        .await
    }

    // Helper functions
    #[tracing::instrument(skip_all)]
    async fn get_json<T: DeserializeOwned>(&self, path: &str) -> Result<T, BeaconApiClientError> {
        let url = format!("{}{}", self.base_url, path);

        debug!(%url, "get_json");

        let res = self.client.get(url).send().await?;

        match res.status() {
            StatusCode::OK => {
                let bytes = res.bytes().await?;

                debug!(response = %String::from_utf8_lossy(&bytes), "get_json");

                Ok(serde_json::from_slice(&bytes).map_err(BeaconApiClientError::Json)?)
            }
            StatusCode::NOT_FOUND => Err(BeaconApiClientError::NotFound(
                res.json::<NotFoundError>().await?,
            )),
            StatusCode::INTERNAL_SERVER_ERROR => Err(BeaconApiClientError::Internal(
                res.json::<InternalServerError>().await?,
            )),
            code => Err(BeaconApiClientError::Other {
                code,
                text: res.text().await?,
            }),
        }
    }
}
