// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IICS02ClientMsgs } from "./IICS02ClientMsgs.sol";

interface ILightClientMsgs {
    /// @notice Message for querying the membership of a key-value pair in the Merkle root at a given height.
    /// @param proof The proof
    /// @param proofHeight The height of the proof
    /// @param path The path of the value in the Merkle tree
    /// @param value The value in the Merkle tree
    struct MsgVerifyMembership {
        bytes proof;
        IICS02ClientMsgs.Height proofHeight;
        bytes[] path;
        bytes value;
    }

    /// @notice Message for querying the non-membership of a key in the Merkle root at a given height.
    /// @param proof The proof
    /// @param proofHeight The height of the proof
    /// @param path The path of the value in the Merkle tree
    struct MsgVerifyNonMembership {
        bytes proof;
        IICS02ClientMsgs.Height proofHeight;
        bytes[] path;
    }

    /// @notice The result of an update operation
    enum UpdateResult {
        /// The update was successful
        Update,
        /// A misbehaviour was detected
        Misbehaviour,
        /// Client is already up to date
        NoOp
    }
}
