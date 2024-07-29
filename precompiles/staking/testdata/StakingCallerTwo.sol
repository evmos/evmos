// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../StakingI.sol" as staking;

/// @title StakingCaller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the staking precompile.
contract StakingCallerTwo {
    /// counter is used to test the state persistence bug, when EVM and Cosmos state were both
    /// changed in the same function.
    uint256 public counter;

    /// @dev This function calls the staking precompile's approve method.
    /// @param _addr The address to approve.
    /// @param _methods The methods to approve.
    function testApprove(
        address _addr,
        string[] calldata _methods,
        uint256 _amount
    ) public {
        bool success = staking.STAKING_CONTRACT.approve(
            _addr,
            _amount,
            _methods
        );
        require(success, "Failed to approve staking methods");
    }

    /// @dev This function showcased, that there was a bug in the EVM implementation, that occurred when
    /// Cosmos state is modified in the same transaction as state information inside
    /// the EVM.
    /// @param _delegator Address of the delegator
    /// @param _validatorAddr Address of the validator to delegate to
    /// @param _amount Amount to delegate
    /// @param _before Boolean to specify if funds should be transferred to delegator before the precompile call
    /// @param _after Boolean to specify if funds should be transferred to delegator after the precompile call
    function testDelegateWithCounterAndTransfer(
        address payable _delegator,
        string memory _validatorAddr,
        uint256 _amount,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            counter++;
            (bool sent, ) = _delegator.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
        staking.STAKING_CONTRACT.delegate(_delegator, _validatorAddr, _amount);
        if (_after) {
            counter++;
            (bool sent, ) = _delegator.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }

    /// @dev This function showcased, that there was a bug in the EVM implementation, that occurred when
    /// Cosmos state is modified in the same transaction as state information inside
    /// the EVM.
    /// @param _dest Address to send some funds from the contract
    /// @param _delegator Address of the delegator
    /// @param _validatorAddr Address of the validator to delegate to
    /// @param _amount Amount to delegate
    /// @param _before Boolean to specify if funds should be transferred to delegator before the precompile call
    /// @param _after Boolean to specify if funds should be transferred to delegator after the precompile call
    function testDelegateWithTransfer(
        address payable _dest,
        address payable _delegator,
        string memory _validatorAddr,
        uint256 _amount,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            counter++;
            (bool sent, ) = _dest.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
        staking.STAKING_CONTRACT.delegate(_delegator, _validatorAddr, _amount);
        if (_after) {
            counter++;
            (bool sent, ) = _dest.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }    

    /// @dev This function calls the staking precompile's create validator method
    /// and transfers of funds to the validator address (if specified).
    /// @param _descr The initial description
    /// @param _commRates The initial commissionRates
    /// @param _minSelfDel The validator's self declared minimum self delegation
    /// @param _validator The validator's operator address
    /// @param _pubkey The consensus public key of the validator
    /// @param _value The amount of the coin to be self delegated to the validator
    /// @param _before Boolean to specify if funds should be transferred to delegator before the precompile call
    /// @param _after Boolean to specify if funds should be transferred to delegator after the precompile call
    function testCreateValidatorWithTransfer(
        staking.Description calldata _descr,
        staking.CommissionRates calldata _commRates,
        uint256 _minSelfDel,
        address _validator,
        string memory _pubkey,
        uint256 _value,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            counter++;
            (bool sent, ) = _validator.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
        bool success = staking.STAKING_CONTRACT.createValidator(
            _descr,
            _commRates,
            _minSelfDel,
            _validator,
            _pubkey,
            _value
        );
        require(success, "Failed to create the validator");
        if (_after) {
            counter++;
            (bool sent, ) = _validator.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }
}
