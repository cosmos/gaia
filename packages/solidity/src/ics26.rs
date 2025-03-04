//! Solidity types for ICS26Router.sol

use ibc_proto_eureka::ibc::core::channel::v2::{Packet, Payload};

#[cfg(feature = "rpc")]
alloy_sol_types::sol!(
    #[sol(rpc)]
    #[derive(Debug, PartialEq, Eq)]
    router,
    "../../abi/ICS26Router.json"
);

// NOTE: Some environments won't compile with the `rpc` features.
#[cfg(not(feature = "rpc"))]
alloy_sol_types::sol!(
    #[derive(Debug, PartialEq, Eq)]
    router,
    "../../abi/ICS26Router.json"
);

impl IICS26RouterMsgs::Packet {
    /// Returns the commitment path for the packet.
    #[must_use]
    pub fn commitment_path(&self) -> Vec<u8> {
        let mut path = Vec::new();
        path.extend_from_slice(self.sourceClient.as_bytes());
        path.push(1_u8);
        path.extend_from_slice(&u64::from(self.sequence).to_be_bytes());
        path
    }

    /// Returns the commitment path for the receipt.
    #[must_use]
    pub fn receipt_commitment_path(&self) -> Vec<u8> {
        let mut path = Vec::new();
        path.extend_from_slice(self.destClient.as_bytes());
        path.push(2_u8);
        path.extend_from_slice(&u64::from(self.sequence).to_be_bytes());
        path
    }

    /// Returns the commitment path for the acknowledgement.
    #[must_use]
    pub fn ack_commitment_path(&self) -> Vec<u8> {
        let mut path = Vec::new();
        path.extend_from_slice(self.destClient.as_bytes());
        path.push(3_u8);
        path.extend_from_slice(&u64::from(self.sequence).to_be_bytes());
        path
    }
}

impl TryFrom<Packet> for IICS26RouterMsgs::Packet {
    type Error = <u64 as TryInto<u32>>::Error;

    fn try_from(packet: Packet) -> Result<Self, Self::Error> {
        Ok(Self {
            sequence: packet.sequence.try_into()?,
            sourceClient: packet.source_client,
            destClient: packet.destination_client,
            timeoutTimestamp: packet.timeout_timestamp,
            payloads: packet.payloads.into_iter().map(Into::into).collect(),
        })
    }
}

impl From<IICS26RouterMsgs::Packet> for Packet {
    fn from(packet: IICS26RouterMsgs::Packet) -> Self {
        Self {
            sequence: packet.sequence.into(),
            source_client: packet.sourceClient,
            destination_client: packet.destClient,
            timeout_timestamp: packet.timeoutTimestamp,
            payloads: packet.payloads.into_iter().map(Into::into).collect(),
        }
    }
}

impl From<Payload> for IICS26RouterMsgs::Payload {
    fn from(payload: Payload) -> Self {
        Self {
            sourcePort: payload.source_port,
            destPort: payload.destination_port,
            version: payload.version,
            encoding: payload.encoding,
            value: payload.value.into(),
        }
    }
}

impl From<IICS26RouterMsgs::Payload> for Payload {
    fn from(payload: IICS26RouterMsgs::Payload) -> Self {
        Self {
            source_port: payload.sourcePort,
            destination_port: payload.destPort,
            version: payload.version,
            encoding: payload.encoding,
            value: payload.value.into(),
        }
    }
}
