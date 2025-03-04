#![doc = include_str!("../README.md")]
#![deny(
    clippy::nursery,
    clippy::pedantic,
    warnings,
    missing_docs,
    unused_crate_dependencies
)]

/// Ensure that a condition is true, otherwise return an error.
/// This macro is used for precondition checks in the light client logic for readability.
macro_rules! ensure {
    ($cond:expr, $err:expr) => {
        if !$cond {
            return Err($err);
        }
    };
}

pub mod client_state;
pub mod consensus_state;
pub mod error;
pub mod header;
pub mod membership;
pub mod misbehaviour;
pub mod trie;
pub mod update;
pub mod verify;

#[cfg(any(test, feature = "test-utils"))]
pub mod test_utils;
