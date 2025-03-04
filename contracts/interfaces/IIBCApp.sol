// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IIBCAppCallbacks } from "../msgs/IIBCAppCallbacks.sol";

/// @title IBC Application Interface
/// @notice IIBCApp is an interface for the IBC Eureka application
interface IIBCApp is IIBCAppCallbacks {
    /// @notice Called when a packet is received from the counterparty chain.
    /// @param msg_ The callback message
    /// @return The acknowledgement data
    function onRecvPacket(OnRecvPacketCallback calldata msg_) external returns (bytes memory);

    /// @notice Called when a packet acknowledgement is received from the counterparty chain.
    /// @param msg_ The callback message
    function onAcknowledgementPacket(OnAcknowledgementPacketCallback calldata msg_) external;

    /// @notice Called when a packet is timed out.
    /// @param msg_ The callback message
    function onTimeoutPacket(OnTimeoutPacketCallback calldata msg_) external;
}
