// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "./IBank.sol";

contract BankCaller {
    function callBalances(address account) external view returns (Balance[] memory balances) {
        return IBANK_CONTRACT.balances(account);
    }

    function callTotalSupply() external view returns (Balance[] memory totalSupply) {
        return IBANK_CONTRACT.totalSupply();
    }

    function callSupplyOf(address erc20Address) external view returns (uint256) {
        return IBANK_CONTRACT.supplyOf(erc20Address);
    }

    /// @dev This function is used to check that both the cosmos and evm state are correctly
    /// updated for a successful transaction or reverted for a failed transaction.
    /// To test this, deploy an ERC20 token contract to chain and mint some tokens to this
    /// contract's address.
    /// This contract will then transfer some tokens to the msg.sender address as well as
    /// run consecutive queries through the bank EVM extension.
    ///
    /// @param _contract Address of the ERC20 to call
    /// @param _address Address of the account to query balances for
    /// @param _amount Amount of tokens to transfer
    /// @return balancesPost The array of native token balances
    function callERC20AndRunQueries(
        address _contract,
        address _address,
        uint256 _amount
    ) public returns (Balance[] memory balancesPost) {
        Balance[] memory balancePre = IBANK_CONTRACT.balances(_address);

        (bool success, ) = _contract.call(
            abi.encodeWithSignature("transfer(address,uint256)", msg.sender, _amount)
        );
        require(success, "transfer failed");

        Balance[] memory balancePost = IBANK_CONTRACT.balances(_address);
        return (balancePost);
    }
}