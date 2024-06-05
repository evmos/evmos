// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../IBank.sol";

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
}