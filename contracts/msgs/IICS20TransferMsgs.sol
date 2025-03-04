// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IICS20TransferMsgs {
    /// @notice Message for sending a transfer
    /// @param denom The address of the ERC20 token contract, used as the denomination
    /// @param amount The amount of tokens to transfer
    /// @param receiver The receiver of the transfer on the counterparty chain
    /// @param sourceClient The source client identifier
    /// @param destPort The destination port on the counterparty chain
    /// @param timeoutTimestamp The absolute timeout timestamp in unix seconds
    /// @param memo Optional memo
    struct SendTransferMsg {
        address denom;
        uint256 amount;
        string receiver;
        string sourceClient;
        string destPort;
        uint64 timeoutTimestamp;
        string memo;
    }

    /// @notice FungibleTokenPacketData is the payload for a fungible token transfer packet.
    /// @dev PacketData is defined in
    /// [ICS-20](https://github.com/cosmos/ibc/tree/main/spec/app/ics-020-fungible-token-transfer).
    /// @param denom The denomination of the token
    /// @param sender The sender of the token
    /// @param receiver The receiver of the token
    /// @param amount The amount of tokens
    /// @param memo Optional memo
    struct FungibleTokenPacketData {
        string denom;
        string sender;
        string receiver;
        uint256 amount;
        string memo;
    }
}
