// SPDX-License-Identifier: ENCL-1.0
pragma solidity >=0.8.18;
  
/// @author Evmos Core Team
/// @dev Interface for directly interacting with Osmosis Outpost
interface IOsmosisOutpost {
	/// @dev Emitted when a user executes a swap
    /// @param sender The address of the sender
    /// @param input The ERC-20 token contract address to be swapped for
    /// @param output The ERC-20 token contract address to be swapped to (received)
    /// @param amount The amount of input tokens to be swapped
		/// @param receiver The bech32-formatted address of the receiver 
		/// of the newly swapped tokens, can be any chain connected to Osmosis
		/// e.g. evmosAddr, junoAddr, cosmosAddr
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
    /// @param receiver The bech32-formatted address of the receiver
    /// of the newly swapped tokens, can be any chain connected to Osmosis
    /// e.g. evmosAddr, junoAddr, cosmosAddr
    /// @return success The boolean value indicating whether the operation succeeded
    function swap(
        address sender,
        address input,
        address output,
        uint256 amount,
        string calldata receiver
    ) external returns (bool success);
}