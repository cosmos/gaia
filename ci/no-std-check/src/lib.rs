// ensure_no_std/src/main.rs
#![no_std]
#![allow(unused_imports)]

// Import the crates that we want to check if they are fully no-std compliance

// #[cfg(feature = "ibc")]
// use ibc;

// #[cfg(feature = "ibc_proto")]
// use ibc_proto;

#[cfg(feature = "sp-core")]
use sp_core;

#[cfg(feature = "sp-io")]
use sp_io;

#[cfg(feature = "sp-runtime")]
use sp_runtime;

#[cfg(feature = "sp-std")]
use sp_std;

// Supported Imports

use bytes;
use chrono;
use contracts;
use crossbeam_channel;
use ed25519;
use ed25519_dalek;
use flex_error;
use futures;
use impl_serde;
use k256;
use num_derive;
use num_traits;
use once_cell;
use prost;
use prost_types;
use ripemd160;
use ryu;
use serde;
use serde_bytes;
use serde_derive;
use serde_json;
use serde_repr;
use serde_cbor;
use serde_json_core;
use sha2;
use signature;
use static_assertions;
use subtle;
use subtle_encoding;
use time;
use tracing;
use zeroize;

// Unsupported Imports

#[cfg(feature = "tonic")]
use tonic;

#[cfg(feature = "socket2")]
use socket2;

#[cfg(feature = "ics23")]
use ics23;

#[cfg(feature = "getrandom")]
use getrandom;

#[cfg(feature = "thiserror")]
use thiserror;

#[cfg(feature = "regex")]
use regex;

#[cfg(feature = "sled")]
use sled;

#[cfg(feature = "tokio")]
use tokio;

#[cfg(feature = "toml")]
use toml;

#[cfg(feature = "url")]
use url;


use core::panic::PanicInfo;

/*

This function definition checks for the compliance of no-std in
dependencies by causing a compile error if  this crate is
linked with `std`. When that happens, you should see error messages
such as follows:

```
error[E0152]: found duplicate lang item `panic_impl`
  --> no-std-check/src/lib.rs
   |
12 | fn panic(_info: &PanicInfo) -> ! {
   | ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
   |
   = note: the lang item is first defined in crate `std` (which `offending-crate` depends on)
```

 */
#[cfg(not(feature = "use-substrate"))]
#[panic_handler]
#[no_mangle]
fn panic(_info: &PanicInfo) -> ! {
    loop {}
}
