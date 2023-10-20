/// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

/// @dev The Osmosis Outpost contract's address.
address constant OSMOSIS_OUTPOST_ADDRESS = 0x0000000000000000000000000000000000000901;

/// @dev The Osmosis Outpost contract's instance.
IOsmosisOutpost constant OSMOSIS_OUTPOST_CONTRACT = IOsmosisOutpost(
    OSMOSIS_OUTPOST_ADDRESS
);

/// @dev The default value used for the slippage_percentage in the swap
string constant DEFAULT_TWAP_SLIPPAGE_PERCENTAGE = "5";
/// @dev the default value used for window_seconds in the swap
uint64 constant DEFAULT_TWAP_WINDOW_SECONDS = 10;

/// @author Evmos Core Team
/// @dev Interface for directly interacting with Osmosis Outpost
interface IOsmosisOutpost {
    /// @dev Emitted when an ICS-20 transfer is executed.
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

    /// @dev Emitted when a user executes a swap
    /// @param sender The address of the sender
    /// @param input The ERC-20 token contract address to be swapped for
    /// @param output The ERC-20 token contract address to be swapped to (received)
    /// @param amount The amount of input tokens to be swapped
    /// @param receiver The bech32-formatted address of the receiver of the newly swapped
    /// tokens, can be any chain connected to Osmosis e.g. evmosAddr, cosmosAddr, ...
    event Swap(
        address indexed sender,
        address indexed input,
        address indexed output,
        uint256 amount,
        string receiver
    );

    /// @dev This function is used to swap tokens on Osmosis
    /// @param sender The address of the sender
    /// @param input The ERC-20 token contract address to be swapped for
    /// @param output The ERC-20 token contract address to be swapped to (received)
    /// @param amount The amount of input tokens to be swapped
    /// @param receiver The bech32-formatted address of the receiver of the newly swapped
    /// tokens. It can be any chain connected to Osmosis e.g. evmos1..., cosmos1..., etc.
    /// @param slippage_percentage The percentage of slippage accepted for the swap.
    /// @param window_seconds The amount of seconds considered to compute TWAP price
    /// @return success The boolean value indicating whether the operation succeeded
    function swap(
        address sender,
        address input,
        address output,
        uint256 amount,
        string calldata slippage_percentage,
        uint64 window_seconds,
        string calldata receiver
    ) external returns (bool success);
}
