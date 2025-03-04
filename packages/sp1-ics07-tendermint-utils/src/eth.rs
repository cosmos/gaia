//! Helpers for interacting with EVM.

use std::env;

use alloy_network::EthereumWallet;
use alloy_signer_local::PrivateKeySigner;

/// Create an Ethereum wallet from the `PRIVATE_KEY` environment variable.
///
/// # Panics
/// Panics if the `PRIVATE_KEY` environment variable is not set.
/// Panics if the `PRIVATE_KEY` environment variable is not a valid private key.
#[must_use]
pub fn wallet_from_env() -> EthereumWallet {
    let mut private_key = env::var("PRIVATE_KEY").expect("PRIVATE_KEY not set");
    if let Some(stripped) = private_key.strip_prefix("0x") {
        private_key = stripped.to_string();
    }

    let signer: PrivateKeySigner = private_key.parse().unwrap();
    EthereumWallet::from(signer)
}
