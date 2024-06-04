// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

/// @dev The ITokenFactory contract's address.
address constant TOKEN_FACTORY_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000900;

/// @dev The ITokenFactory contract's instance.
ITokenFactory constant TOKEN_FACTORY_CONTRACT = ITokenFactory(TOKEN_FACTORY_PRECOMPILE_ADDRESS);


/// @author Evmos Team
/// @title ICS20 Transfer Precompiled Contract
/// @dev The interface through which solidity contracts will interact with IBC Transfer (ICS20)
/// @custom:address 0x0000000000000000000000000000000000000900
interface ITokenFactory {

    /// @dev Creates a native Coin and an ERC20 extension for it.
    /// @dev This method creates a token pair with the native coin and an ERC20 extension.
    /// @param name The name of the token.
    /// @param symbol The symbol of the token.
    /// @param decimals The number of decimals of the token.
    /// @param initialSupply The initial supply of the token.
    /// @return success true if the transfer was successful, false otherwise.
    function createERC20(
        string memory name,
        string memory symbol,
        uint8 decimals,
        uint256 initialSupply
    ) external returns (bool success);


    /// @dev Creates a native Coin and an ERC20 extension for it.
    /// @dev This method creates a token pair with the native coin and an ERC20 extension.
    /// @param name The name of the token.
    /// @param symbol The symbol of the token.
    /// @param decimals The number of decimals of the token.
    /// @param initialSupply The initial supply of the token.
    /// @param salt The salt for the deterministic address generation.
    /// @return success true if the transfer was successful, false otherwise.
    function create2ERC20(
        string memory name,
        string memory symbol,
        uint8 decimals,
        uint256 initialSupply,
        bytes32 salt
    ) external returns (bool success);
}
