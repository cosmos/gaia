// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IICS02ClientMsgs {
    /// @notice Counterparty client information.
    /// @custom:spec
    /// https://github.com/cosmos/ibc/blob/67fe813f7e4ec603a7c5dec35bc654f3b012afda/spec/micro/README.md?plain=1#L91
    /// @param clientId The client identifier from the counterparty chain.
    /// @param merklePrefix The counterparty chain's merkle prefix.
    struct CounterpartyInfo {
        string clientId;
        bytes[] merklePrefix;
    }

    /// @notice Height of the counterparty chain
    /// @param revisionNumber The revision number of the counterparty chain
    /// @param revisionHeight The height of the counterparty chain
    struct Height {
        uint32 revisionNumber;
        uint32 revisionHeight;
    }
}
