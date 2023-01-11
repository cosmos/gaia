/// Params struct
#[derive(Clone, PartialEq, Eq, ::prost::Message)]
pub struct Params {
    /// The lockup module is engaged if locked is true (chain is "locked up")
    #[prost(bool, tag="1")]
    pub locked: bool,
    /// Addresses not affected by the lockup module
    #[prost(string, repeated, tag="2")]
    pub lock_exempt: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    /// Messages with one of these types are blocked when the chain is locked up
    /// and not sent from a lock_exempt address
    #[prost(string, repeated, tag="3")]
    pub locked_message_types: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[derive(Clone, PartialEq, Eq, ::prost::Message)]
pub struct GenesisState {
    #[prost(message, optional, tag="1")]
    pub params: ::core::option::Option<Params>,
}
