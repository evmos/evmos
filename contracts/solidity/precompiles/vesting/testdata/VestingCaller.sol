// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.17;

import "../../common/Types.sol";
import "../Vesting.sol" as vesting;

/// @title VestingCaller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the vesting precompile.
contract VestingCaller {

    /// @dev Defines a method to test creating a new clawback vesting account.
    /// @param funder The address of the account that will fund the vesting account.
    /// @param to The address of the account that will receive the vesting account.
    /// @param enableGovClawback If the vesting account will be subject to governance clawback.
    function createClawbackVestingAccount(
        address funder,
        address to,
        bool enableGovClawback
    ) public {
        bool success = vesting.VESTING_CONTRACT.createClawbackVestingAccount(
            funder,
            to,
            enableGovClawback
        );
        require(success, "VestingCaller: create clawback vesting account failed");
    }

    /// @dev Defines a method to test funding a vesting account
    /// @param funder The address of the account that will fund the vesting account.
    /// @param to The address of the account that will receive the vesting account.
    /// @param startTime The time at which the vesting account will start.
    /// @param lockupPeriods The lockup periods of the vesting account.
    /// @param vestingPeriods The vesting periods of the vesting account.
    function fundVestingAccount(
        address funder,
        address to,
        uint64 startTime,
        vesting.Period[] calldata lockupPeriods,
        vesting.Period[] calldata vestingPeriods
    ) public {
        bool success = vesting.VESTING_CONTRACT.fundVestingAccount(
            funder,
            to,
            startTime,
            lockupPeriods,
            vestingPeriods
        );
        require(success, "VestingCaller: create clawback vesting account failed");
    }

    /// @dev Defines a method to test clawing back coins from a vesting account.
    /// @param funder The address of the account that funded the vesting account.
    /// @param account The address of the vesting account.
    /// @param dest The address of the account that will receive the clawed back coins.
    /// @return coins The coins that were clawed back from the vesting account.
    function clawback(
        address funder,
        address account,
        address dest
    ) public returns (Coin[] memory coins) {
        return vesting.VESTING_CONTRACT.clawback(funder, account, dest);
    }

    /// @dev Defines a method to test updating the funder of a vesting account.
    /// @param funder The address of the account that funded the vesting account.
    /// @param newFunder The address of the new funder of the vesting account.
    /// @param vestingAddr The address of the vesting account.
    function updateVestingFunder(
        address funder,
        address newFunder,
        address vestingAddr
    ) public {
        bool success = vesting.VESTING_CONTRACT.updateVestingFunder(funder, newFunder, vestingAddr);
        require(success, "VestingCaller: update vesting funder failed");
    }

    /// @dev Defines a method to test converting a vesting account to a clawback vesting account.
    /// @param vestingAddr The address of the vesting account.
    function convertVestingAccount(
        address vestingAddr
    ) public {
        bool success = vesting.VESTING_CONTRACT.convertVestingAccount(vestingAddr);
        require(success, "VestingCaller: convert to clawback vesting account failed");
    }

    /// @dev Converts a smart contract address to a vesting account on top of it being a smart contract
    function createClawbackVestingAccountForContract() public {
        bool success = vesting.VESTING_CONTRACT.createClawbackVestingAccount(msg.sender, address(this), false);
        require(success, "VestingCaller: create clawback vesting account for contract failed");
    }

    /// @dev Defines a method to test getting the balances of a vesting account.
    /// @param vestingAddr The address of the vesting account.
    function balances(address vestingAddr) public view returns (
        Coin[] memory locked, 
        Coin[] memory unvested, 
        Coin[] memory vested
    ) {
        return vesting.VESTING_CONTRACT.balances(vestingAddr);
    }
}
