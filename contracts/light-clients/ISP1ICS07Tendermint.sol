// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { ISP1Verifier } from "@sp1-contracts/ISP1Verifier.sol";

/// @title ISP1ICS07Tendermint
/// @notice ISP1ICS07Tendermint is the interface for the ICS07 Tendermint light client
interface ISP1ICS07Tendermint {
    /// @notice Immutable update client program verification key.
    /// @return The verification key for the update client program.
    function UPDATE_CLIENT_PROGRAM_VKEY() external view returns (bytes32);

    /// @notice Immutable membership program verification key.
    /// @return The verification key for the membership program.
    function MEMBERSHIP_PROGRAM_VKEY() external view returns (bytes32);

    /// @notice Immutable update client and membership program verification key.
    /// @return The verification key for the update client and membership program.
    function UPDATE_CLIENT_AND_MEMBERSHIP_PROGRAM_VKEY() external view returns (bytes32);

    /// @notice Immutable misbehaviour program verification key.
    /// @return The verification key for the misbehaviour program.
    function MISBEHAVIOUR_PROGRAM_VKEY() external view returns (bytes32);

    /// @notice Immutable SP1 verifier contract address.
    /// @return The SP1 verifier contract.
    function VERIFIER() external view returns (ISP1Verifier);

    /// @notice Constant allowed clock drift in seconds.
    /// @return The allowed clock drift in seconds.
    function ALLOWED_SP1_CLOCK_DRIFT() external view returns (uint16);

    /// @notice Returns the consensus state keccak256 hash at the given revision height.
    /// @param revisionHeight The revision height.
    /// @return The consensus state at the given revision height.
    function getConsensusStateHash(uint32 revisionHeight) external view returns (bytes32);
}
