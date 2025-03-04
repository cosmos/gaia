//! Defines the account trie and the account type.

use alloy_primitives::{Address, B256};
use hash_db::HashDB;
use memory_db::{HashKey, MemoryDB};
use primitive_types::{H160, H256, U256};
use rlp_derive::RlpDecodable;
use trie_db::{Trie, TrieDBBuilder};

use crate::{
    types::{keccak_256, EthLayout, KeccakHasher},
    TrieDBError,
};

/// A smart contract account.
#[derive(Debug, Clone, RlpDecodable)]
pub struct Account {
    /// The nonce of the account.
    pub nonce: u64,
    /// The balance of the account.
    pub balance: U256,
    /// The storage root of the account.
    pub storage_root: H256,
    /// The code hash of the account.
    pub code_hash: H256,
}

/// Verifies if the `storage_root` of a contract can be verified against the state `root`.
///
/// * `root`: Light client update's (attested/finalized) execution block's state root.
/// * `address`: Address of the contract.
/// * `proof`: Proof of storage.
/// * `storage_root`: Storage root of the contract.
///
/// WARNING: You must not trust the `root` unless you've verified it.
/// # Errors
/// Returns an error if the verification fails.
pub fn verify_account_storage_root(
    root: B256,
    address: Address,
    proof: impl IntoIterator<Item = impl AsRef<[u8]>>,
    storage_root: B256,
) -> Result<(), TrieDBError> {
    let storage_root: H256 = H256(storage_root.into());
    let address: H160 = H160(address.into());

    match get_node(root, address.as_ref(), proof)? {
        Some(account) => {
            let account =
                rlp::decode::<Account>(account.as_ref()).map_err(TrieDBError::RlpDecode)?;
            if account.storage_root != storage_root {
                return Err(TrieDBError::ValueMismatch {
                    expected: storage_root.as_ref().into(),
                    actual: account.storage_root.as_ref().into(),
                });
            }
            Ok(())
        }
        None => Err(TrieDBError::ValueMissing {
            value: address.as_ref().into(),
        })?,
    }
}

fn get_node(
    root: B256,
    key: impl AsRef<[u8]>,
    proof: impl IntoIterator<Item = impl AsRef<[u8]>>,
) -> Result<Option<Vec<u8>>, TrieDBError> {
    let mut db = MemoryDB::<KeccakHasher, HashKey<_>, Vec<u8>>::default();
    proof.into_iter().for_each(|n| {
        db.insert(hash_db::EMPTY_PREFIX, n.as_ref());
    });

    let root: H256 = H256(root.into());

    let trie = TrieDBBuilder::<EthLayout>::new(&db, &root).build();

    trie.get(&keccak_256(key.as_ref()))
        .map_err(|e| TrieDBError::GetTrieNodeFailed(e.to_string()))
}
