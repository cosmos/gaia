// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @title IBC Identifiers
/// @notice Utilities for validating IBC identifiers
library IBCIdentifiers {
    /// @notice validatePortIdentifier checks if the port identifier is allowed
    /**
     * @dev validatePortIdentifier validates a port identifier string
     *     check if the string consist of characters in one of the following categories only:
     *     - Alphanumeric
     *     - `.`, `_`, `+`, `-`, `#`
     *     - `[`, `]`, `<`, `>`
     */
    /// @custom:url https://github.com/hyperledger-labs/yui-ibc-solidity/blob/49d88ae8151a92e086e6ca7d27a2d3651889edff/
    /// contracts/core/26-router/IBCModuleManager.sol#L123
    /// @param portId The port identifier
    /// @return True if the port identifier is valid
    function validatePortIdentifier(bytes memory portId) internal pure returns (bool) {
        if (portId.length < 2 || portId.length > 128) {
            return false;
        }
        unchecked {
            for (uint256 i = 0; i < portId.length; i++) {
                uint256 c = uint256(uint8(portId[i]));
                if (
                    // a-z
                    // 0-9
                    // A-Z
                    // ".", "_", "+", "-"
                    // "#", "[", "]", "<", ">"
                    (c >= 0x61 && c <= 0x7A) || (c >= 0x30 && c <= 0x39) || (c >= 0x41 && c <= 0x5A)
                        || (c == 0x2E || c == 0x5F || c == 0x2B || c == 0x2D)
                        || (c == 0x23 || c == 0x5B || c == 0x5D || c == 0x3C || c == 0x3E)
                ) {
                    continue;
                }
                return false;
            }
        }
        return true;
    }
}
