// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.17;

import "../../common/Types.sol";
import "../VestingI.sol" as vesting;
import "../../testutil/contracts/ICounter.sol";

/// @title VestingCaller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the vesting precompile.
contract VestingCaller {
    /// counter is used to test the state persistence bug, when EVM and Cosmos state were both
    /// changed in the same function.
    uint256 public counter;
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
        require(
            success,
            "VestingCaller: create clawback vesting account failed"
        );
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
        counter++;
        bool success = vesting.VESTING_CONTRACT.fundVestingAccount(
            funder,
            to,
            startTime,
            lockupPeriods,
            vestingPeriods
        );
        require(
            success,
            "VestingCaller: create clawback vesting account failed"
        );
        counter--;
    }

    /// @dev Defines a method to test funding a vesting account.
    /// If specified, it sends 15 aevmos to the funder before and/or after
    /// the precompile call
    /// @param funder The address of the account that will fund the vesting account.
    /// @param to The address of the account that will receive the vesting account.
    /// @param _transferTo The address to send some funds to.
    /// @param startTime The time at which the vesting account will start.
    /// @param lockupPeriods The lockup periods of the vesting account.
    /// @param vestingPeriods The vesting periods of the vesting account.
    /// @param transferBefore A boolean to specify if the contract should transfer
    /// funds to the funder before the precompile call.
    /// @param transferAfter A boolean to specify if the contract should transfer
    /// funds to the funder after the precompile call.
    function fundVestingAccountAndTransfer(
        address payable funder,
        address to,
        address payable _transferTo,
        uint64 startTime,
        vesting.Period[] calldata lockupPeriods,
        vesting.Period[] calldata vestingPeriods,
        bool transferBefore,
        bool transferAfter
    ) public {
        if (transferBefore) {
            counter++;
            (bool sent, ) = _transferTo.call{value: 15}("");
            require(sent, "Failed to send Ether to funder");
        }
        bool success = vesting.VESTING_CONTRACT.fundVestingAccount(
            funder,
            to,
            startTime,
            lockupPeriods,
            vestingPeriods
        );
        require(
            success,
            "VestingCaller: create clawback vesting account failed"
        );
        if (transferAfter) {
            (bool sent, ) = _transferTo.call{value: 15}("");
            require(sent, "Failed to send Ether to funder");
            counter++;
        }
    }

    /// @dev Defines a method to test funding a vesting account
    /// @param funder The address of the Counter contract that will fund the vesting account.
    /// @param to The address of the account that will receive the vesting account.
    /// @param startTime The time at which the vesting account will start.
    /// @param lockupPeriods The lockup periods of the vesting account.
    /// @param vestingPeriods The vesting periods of the vesting account.
    function fundVestingAccountWithCounterContract(
        address funder,
        address to,
        uint64 startTime,
        vesting.Period[] calldata lockupPeriods,
        vesting.Period[] calldata vestingPeriods
    ) public {
        ICounter counterContract = ICounter(funder);
        counterContract.add();
        counter++;
        bool success = vesting.VESTING_CONTRACT.fundVestingAccount(
            funder,
            to,
            startTime,
            lockupPeriods,
            vestingPeriods
        );
        require(
            success,
            "VestingCaller: create clawback vesting account failed"
        );
        counterContract.subtract();
        counter--;
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
        counter++;
        coins = vesting.VESTING_CONTRACT.clawback(funder, account, dest);
        counter--;
        return coins;
    }

    /// @dev Defines a method to test clawing back coins from a vesting account.
    /// @param funder The address of the account that funded the vesting account.
    /// @param account The address of the vesting account.
    /// @param dest The address of the account that will receive the clawed back coins.
    /// @param _transferTo The address to send some funds to.
    /// @param _before Boolean to specify if funds should be transferred to _transferTo before the precompile call
    /// @param _after Boolean to specify if funds should be transferred to _transferTo after the precompile call    
    /// @return coins The coins that were clawed back from the vesting account.
    function clawbackWithTransfer(
        address funder,
        address account,
        address dest,
        address payable _transferTo,
        bool _before,
        bool _after
    ) public returns (Coin[] memory coins) {
        if (_before) {
            counter++;
            if (dest != address(this)) {
                (bool sent, ) = _transferTo.call{value: 15}("");
                require(sent, "Failed to send Ether to delegator");
            }
        }
        coins = vesting.VESTING_CONTRACT.clawback(funder, account, dest);
        if (_after) {
            counter++;
            if (dest != address(this)) {
                (bool sent, ) = _transferTo.call{value: 15}("");
                require(sent, "Failed to send Ether to delegator");
            }
        }
        return coins;
    }

    /// @dev Defines a method to test clawing back coins from a vesting account.
    /// It is used for testing the state revert.
    /// @param funder The address of the account that funded the vesting account.
    /// @param account The address of the vesting account.
    /// @param dest The address of the account that will receive the clawed back coins.
    /// @param before Boolean to specify if should revert before counter change.
    /// @return coins The coins that were clawed back from the vesting account.
    function clawbackWithRevert(
        address funder,
        address account,
        address dest,
        bool before
    ) public returns (Coin[] memory coins) {
        counter++;
        coins = vesting.VESTING_CONTRACT.clawback(funder, account, dest);
        if (before) {
            require(false, "revert here");
        }
        counter--;
        require(false, "revert here");
        return coins;
    }

    /// @dev Defines a method to test clawing back coins from a vesting account.
    /// @param funder The address of the account that funded the vesting account.
    /// @param account The address of the vesting account.
    /// @param dest The address of the Counter smart contract that will receive the clawed back coins.
    /// @return coins The coins that were clawed back from the vesting account.
    function clawbackWithCounterCall(
        address funder,
        address account,
        address dest
    ) public returns (Coin[] memory coins) {
        ICounter counterContract = ICounter(dest);
        counterContract.add();
        counter++;
        coins = vesting.VESTING_CONTRACT.clawback(funder, account, dest);
        counterContract.subtract();
        counter--;
        return coins;
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
        bool success = vesting.VESTING_CONTRACT.updateVestingFunder(
            funder,
            newFunder,
            vestingAddr
        );
        require(success, "VestingCaller: update vesting funder failed");
    }

    /// @dev Defines a method to test converting a vesting account to a clawback vesting account.
    /// @param vestingAddr The address of the vesting account.
    function convertVestingAccount(address vestingAddr) public {
        bool success = vesting.VESTING_CONTRACT.convertVestingAccount(
            vestingAddr
        );
        require(
            success,
            "VestingCaller: convert to clawback vesting account failed"
        );
    }

    /// @dev Converts a smart contract address to a vesting account on top of it being a smart contract
    function createClawbackVestingAccountForContract() public {
        bool success = vesting.VESTING_CONTRACT.createClawbackVestingAccount(
            msg.sender,
            address(this),
            false
        );
        require(
            success,
            "VestingCaller: create clawback vesting account for contract failed"
        );
    }

    /// @dev Defines a method to test getting the balances of a vesting account.
    /// @param vestingAddr The address of the vesting account.
    function balances(
        address vestingAddr
    )
        public
        view
        returns (
            Coin[] memory locked,
            Coin[] memory unvested,
            Coin[] memory vested
        )
    {
        return vesting.VESTING_CONTRACT.balances(vestingAddr);
    }
}
