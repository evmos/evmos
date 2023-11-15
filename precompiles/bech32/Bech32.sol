// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

/// @dev The Bech32I contract's address.
address constant Bech32_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000400;

/// @author Evmos Team
/// @title Bech32 Precompiled Contract
/// @dev The interface through which solidity contracts can convert addresses from
/// hex to bech32 and vice versa.
/// @custom:address 0x0000000000000000000000000000000000000400
interface Bech32I {
    /// @dev Defines a method for converting a hex formatted address to bech32.
    /// @param addr The hex address to be converted.
    /// @param prefix The human readable prefix (HRP) of the bech32 address.
    /// @return bech32Address The address in bech32 format.
    function hexToBech32(
        address addr,
        string memory prefix
    ) external returns (string memory bech32Address);

    /// @dev Defines a method for converting a bech32 formatted address to hex.
    /// @param bech32Address The bech32 address to be converted.
    /// @return addr The address in hex format.
    function bech32ToHex(
        string memory bech32Address
    ) external returns (address addr);
}
