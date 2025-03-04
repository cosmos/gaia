// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { ILightClientMsgs } from "../msgs/ILightClientMsgs.sol";
import { IICS02ClientMsgs } from "../msgs/IICS02ClientMsgs.sol";

import { IICS02ClientErrors } from "../errors/IICS02ClientErrors.sol";
import { IICS02Client } from "../interfaces/IICS02Client.sol";
import { ILightClient } from "../interfaces/ILightClient.sol";

import { Strings } from "@openzeppelin-contracts/utils/Strings.sol";
import { AccessControlUpgradeable } from "@openzeppelin-upgradeable/access/AccessControlUpgradeable.sol";

/// @title ICS02 Client contract
/// @notice This contract implements the ICS02 Client Router interface
/// @dev Light client migrations/upgrades are supported via `AccessControl` role-based access control
/// @dev Each client is identified by a unique identifier, hash of which also serves as the role identifier
/// @dev The light client migrator role is granted to whoever called `addClient` for the client, and can be revoked (not
/// transferred)
abstract contract ICS02ClientUpgradeable is IICS02Client, IICS02ClientErrors, AccessControlUpgradeable {
    /// @notice Storage of the ICS02Client contract
    /// @dev It's implemented on a custom ERC-7201 namespace to reduce the
    /// @dev risk of storage collisions when using with upgradeable contracts.
    /// @param clients Mapping of client identifiers to light client contracts
    /// @param counterpartyInfos Mapping of client identifiers to counterparty info
    /// @param nextClientSeq The next sequence number for the next client identifier
    /// @custom:storage-location erc7201:ibc.storage.ICS02Client
    struct ICS02ClientStorage {
        mapping(string clientId => ILightClient) clients;
        mapping(string clientId => IICS02ClientMsgs.CounterpartyInfo info) counterpartyInfos;
        uint256 nextClientSeq;
    }

    /// @notice ERC-7201 slot for the ICS02Client storage
    /// @dev keccak256(abi.encode(uint256(keccak256("ibc.storage.ICS02Client")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant ICS02CLIENT_STORAGE_SLOT =
        0x515a8336edcaab4ae6524d41223c1782132890f89189ba6632107a7b5a449600;

    /// @notice Prefix for the light client migrator roles
    /// @dev The role identifier is driven in _getLightClientMigratorRole
    string private constant MIGRATOR_ROLE_PREFIX = "LIGHT_CLIENT_MIGRATOR_ROLE_";

    /// @notice Prefix for the client identifiers
    string private constant CLIENT_ID_PREFIX = "client-";

    // no need to run any initialization logic
    // solhint-disable-next-line no-empty-blocks
    function __ICS02Client_init() internal onlyInitializing { }

    /// @inheritdoc IICS02Client
    function getNextClientSeq() external view returns (uint256) {
        return _getICS02ClientStorage().nextClientSeq;
    }

    /// @notice Generates the next client identifier
    /// @return The next client identifier
    function nextClientId() private returns (string memory) {
        ICS02ClientStorage storage $ = _getICS02ClientStorage();
        // initial client sequence should be 0, hence we use x++ instead of ++x
        return string.concat(CLIENT_ID_PREFIX, Strings.toString($.nextClientSeq++));
    }

    /// @inheritdoc IICS02Client
    function getCounterparty(string calldata clientId) public view returns (IICS02ClientMsgs.CounterpartyInfo memory) {
        IICS02ClientMsgs.CounterpartyInfo memory counterpartyInfo = _getICS02ClientStorage().counterpartyInfos[clientId];
        require(bytes(counterpartyInfo.clientId).length != 0, IBCCounterpartyClientNotFound(clientId));

        return counterpartyInfo;
    }

    /// @inheritdoc IICS02Client
    function getClient(string calldata clientId) public view returns (ILightClient) {
        ILightClient client = _getICS02ClientStorage().clients[clientId];
        require(address(client) != address(0), IBCClientNotFound(clientId));

        return client;
    }

    /// @inheritdoc IICS02Client
    function addClient(
        IICS02ClientMsgs.CounterpartyInfo calldata counterpartyInfo,
        address client
    )
        external
        returns (string memory)
    {
        ICS02ClientStorage storage $ = _getICS02ClientStorage();

        string memory clientId = nextClientId();
        $.clients[clientId] = ILightClient(client);
        $.counterpartyInfos[clientId] = counterpartyInfo;

        emit ICS02ClientAdded(clientId, counterpartyInfo);

        bytes32 role = getLightClientMigratorRole(clientId);
        require(_grantRole(role, _msgSender()), Unreachable());

        return clientId;
    }

    /// @inheritdoc IICS02Client
    function migrateClient(
        string calldata subjectClientId,
        string calldata substituteClientId
    )
        external
        onlyRole(getLightClientMigratorRole(subjectClientId))
    {
        ICS02ClientStorage storage $ = _getICS02ClientStorage();

        getClient(subjectClientId); // Ensure subject client exists
        ILightClient substituteClient = getClient(substituteClientId);

        getCounterparty(subjectClientId); // Ensure subject client's counterparty exists
        IICS02ClientMsgs.CounterpartyInfo memory substituteCounterpartyInfo = getCounterparty(substituteClientId);

        $.counterpartyInfos[subjectClientId] = substituteCounterpartyInfo;
        $.clients[subjectClientId] = substituteClient;

        emit ICS02ClientMigrated(subjectClientId, substituteClientId);
    }

    /// @inheritdoc IICS02Client
    function updateClient(
        string calldata clientId,
        bytes calldata updateMsg
    )
        external
        returns (ILightClientMsgs.UpdateResult)
    {
        return getClient(clientId).updateClient(updateMsg);
    }

    /// @inheritdoc IICS02Client
    function submitMisbehaviour(string calldata clientId, bytes calldata misbehaviourMsg) external {
        getClient(clientId).misbehaviour(misbehaviourMsg);
        emit ICS02MisbehaviourSubmitted(clientId, misbehaviourMsg);
    }

    /// @inheritdoc IICS02Client
    function upgradeClient(string calldata clientId, bytes calldata upgradeMsg) external {
        getClient(clientId).upgradeClient(upgradeMsg);
    }

    /// @inheritdoc IICS02Client
    function grantLightClientMigratorRole(string calldata clientId, address account) external {
        _authorizeSetLightClientMigratorRole(clientId, account);
        _grantRole(getLightClientMigratorRole(clientId), account);
    }

    /// @inheritdoc IICS02Client
    function revokeLightClientMigratorRole(string calldata clientId, address account) external {
        _authorizeSetLightClientMigratorRole(clientId, account);
        _revokeRole(getLightClientMigratorRole(clientId), account);
    }

    /// @notice Authorizes the granting or revoking of the light client migrator role
    /// @param clientId The client identifier
    /// @param account The account to authorize
    function _authorizeSetLightClientMigratorRole(string calldata clientId, address account) internal virtual;

    /// @notice Returns the storage of the ICS02Client contract
    function _getICS02ClientStorage() private pure returns (ICS02ClientStorage storage $) {
        // solhint-disable-next-line no-inline-assembly
        assembly {
            $.slot := ICS02CLIENT_STORAGE_SLOT
        }
    }

    /// @inheritdoc IICS02Client
    function getLightClientMigratorRole(string memory clientId) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(MIGRATOR_ROLE_PREFIX, clientId));
    }
}
