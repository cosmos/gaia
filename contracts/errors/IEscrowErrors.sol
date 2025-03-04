// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IEscrowErrors {
    /// @notice Unauthorized function call
    /// @param caller The caller of the function
    error EscrowUnauthorized(address caller);
}
