// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IIBCUUPSUpgradeableErrors } from "../errors/IIBCUUPSUpgradeableErrors.sol";
import { ContextUpgradeable } from "@openzeppelin-upgradeable/utils/ContextUpgradeable.sol";
import { UUPSUpgradeable } from "@openzeppelin-contracts/proxy/utils/UUPSUpgradeable.sol";
import { IIBCUUPSUpgradeable } from "../interfaces/IIBCUUPSUpgradeable.sol";

/// @title IBC UUPSUpgradeable contract
/// @notice This contract is an abstract contract for managing upgradability of IBC contracts.
/// @dev This contract is developed with OpenZeppelin's UUPS upgradeable proxy pattern.
/// @dev This contract is meant to be inherited by ICS26Router implementation, and it manages its own upgradability.
/// @dev Other IBC contracts can directly query ICS26Router for the admin addresses to authorize UUPS upgrades (see
/// ICS20Transfer).
/// @dev This contract manages two roles: the timelocked admin, and the governance admin. The timelocked admin
/// represents a timelocked security council, and the governance admin represents an interchain account from the
/// governance of a counterparty chain. The timelocked admin must be set during initialization, and the governance admin
/// should be set later by the timelocked admin.
/// @dev We recommend using `openzeppelin-contracts/contracts/governance/TimelockController.sol` for the timelocked
/// admin
abstract contract IBCUUPSUpgradeable is
    IIBCUUPSUpgradeableErrors,
    IIBCUUPSUpgradeable,
    UUPSUpgradeable,
    ContextUpgradeable
{
    /// @notice Storage of the IBCUUPSUpgradeable contract
    /// @dev It's implemented on a custom ERC-7201 namespace to reduce the risk of storage collisions when using with
    /// upgradeable contracts.
    /// @param timelockedAdmin The timelocked admin address, assumed to be timelocked
    /// @param govAdmin The governance admin address
    struct IBCUUPSUpgradeableStorage {
        address timelockedAdmin;
        address govAdmin;
    }

    /// @notice ERC-7201 slot for the IBCUUPSUpgradeable storage
    /// @dev keccak256(abi.encode(uint256(keccak256("ibc.storage.IBCUUPSUpgradeable")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant IBCUUPSUPGRADEABLE_STORAGE_SLOT =
        0xba83ed17c16070da0debaa680185af188d82c999a75962a12a40699ca48a2b00;

    /// @dev This contract is meant to be initialized with only the timelockedAdmin, and the govAdmin should be set by
    /// the timelockedAdmin later
    /// @dev It makes sense to have the timelockedAdmin not be timelocked until the govAdmin is set
    /// @param timelockedAdmin The timelocked admin address, assumed to be timelocked
    function __IBCUUPSUpgradeable_init(address timelockedAdmin) internal onlyInitializing {
        _getIBCUUPSUpgradeableStorage().timelockedAdmin = timelockedAdmin;
    }

    /// @inheritdoc IIBCUUPSUpgradeable
    function getTimelockedAdmin() external view returns (address) {
        return _getIBCUUPSUpgradeableStorage().timelockedAdmin;
    }

    /// @inheritdoc IIBCUUPSUpgradeable
    function getGovAdmin() external view returns (address) {
        return _getIBCUUPSUpgradeableStorage().govAdmin;
    }

    /// @inheritdoc IIBCUUPSUpgradeable
    function setTimelockedAdmin(address newTimelockedAdmin) external onlyAdmin {
        _getIBCUUPSUpgradeableStorage().timelockedAdmin = newTimelockedAdmin;
    }

    /// @inheritdoc IIBCUUPSUpgradeable
    function setGovAdmin(address newGovAdmin) external onlyAdmin {
        _getIBCUUPSUpgradeableStorage().govAdmin = newGovAdmin;
    }

    /// @inheritdoc IIBCUUPSUpgradeable
    function isAdmin(address account) external view returns (bool) {
        IBCUUPSUpgradeableStorage storage $ = _getIBCUUPSUpgradeableStorage();
        return account == $.timelockedAdmin || account == $.govAdmin;
    }

    /// @inheritdoc UUPSUpgradeable
    function _authorizeUpgrade(address) internal view virtual override onlyAdmin { }
    // solhint-disable-previous-line no-empty-blocks

    /// @notice Returns the storage of the IBCUUPSUpgradeable contract
    function _getIBCUUPSUpgradeableStorage() internal pure returns (IBCUUPSUpgradeableStorage storage $) {
        // solhint-disable-next-line no-inline-assembly
        assembly {
            $.slot := IBCUUPSUPGRADEABLE_STORAGE_SLOT
        }
    }

    /// @notice Modifier to check if the caller is an admin
    modifier onlyAdmin() {
        IBCUUPSUpgradeableStorage storage $ = _getIBCUUPSUpgradeableStorage();
        require(_msgSender() == $.timelockedAdmin || _msgSender() == $.govAdmin, Unauthorized());
        _;
    }
}
