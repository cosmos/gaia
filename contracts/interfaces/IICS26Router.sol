// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IICS26RouterMsgs } from "../msgs/IICS26RouterMsgs.sol";
import { IIBCApp } from "./IIBCApp.sol";

/// @title ICS26 Router Interface
/// @notice IICS26Router is an interface for the IBC Eureka router
interface IICS26Router {
    /// @notice The role identifier for the port customizer role
    /// @dev The port identifier role is used to add IBC applications with custom port identifiers
    /// @return The role identifier
    function PORT_CUSTOMIZER_ROLE() external view returns (bytes32);

    /// @notice Returns the address of the IBC application given the port identifier
    /// @param portId The port identifier
    /// @return The address of the IBC application contract
    function getIBCApp(string calldata portId) external view returns (IIBCApp);

    /// @notice Adds an IBC application to the router
    /// @dev Only the admin can submit non-empty port identifiers.
    /// @dev The default port identifier is the address of the IBC application contract.
    /// @param portId The port identifier, only admin can submit non-empty port identifiers.
    /// @param app The address of the IBC application contract
    function addIBCApp(string calldata portId, address app) external;

    /// @notice Sends a packet
    /// @param msg The message for sending packets
    /// @return The sequence number of the packet
    function sendPacket(IICS26RouterMsgs.MsgSendPacket calldata msg) external returns (uint32);

    /// @notice Receives a packet
    /// @param msg The message for receiving packets
    function recvPacket(IICS26RouterMsgs.MsgRecvPacket calldata msg) external;

    /// @notice Acknowledges a packet
    /// @param msg The message for acknowledging packets
    function ackPacket(IICS26RouterMsgs.MsgAckPacket calldata msg) external;

    /// @notice Timeouts a packet
    /// @param msg The message for timing out packets
    function timeoutPacket(IICS26RouterMsgs.MsgTimeoutPacket calldata msg) external;

    /// @notice Grants the port customizer role to an account
    /// @dev Can only be called by an admin
    /// @param account The account to grant the role to
    function grantPortCustomizerRole(address account) external;

    /// @notice Revokes the port customizer role from an account
    /// @dev Can only be called by an admin
    /// @param account The account to revoke the role from
    function revokePortCustomizerRole(address account) external;

    // --------------------- Events --------------------- //

    /// @notice Emitted when an IBC application is added to the router
    /// @param portId The port identifier
    /// @param app The address of the IBC application contract
    event IBCAppAdded(string portId, address app);
    /// @notice Emitted when an error occurs during the IBC application's recvPacket callback
    /// @param reason The error message
    event IBCAppRecvPacketCallbackError(bytes reason);
    /// @notice Emitted when a packet is sent
    /// @param clientId The source client identifier
    /// @param sequence The sequence number of the packet
    /// @param packet The sent packet
    event SendPacket(string indexed clientId, uint256 indexed sequence, IICS26RouterMsgs.Packet packet);
    /// @notice Emitted when a packet acknowledgement is written
    /// @param clientId The destination client identifier
    /// @param sequence The sequence number of the packet
    /// @param packet The packet that was acknowledged
    /// @param acknowledgements The list of acknowledgements data
    event WriteAcknowledgement(
        string indexed clientId, uint256 indexed sequence, IICS26RouterMsgs.Packet packet, bytes[] acknowledgements
    );
    /// @notice Emitted when a packet is timed out
    /// @param clientId The source client identifier
    /// @param sequence The sequence number of the packet
    /// @param packet The packet that was timed out
    event TimeoutPacket(string indexed clientId, uint256 indexed sequence, IICS26RouterMsgs.Packet packet);
    /// @notice Emitted when a packet is acknowledged
    /// @param clientId The source client identifier
    /// @param sequence The sequence number of the packet
    /// @param packet The packet that was acknowledged
    /// @param acknowledgement The acknowledgement data
    event AckPacket(
        string indexed clientId, uint256 indexed sequence, IICS26RouterMsgs.Packet packet, bytes acknowledgement
    );
    /// @notice Emitted when a redundant relay occurs
    event Noop();
}
