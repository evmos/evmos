// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../common/Types.sol";

/// @author Evmos Team
/// @title Authorization Interface
/// @dev The interface through which solidity contracts will interact with smart contract approvals.
interface IICS20Authorization {
    /// @dev Emitted when an ICS-20 transfer authorization is granted.
    /// @param grantee The address of the grantee.
    /// @param granter The address of the granter.
    /// @param allocations An array of Allocation authorized with this grant.
    event IBCTransferAuthorization(
        address indexed grantee,
        address indexed granter,
        ICS20Allocation[] allocations
    );

    /// @dev Approves IBC transfer with a specific amount of tokens.
    /// @param grantee The address for which the transfer authorization is granted.
    /// @param allocations An array of Allocation for the authorization.
    function approve(address grantee, ICS20Allocation[] calldata allocations) external returns (bool approved);

    /// @dev Revokes IBC transfer authorization for a specific grantee.
    /// @param grantee The address for which the transfer authorization will be revoked.
    function revoke(address grantee) external returns (bool revoked);

    /// @dev Increase the allowance of a given grantee by a specific amount of tokens for IBC transfer methods.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param sourcePort The port on which the packet will be sent.
    /// @param sourceChannel The channel by which the packet will be sent.
    /// @param denom The denomination of the Coin to be transferred to the receiver.
    /// @param amount The increase in amount of tokens that can be spent.
    /// @return approved Is true if the operation ran successfully.
    function increaseAllowance(
        address grantee,
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount
    ) external returns (bool approved);

    /// @dev Decreases the allowance of a given grantee by a specific amount of tokens for IBC transfer methods.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param sourcePort The port on which the packet will be sent.
    /// @param sourceChannel The channel by which the packet will be sent.
    /// @param denom The denomination of the Coin to be transferred to the receiver.
    /// @param amount The amount by which the spendable tokens are decreased.
    /// @return approved Is true if the operation ran successfully.
    function decreaseAllowance(
        address grantee,
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount
    ) external returns (bool approved);

    /// @dev Returns the remaining number of tokens that a grantee
    /// will be allowed to spend on behalf of granter through
    /// IBC transfers. This is an empty array by default.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param granter The address of the account able to transfer the tokens.
    /// @return allocations The remaining amounts allowed to be spent for
    /// corresponding source port and channel.
    function allowance(
        address grantee,
        address granter
    ) external view returns (ICS20Allocation[] memory allocations);
}