// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IIBCERC20 } from "../interfaces/IIBCERC20.sol";
import { IIBCERC20Errors } from "../errors/IIBCERC20Errors.sol";

import { ERC20Upgradeable } from "@openzeppelin-upgradeable/token/ERC20/ERC20Upgradeable.sol";

contract IBCERC20 is IIBCERC20Errors, IIBCERC20, ERC20Upgradeable {
    /// @notice Storage of the IBCERC20 contract
    /// @dev It's implemented on a custom ERC-7201 namespace to reduce the risk of storage collisions when using with
    /// upgradeable contracts.
    /// @param _fullDenomPath The full IBC denom path for this token
    /// @param _escrow The escrow contract address
    /// @param _ics20 The ICS20 contract address, can burn and mint tokens
    struct IBCERC20Storage {
        string _fullDenomPath;
        address _escrow;
        address _ics20;
    }

    /// @notice ERC-7201 slot for the IBCERC20 storage
    /// @dev keccak256(abi.encode(uint256(keccak256("ibc.storage.IBCERC20")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant IBCERC20_STORAGE_SLOT = 0x1dd677b5a02f77610493322b5fdbbfdb607b541c6e6045daab3464e895dea800;

    /// @dev This contract is meant to be deployed by a proxy, so the constructor is not used
    constructor() {
        _disableInitializers();
    }

    /// @inheritdoc IIBCERC20
    function initialize(address ics20_, address escrow_, string memory fullDenomPath_) external initializer {
        __ERC20_init(fullDenomPath_, fullDenomPath_);

        IBCERC20Storage storage $ = _getIBCERC20Storage();

        $._fullDenomPath = fullDenomPath_;
        $._escrow = escrow_;
        $._ics20 = ics20_;
    }

    /// @inheritdoc IIBCERC20
    function fullDenomPath() public view returns (string memory) {
        return _getIBCERC20Storage()._fullDenomPath;
    }

    /// @inheritdoc IIBCERC20
    function mint(address mintAddress, uint256 amount) external onlyICS20 {
        require(mintAddress == escrow(), IBCERC20NotEscrow(escrow(), mintAddress));
        _mint(mintAddress, amount);
    }

    /// @inheritdoc IIBCERC20
    function burn(address mintAddress, uint256 amount) external onlyICS20 {
        require(mintAddress == escrow(), IBCERC20NotEscrow(escrow(), mintAddress));
        _burn(mintAddress, amount);
    }

    /// @inheritdoc IIBCERC20
    function escrow() public view returns (address) {
        return _getIBCERC20Storage()._escrow;
    }

    /// @inheritdoc IIBCERC20
    function ics20() external view returns (address) {
        return _getIBCERC20Storage()._ics20;
    }

    /// @notice Returns the storage of the IBCERC20 contract
    function _getIBCERC20Storage() private pure returns (IBCERC20Storage storage $) {
        // solhint-disable-next-line no-inline-assembly
        assembly {
            $.slot := IBCERC20_STORAGE_SLOT
        }
    }

    /// @notice Modifier to check if the caller is the ICS20 contract
    modifier onlyICS20() {
        require(_msgSender() == _getIBCERC20Storage()._ics20, IBCERC20Unauthorized(_msgSender()));
        _;
    }
}
