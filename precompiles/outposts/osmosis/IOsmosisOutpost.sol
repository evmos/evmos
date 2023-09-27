// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

/// @dev The OSMOSIS contract's address.
address constant OSMOSIS_PRECOMPILED_ADDRESS = 0x0000000000000000000000000000000000000901;

/// @dev The Osmosis contract's instance.
IOsmosisOutpost constant OSMOSIS_CONTRACT = IOsmosisOutpost(OSMOSIS_PRECOMPILED_ADDRESS);

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

    // @dev This function is used to swap tokens on Osmosis
    /// @param amount The amount of tokens to be swapped
    /// @param receiver The bech32-formatted address of the receiver
    /// of the newly swapped tokens, can be any chain connected to Osmosis
    /// e.g. evmosAddr, junoAddr, cosmosAddr
    /// @param baseDenom The denomination of the tokens to be swapped in ERC20 address format
    /// @param outputDenom The denomination of the tokens to be received in ERC20 address format
    /// @return success The boolean value indicating whether the operation succeeded
    function swap(
        uint256 amount,
        string calldata receiver,
        address baseDenom,
        address outputDenom
    ) external returns (bool success);
}