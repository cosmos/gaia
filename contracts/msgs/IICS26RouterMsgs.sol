// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IICS02ClientMsgs } from "./IICS02ClientMsgs.sol";

interface IICS26RouterMsgs {
    /// @notice Packet struct
    /// @param sequence The sequence number of the packet
    /// @param sourceClient The source client identifier (client id)
    /// @param destClient The destination client identifier
    /// @param timeoutTimestamp The timeout timestamp in the counterparty chain, in unix seconds
    /// @param payloads The packet payloads
    struct Packet {
        uint32 sequence;
        string sourceClient;
        string destClient;
        uint64 timeoutTimestamp;
        Payload[] payloads;
    }

    /// @notice Payload struct
    /// @notice Used in the Packet struct and handled by IBC applications
    /// @param sourcePort The source port identifier
    /// @param destPort The destination port identifier
    /// @param version The application version of the packet data
    /// @param encoding The encoding of the packet date (value)
    /// @param value The packet data
    struct Payload {
        string sourcePort;
        string destPort;
        string version;
        string encoding;
        bytes value;
    }

    /// @notice Message for sending a packet
    /// @dev Submitted by the IBC application
    /// @param sourceClient The source client identifier (client id)
    /// @param timeoutTimestamp The timeout timestamp in unix seconds
    /// @param payload The packet payload
    struct MsgSendPacket {
        string sourceClient;
        uint64 timeoutTimestamp;
        Payload payload;
    }

    /// @notice Message for receiving packets, submitted by relayer
    /// @param packet The packet to be received
    /// @param proofCommitment The proof of the packet commitment
    /// @param proofHeight The proof height
    struct MsgRecvPacket {
        Packet packet;
        bytes proofCommitment;
        IICS02ClientMsgs.Height proofHeight;
    }

    /// @notice Message for acknowledging packets, submitted by relayer
    /// @param packet The packet to be acknowledged
    /// @param acknowledgement The acknowledgement
    /// @param proofAcked The proof of the acknowledgement commitment
    /// @param proofHeight The proof height
    struct MsgAckPacket {
        Packet packet;
        bytes acknowledgement;
        bytes proofAcked;
        IICS02ClientMsgs.Height proofHeight;
    }

    /// @notice Message for timing out packets, submitted by relayer
    /// @param packet The packet to be timed out
    /// @param proofTimeout The proof of the packet commitment
    /// @param proofHeight The proof height
    struct MsgTimeoutPacket {
        Packet packet;
        bytes proofTimeout;
        IICS02ClientMsgs.Height proofHeight;
    }
}
