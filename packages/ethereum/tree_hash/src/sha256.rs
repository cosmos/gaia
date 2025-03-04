use std::sync::LazyLock;

use sha2::Digest;

/// Length of a SHA256 hash in bytes.
pub const HASH_LEN: usize = 32;

/// The max index that can be used with `ZERO_HASHES`.
pub const ZERO_HASHES_MAX_INDEX: usize = 48;

pub fn hash_fixed(input: &[u8]) -> [u8; HASH_LEN] {
    sha2::Sha256::digest(input).into()
}

pub fn hash(input: &[u8]) -> Vec<u8> {
    sha2::Sha256::digest(input).into_iter().collect()
}

/// Cached zero hashes where `ZERO_HASHES[i]` is the hash of a Merkle tree with 2^i zero leaves.
pub static ZERO_HASHES: LazyLock<Vec<[u8; HASH_LEN]>> = LazyLock::new(|| {
    let mut hashes = vec![[0; HASH_LEN]; ZERO_HASHES_MAX_INDEX + 1];

    for i in 0..ZERO_HASHES_MAX_INDEX {
        hashes[i + 1] = hash32_concat(&hashes[i], &hashes[i]);
    }

    hashes
});

/// Compute the hash of two slices concatenated.
pub fn hash32_concat(h1: &[u8], h2: &[u8]) -> [u8; 32] {
    let mut ctxt = sha2::Sha256::new();
    ctxt.update(h1);
    ctxt.update(h2);
    ctxt.finalize().into()
}
