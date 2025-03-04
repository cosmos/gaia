//! Ethereum trie database utilities.

#![deny(
    clippy::nursery,
    clippy::pedantic,
    warnings,
    missing_docs,
    unused_crate_dependencies
)]

mod error;
pub mod trie_db;
pub mod types;

pub use error::TrieDBError;
