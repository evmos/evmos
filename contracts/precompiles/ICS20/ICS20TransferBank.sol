// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

import "./ICS20Transfer.sol";
import "./IICS20Bank.sol";
import "../core/IBCHandler.sol";
import "../core/IBCHost.sol";
import "../core/types/App.sol";

contract ICS20TransferBank is ICS20Transfer {
    IICS20Bank bank;

    constructor(IBCHost host_, IBCHandler ibcHandler_, IICS20Bank bank_) ICS20Transfer(host_, ibcHandler_) {
        bank = bank_;
    }

    function _transferFrom(address sender, address receiver, string memory denom, uint256 amount) override internal returns (bool) {
        try bank.transferFrom(sender, receiver, denom, amount) {
            return true;
        } catch (bytes memory) {
            return false;
        }
    }

    function _mint(address account, string memory denom, uint256 amount) override internal returns (bool) {
        try bank.mint(account, denom, amount) {
            return true;
        } catch (bytes memory) {
            return false;
        }
    }

    function _burn(address account, string memory denom, uint256 amount) override internal returns (bool) {
        try bank.burn(account, denom, amount) {
            return true;
        } catch (bytes memory) {
            return false;
        }
    }

}
