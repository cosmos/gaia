//! Defines the [`RelayerModule`] trait that must be implemented by all relayer modules.

use std::marker::{Send, Sync};

use anyhow::Result;

use crate::api::relayer_service_server::RelayerService;

/// The `RelayerModule` trait defines the interface for interacting with a relayer module.
#[tonic::async_trait]
pub trait RelayerModule: Send + Sync + 'static {
    /// Returns the name of the relayer module.
    fn name(&self) -> &'static str;

    /// Creates a relayer service of the given module type with the provided config.
    async fn create_service(&self, config: serde_json::Value) -> Result<Box<dyn RelayerService>>;
}
