//! Contains the types needed to parse Cosmos SDK's IBC Eureka events.
//!
//! Should be kept in sync with
//! <https://github.com/cosmos/ibc-go/blob/13a13abea09415f2d5c2b4c4ac8edf6b756b8e74/modules/core/04-channel/v2/types/events.go#L9>.

/// The event type for a send packet event.
pub const EVENT_TYPE_SEND_PACKET: &str = "send_packet";
/// The event type for a receive packet event.
pub const EVENT_TYPE_RECV_PACKET: &str = "recv_packet";
/// The event type for a timeout packet event.
pub const EVENT_TYPE_TIMEOUT_PACKET: &str = "timeout_packet";
/// The event type for an acknowledge packet event.
pub const EVENT_TYPE_ACKNOWLEDGE_PACKET: &str = "acknowledge_packet";
/// The event type for a write acknowledgement event.
pub const EVENT_TYPE_WRITE_ACK: &str = "write_acknowledgement";

/// The attribute key for the source client.
pub const ATTRIBUTE_KEY_SRC_CLIENT: &str = "packet_source_client";
/// The attribute key for the destination client.
pub const ATTRIBUTE_KEY_DST_CLIENT: &str = "packet_dest_client";
/// The attribute key for the sequence.
pub const ATTRIBUTE_KEY_SEQUENCE: &str = "packet_sequence";
/// The attribute key for the timeout timestamp.
pub const ATTRIBUTE_KEY_TIMEOUT_TIMESTAMP: &str = "packet_timeout_timestamp";
/// The attribute key for the encoded packet hex.
pub const ATTRIBUTE_KEY_ENCODED_PACKET_HEX: &str = "encoded_packet_hex";
/// The attribute key for the encoded acknowledgement hex.
pub const ATTRIBUTE_KEY_ENCODED_ACK_HEX: &str = "encoded_acknowledgement_hex";
