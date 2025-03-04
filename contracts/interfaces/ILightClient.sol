// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { ILightClientMsgs } from "../msgs/ILightClientMsgs.sol";

/// @title Light Client Interface
/// @notice ILightClient is the light client interface for the IBC Eureka light client
interface ILightClient {
    /// @notice Updating the client and consensus state
    /// @param updateMsg The update message e.g., an SP1 proof and public value pair.
    /// @return The result of the update operation
    function updateClient(bytes calldata updateMsg) external returns (ILightClientMsgs.UpdateResult);

    /// @notice Querying the membership of a key-value pair
    /// @dev Notice that this message is not view, as it may update the client state for caching purposes.
    /// @param msg_ The membership message
    /// @return The unix timestamp of the verification height in the counterparty chain in seconds.
    function verifyMembership(ILightClientMsgs.MsgVerifyMembership calldata msg_) external returns (uint256);

    /// @notice Querying the non-membership of a key
    /// @dev Notice that this message is not view, as it may update the client state for caching purposes.
    /// @param msg_ The membership message
    /// @return The unix timestamp of the verification height in the counterparty chain in seconds.
    function verifyNonMembership(ILightClientMsgs.MsgVerifyNonMembership calldata msg_) external returns (uint256);

    /// @notice Misbehaviour handling, moves the light client to the frozen state if misbehaviour is detected
    /// @param misbehaviourMsg The misbehaviour message
    function misbehaviour(bytes calldata misbehaviourMsg) external;

    /// @notice Upgrading the client
    /// @param upgradeMsg The upgrade message
    function upgradeClient(bytes calldata upgradeMsg) external;

    /// @notice Returns the client state.
    /// @return The client state.
    function getClientState() external view returns (bytes memory);
}
