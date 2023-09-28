// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../../common/Types.sol";

/// @dev The OSMOSIS contract's address.
address constant OSMOSIS_PRECOMPILED_ADDRESS = 0x0000000000000000000000000000000000000901;

/// @dev The Osmosis contract's instance.
IOsmosisOutpost constant OSMOSIS_CONTRACT = IOsmosisOutpost(OSMOSIS_PRECOMPILED_ADDRESS);

/// @dev Allocation represents a single allocation for an IBC fungible token transfer.
struct Allocation {
    string sourcePort;
    string sourceChannel;
    Coin[] spendLimit;
    string[] allowList;
}

interface IOsmosisOutpost {
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

    /// @dev Emitted when an ICS-20 transfer authorization is granted.
    /// @param grantee The address of the grantee.
    /// @param granter The address of the granter.
    /// @param sourcePort The source port of the IBC transaction.
    /// @param sourceChannel The source channel of the IBC transaction.
    /// @param spendLimit The coins approved in the allocation
    event IBCTransferAuthorization(
        address indexed grantee,
        address indexed granter,
        string sourcePort,
        string sourceChannel,
        Coin[] spendLimit
    );

    /// @dev Approves IBC transfer with a specific amount of tokens.
    /// @param grantee The address for which the transfer authorization is granted.
    /// @param allocations the allocations for the authorization.
    function approve(
        address grantee,
        Allocation[] calldata allocations
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