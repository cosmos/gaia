// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @title IBC Store Interface
/// @dev Non-view functions can only be called by owner.
interface IIBCStore {
    /// @notice Gets the commitment for a given path.
    /// @param hashedPath The hashed path to get the commitment for.
    /// @return The commitment for the given path.
    function getCommitment(bytes32 hashedPath) external view returns (bytes32);

    /// @notice Checks if a packet receipt exists.
    /// @param clientId The packet destination client identifier.
    /// @param sequence The packet sequence number.
    /// @return True if the packet receipt exists, false otherwise.
    function queryPacketReceipt(string calldata clientId, uint64 sequence) external view returns (bool);

    /// @notice Returns the packet commitment for a given packet.
    /// @param clientId The packet source client identifier.
    /// @param sequence The packet sequence number.
    /// @return The packet commitment for the given packet.
    function queryPacketCommitment(string calldata clientId, uint64 sequence) external view returns (bytes32);

    /// @notice Returns the packet acknowledgement commitment for a given packet.
    /// @param clientId The packet destination client identifier.
    /// @param sequence The packet sequence number.
    /// @return The packet acknowledgement commitment for the given packet.
    function queryAckCommitment(string calldata clientId, uint64 sequence) external view returns (bytes32);
}
