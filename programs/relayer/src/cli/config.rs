//! Defines the top level configuration for the relayer.

use std::str::FromStr;

use serde_json::Value;
use tracing::Level;

/// The top level configuration for the relayer.
#[derive(Clone, Debug, serde::Deserialize, serde::Serialize)]
#[allow(clippy::module_name_repetitions)]
pub struct RelayerConfig {
    /// The configuration for the relayer modules.
    pub modules: Vec<ModuleConfig>,
    /// The configuration for the relayer server.
    pub server: ServerConfig,
}

/// The configuration for the relayer modules.
#[derive(Clone, Debug, serde::Deserialize, serde::Serialize)]
#[allow(clippy::module_name_repetitions)]
pub struct ModuleConfig {
    /// The name of the module.
    pub name: String,
    /// The source chain identifier for the module.
    /// Used to route requests to the correct module.
    pub src_chain: String,
    /// The destination chain identifier for the module.
    /// Used to route requests to the correct module.
    pub dst_chain: String,
    /// The custom configuration for the module.
    pub config: Value,
    /// Whether the module is enabled.
    #[serde(default = "default_true")]
    pub enabled: bool,
}

/// The configuration for the relayer server.
#[derive(Clone, Debug, serde::Deserialize, serde::Serialize)]
#[allow(clippy::module_name_repetitions)]
pub struct ServerConfig {
    /// The address to bind the server to.
    pub address: String,
    /// The port to bind the server to.
    pub port: u16,
    /// The log level for the server.
    #[serde(default)]
    pub log_level: String,
}

/// Returns true, used as a default value for boolean fields.
const fn default_true() -> bool {
    true
}

impl ServerConfig {
    /// Returns the log level for the server.
    #[must_use]
    pub fn log_level(&self) -> Level {
        Level::from_str(&self.log_level).unwrap_or(Level::INFO)
    }
}
