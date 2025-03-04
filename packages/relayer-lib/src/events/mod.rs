//! Contains the `EurekaEvent` type, which is used to parse Cosmos SDK and EVM IBC Eureka events.

pub mod cosmos_sdk;
mod eureka;

pub use eureka::EurekaEvent;
