// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../../common/Types.sol";
import "../../authorization/IICS20Authorization.sol";

/// @dev The OSMOSIS contract's address.
address constant OSMOSIS_PRECOMPILED_ADDRESS = 0x0000000000000000000000000000000000000901;

/// @dev The Osmosis contract's instance.
IOsmosisOutpost constant OSMOSIS_CONTRACT = IOsmosisOutpost(OSMOSIS_PRECOMPILED_ADDRESS);

/// @dev Allocation represents a single allocation for an IBC fungible token transfer.
struct Allocation {
    string   sourcePort;
    string   sourceChannel;
    Coin[]   spendLimit;
    string[] allowList;
}

interface IOsmosisOutpost {
    /// @dev Emitted when an ICS-20 transfer authorization is granted.
    /// @param grantee The address of the grantee.
    /// @param granter The address of the granter.
    /// @param allocations the Allocations authorized with this grant
    event IBCTransferAuthorization(
        address indexed grantee,
        address indexed granter,
        Allocation[] allocations
    );

    /// @dev This event is emitted when an granter revokes a grantee's allowance.
    /// @param grantee The address of the grantee.
    /// @param granter The address of the granter.
    event RevokeIBCTransferAuthorization(
        address indexed grantee,
        address indexed granter
    );

    /// @dev Emitted when a user executes a swap
    /// @param sender The address of the sender
    /// @param receiver The bech32-formatted address of the receiver
    /// of the newly swapped tokens, can be any chain connected to Osmosis
    /// e.g. evmosAddr, junoAddr, cosmosAddr
    event Swap(
        address indexed sender,
        address indexed receiver,
        uint256 amount,
        string baseDenom,
        string outputDenom,
        string chainPrefix
    );

    // @dev Emitted when an ICS-20 transfer is executed.
    /// @param sender The address of the sender.
    /// @param receiver The address of the receiver.
    /// @param sourcePort The source port of the IBC transaction.
    /// @param sourceChannel The source channel of the IBC transaction.
    /// @param denom The denomination of the tokens transferred.
    /// @param amount The amount of tokens transferred.
    /// @param memo The IBC transaction memo.
    event IBCTransfer(
        address indexed sender,
        address indexed receiver,
        string sourcePort,
        string sourceChannel,
        string denom,
        uint256 amount,
        string memo
    );

    /// @dev Approves IBC transfer with a specific amount of tokens to use only with the Osmosis channel.
    /// @param grantee The address for which the transfer authorization is granted.
    /// @param spendLimit The amount of tokens that can be transferred.
    /// @param allowList The list of allowed tokens to be transferred.
    /// @return approved The boolean value indicating whether the operation succeeded.
    function approve(
        address grantee,
        Coin[] calldata spendLimit,
        string[] calldata allowList
    ) external returns (bool approved);

    /// @dev Revokes IBC transfer authorization for a specific grantee.
    /// @param grantee The address for which the transfer authorization will be revoked.
    function revoke(address grantee) external returns (bool revoked);

    /// @dev Returns the remaining number of tokens that a grantee smart contract
    /// will be allowed to spend on behalf of granter through
    /// IBC transfers. This is an empty by array.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param granter The address of the account able to transfer the tokens.
    /// @return allocations The remaining amounts allowed to spend for
    /// corresponding source port and channel.
    function allowance(
        address grantee,
        address granter
    ) external view returns (Allocation[] memory allocations);

    /// @dev Increase the allowance of a given grantee by a specific amount of tokens for IBC transfer methods.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param denom the denomination of the Coin to be transferred to the receiver
    /// @param amount The amount of tokens to be spent.
    /// @return approved is true if the operation ran successfully
    function increaseAllowance(
        address grantee,
        string calldata denom,
        uint256 amount
    ) external returns (bool approved);


    /// @dev Decreases the allowance of a given grantee by a specific amount of tokens for for IBC transfer methods.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
     /// @param denom the denomination of the Coin to be transferred to the receiver
    /// @param amount The amount of tokens to be spent.
    /// @return approved is true if the operation ran successfully
    function decreaseAllowance(
        address grantee,
        string calldata denom,
        uint256 amount
    ) external returns (bool approved);

    // @dev This function is used to swap tokens on Osmosis
    /// @param sender The address of the sender
    /// @param amount The amount of tokens to be swapped
    /// @param receiver The bech32-formatted address of the receiver
    /// of the newly swapped tokens, can be any chain connected to Osmosis
    /// e.g. evmosAddr, junoAddr, cosmosAddr
    /// @param inputDenom The denomination of the tokens to be swapped in ERC20 address format
    /// @param outputDenom The denomination of the tokens to be received in ERC20 address format
    /// @return success The boolean value indicating whether the operation succeeded
    function swap(
        address sender,
        uint256 amount,
        string calldata receiver,
        string calldata inputDenom,
        string calldata outputDenom
    ) external returns (bool success);
}