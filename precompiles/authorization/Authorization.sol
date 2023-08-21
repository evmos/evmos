// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17 .0;

/// @author Evmos Team
/// @title Authorization Interface
/// @dev The interface through which solidity contracts will interact with smart contract approvals.
interface AuthorizationI {
    /// @dev Approves a list of Cosmos or IBC transactions with a specific amount of tokens.
    /// @param grantee The contract address which will have an authorization to spend the origin funds.
    /// @param amount The amount of tokens to be spent.
    /// @param methods The message type URLs of the methods to approve.
    /// @return approved Boolean value to indicate if the approval was successful.
    function approve(
        address grantee,
        uint256 amount,
        string[] calldata methods
    ) external returns (bool approved);

    /// @dev Revokes a list of Cosmos transactions.
    /// @param grantee The contract address which will have its allowances revoked.
    /// @param methods The message type URLs of the methods to revoke.
    /// @return revoked Boolean value to indicate if the revocation was successful.
    function revoke(
        address grantee,
        string[] calldata methods
    ) external returns (bool revoked);

    /// @dev Increase the allowance of a given spender by a specific amount of tokens for IBC
    /// transfer methods or staking.
    /// @param grantee The contract address which allowance will be increased.
    /// @param amount The amount of tokens to be spent.
    /// @param methods The message type URLs of the methods to approve.
    /// @return approved Boolean value to indicate if the approval was successful.
    function increaseAllowance(
        address grantee,
        uint256 amount,
        string[] calldata methods
    ) external returns (bool approved);

    /// @dev Decreases the allowance of a given spender by a specific amount of tokens for IBC
    /// transfer methods or staking.
    /// @param grantee The contract address which allowance will be decreased.
    /// @param amount The amount of tokens to be spent.
    /// @param methods The message type URLs of the methods to approve.
    /// @return approved Boolean value to indicate if the approval was successful.
    function decreaseAllowance(
        address grantee,
        uint256 amount,
        string[] calldata methods
    ) external returns (bool approved);

    /// @dev Returns the remaining number of tokens that spender will be allowed to spend
    /// on behalf of the owner through IBC transfer methods or staking. This is zero by default.
    /// @param grantee The contract address which has the Authorization.
    /// @param granter The account address that grants an Authorization.
    /// @param method The message type URL of the methods for which the approval should be queried.
    /// @return remaining The remaining number of tokens available to be spent.
    function allowance(
        address grantee,
        address granter,
        string calldata method
    ) external view returns (uint256 remaining);

    /// @dev This event is emitted when the allowance of a granter is set by a call to the approve method.
    /// The value field specifies the new allowance and the methods field holds the information for which methods
    /// the approval was set.
    /// @param grantee The contract address that received an Authorization from the granter.
    /// @param granter The account address that granted an Authorization.
    /// @param methods The message type URLs of the methods for which the approval is set.
    /// @param value The amount of tokens approved to be spent.
    event Approval(
        address indexed grantee,
        address indexed granter,
        string[] methods,
        uint256 value
    );

    /// @dev This event is emitted when an owner revokes a spender's allowance.
    /// @param grantee The contract address that has it's Authorization revoked.
    /// @param granter The account address of the granter.
    /// @param methods The message type URLs of the methods for which the approval is set.
    event Revocation(
        address indexed grantee,
        address indexed granter,
        string[] methods
    );

    /// @dev This event is emitted when the allowance of a granter is changed by a call to the decrease or increase
    /// allowance method. The values field specifies the new allowances and the methods field holds the
    /// information for which methods the approval was set.
    /// @param grantee The contract address for which the allowance changed.
    /// @param granter The account address of the granter.
    /// @param methods The message type URLs of the methods for which the approval is set.
    /// @param values The amounts of tokens approved to be spent.
    event AllowanceChange(
        address indexed grantee,
        address indexed granter,
        string[] methods,
        uint256[] values
    );
}
