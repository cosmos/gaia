// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IIBCERC20Errors {
    /// @notice Unauthorized function call
    /// @param caller The caller of the function
    error IBCERC20Unauthorized(address caller);

    /// @notice Minting or burining is only allowed for escrow
    /// @param escrow The escrow contract address
    /// @param mintAddress The address funds are being minted or burned from
    error IBCERC20NotEscrow(address escrow, address mintAddress);
}
