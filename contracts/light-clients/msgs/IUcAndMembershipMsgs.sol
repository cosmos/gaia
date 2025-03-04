// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import { IUpdateClientMsgs } from "./IUpdateClientMsgs.sol";
import { IMembershipMsgs } from "./IMembershipMsgs.sol";

/// @title Update Client and Membership Program Messages
/// @author srdtrk
/// @notice Defines shared types for the update client and membership program.
interface IUpdateClientAndMembershipMsgs {
    /// @notice The public value output for the sp1 update client and membership program.
    /// @param updateClientOutput The output of the update client program.
    /// @param kvPairs The key-value pairs verified by the membership program in the proposed header.
    struct UcAndMembershipOutput {
        IUpdateClientMsgs.UpdateClientOutput updateClientOutput;
        IMembershipMsgs.KVPair[] kvPairs;
    }
}
