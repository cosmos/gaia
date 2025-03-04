// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @title SP1 Messages
interface ISP1Msgs {
    /// @notice The SP1 proof that can be submitted to the SP1Verifier contract.
    /// @param vKey The verification key for the program.
    /// @param publicValues The public values for the program.
    /// @param proof The proof for the program.
    struct SP1Proof {
        bytes32 vKey;
        bytes publicValues;
        bytes proof;
    }
}
