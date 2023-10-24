// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

/// @dev The IBank contract's address.
address constant IBANK_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000804; 

/// @dev The IBank contract's instance.
IBank constant IBANK_CONTRACT = IBank(IBANK_PRECOMPILE_ADDRESS);

/// @dev Balance specifies the ERC20 contract address and the amount of tokens.
struct Balance {
  /// contractAddress defines the ERC20 contract address.
  address contractAddress;
  /// amount of tokens
  uint256 amount;
}

/**
 * @author Evmos Team
 * @title Bank Interface
 * @dev Interface for querying balances and supply from the Bank module.
 */
interface IBank {
  /// @dev Balances defines a method for retrieving all the native token balances
  /// for a given account.
  /// @param account the address of the account to query balances for
  /// @return balances the array of native token balances
  function balances(address account) external view returns (Balance[] memory balances);

  /// @dev TotalSupply defines a method for retrieving the total supply of all
  /// native tokens.
  /// @return totalSupply the supply as an array of native token balances
  function totalSupply() external view returns (Balance[] memory totalSupply);
}
