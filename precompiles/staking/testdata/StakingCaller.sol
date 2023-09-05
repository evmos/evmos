// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../StakingI.sol" as staking;

/// @title StakingCaller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the staking precompile.
contract StakingCaller {
    /// counter is used to test the state persistence bug, when EVM and Cosmos state were both changed in the same function.
    uint256 public counter;
    string[] private delegateMethod = [staking.MSG_DELEGATE];

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

    /// @dev This function calls the staking precompile's revoke method.
    /// @param _grantee The address that was approved to spend the funds.
    /// @param _methods The methods to revoke.
    function testRevoke(
        address _grantee,
        string[] calldata _methods
    ) public {
        bool success = staking.STAKING_CONTRACT.revoke(
            _grantee,
            _methods
        );
        require(success, "Failed to revoke approval for staking methods");
    }

    /// @dev This function calls the staking precompile's delegate method.
    /// @param _addr The address to approve.
    /// @param _validatorAddr The validator address to delegate to.
    /// @param _amount The amount to delegate.
    function testDelegate(
        address _addr,
        string memory _validatorAddr,
        uint256 _amount
    ) public {
        staking.STAKING_CONTRACT.delegate(_addr, _validatorAddr, _amount);
    }

    /// @dev This function calls the staking precompile's undelegate method.
    /// @param _addr The address to approve.
    /// @param _validatorAddr The validator address to delegate to.
    /// @param _amount The amount to delegate.
    function testUndelegate(
        address _addr,
        string memory _validatorAddr,
        uint256 _amount
    ) public {
        staking.STAKING_CONTRACT.undelegate(_addr, _validatorAddr, _amount);
    }

    /// @dev This function calls the staking precompile's redelegate method.
    /// @param _addr The address to approve.
    /// @param _validatorSrcAddr The validator address to delegate from.
    /// @param _validatorDstAddr The validator address to delegate to.
    /// @param _amount The amount to delegate.
    function testRedelegate(
        address _addr,
        string memory _validatorSrcAddr,
        string memory _validatorDstAddr,
        uint256 _amount
    ) public {
        staking.STAKING_CONTRACT.redelegate(
            _addr,
            _validatorSrcAddr,
            _validatorDstAddr,
            _amount
        );
    }

    /// @dev This function calls the staking precompile's cancel unbonding delegation method.
    /// @param _addr The address to approve.
    /// @param _validatorAddr The validator address to delegate from.
    /// @param _amount The amount to delegate.
    /// @param _creationHeight The creation height of the unbonding delegation.
    function testCancelUnbonding(
        address _addr,
        string memory _validatorAddr,
        uint256 _amount,
        uint256 _creationHeight
    ) public {
        staking.STAKING_CONTRACT.cancelUnbondingDelegation(
            _addr,
            _validatorAddr,
            _amount,
            _creationHeight
        );
    }

    /// @dev This function calls the staking precompile's allowance query method.
    /// @param _grantee The address that received the grant.
    /// @param method The method to query.
    /// @return allowance The allowance.
    function getAllowance(
        address _grantee,
        string memory method
    ) public view returns (uint256 allowance) {
        return staking.STAKING_CONTRACT.allowance(_grantee, msg.sender, method);
    }

    /// @dev This function calls the staking precompile's validator query method.
    /// @param _validatorAddr The validator address to query.
    /// @return validator The validator.
    function getValidator(
        string memory _validatorAddr
    ) public view returns (staking.Validator memory validator) {
        return staking.STAKING_CONTRACT.validator(_validatorAddr);
    }

    /// @dev This function calls the staking precompile's validators query method.
    /// @param _status The status of the validators to query.
    /// @param _pageRequest The page request to query.
    /// @return validators The validators.
    /// @return pageResponse The page response.
    function getValidators(
        string memory _status,
        staking.PageRequest calldata _pageRequest
    )
        public
        view
        returns (
            staking.Validator[] memory validators,
            staking.PageResponse memory pageResponse
        )
    {
        return staking.STAKING_CONTRACT.validators(_status, _pageRequest);
    }

    /// @dev This function calls the staking precompile's delegation query method.
    /// @param _addr The address to approve.
    /// @param _validatorAddr The validator address to delegate from.
    /// @return shares The shares of the delegation.
    /// @return balance The balance of the delegation.
    function getDelegation(
        address _addr,
        string memory _validatorAddr
    ) public view returns (uint256 shares, staking.Coin memory balance) {
        return staking.STAKING_CONTRACT.delegation(_addr, _validatorAddr);
    }

    /// @dev This function calls the staking precompile's redelegations query method.
    /// @param _addr The address to approve.
    /// @param _validatorSrcAddr The validator address to delegate from.
    /// @param _validatorDstAddr The validator address to delegate to.
    /// @return redelegation The redelegation output.
    function getRedelegation(
        address _addr,
        string memory _validatorSrcAddr,
        string memory _validatorDstAddr
    ) public view returns (staking.RedelegationOutput memory redelegation) {
        return
            staking.STAKING_CONTRACT.redelegation(
                _addr,
                _validatorSrcAddr,
                _validatorDstAddr
            );
    }

    /// @dev This function calls the staking precompile's redelegations query method.
    /// @param _delegatorAddr The delegator address.
    /// @param _validatorSrcAddr The validator address to delegate from.
    /// @param _validatorDstAddr The validator address to delegate to.
    /// @param _pageRequest The page request to query.
    /// @return response The redelegation response.
    function getRedelegations(
        address _delegatorAddr,
        string memory _validatorSrcAddr,
        string memory _validatorDstAddr,
        staking.PageRequest memory _pageRequest
    )
        public
        view
        returns (
            staking.RedelegationResponse[] memory response,
            staking.PageResponse memory pageResponse
        )
    {
        return
            staking.STAKING_CONTRACT.redelegations(
                _delegatorAddr,
                _validatorSrcAddr,
                _validatorDstAddr,
                _pageRequest
            );
    }

    /// @dev This function calls the staking precompile's unbonding delegation query method.
    /// @param _addr The address to approve.
    /// @param _validatorAddr The validator address to delegate from.
    /// @return unbondingDelegation The unbonding delegation output.
    function getUnbondingDelegation(
        address _addr,
        string memory _validatorAddr
    ) public view returns (staking.UnbondingDelegationOutput memory unbondingDelegation) {
        return
            staking.STAKING_CONTRACT.unbondingDelegation(_addr, _validatorAddr);
    }

    /// @dev This function calls the staking precompile's approve method to grant approval for an undelegation.
    /// Next, the undelegate method is called to execute an unbonding.
    /// @param _addr The address to approve.
    /// @param _approveAmount The amount to approve.
    /// @param _undelegateAmount The amount to undelegate.
    /// @param _validatorAddr The validator address to delegate from.
    function testApproveAndThenUndelegate(
        address _addr,
        uint256 _approveAmount,
        uint256 _undelegateAmount,
        string memory _validatorAddr
    ) public {
        string[] memory approvedMethods = new string[](1);
        approvedMethods[0] = staking.MSG_UNDELEGATE;
        bool success = staking.STAKING_CONTRACT.approve(
            _addr,
            _approveAmount,
            approvedMethods
        );
        require(success, "failed to approve undelegation method");
        staking.STAKING_CONTRACT.undelegate(
            tx.origin,
            _validatorAddr,
            _undelegateAmount
        );
    }

    /// @dev This function is used to test the behaviour when executing transactions using special function calling opcodes,
    /// like call, delegatecall, staticcall, and callcode.
    /// @param _addr The address to approve.
    /// @param _validatorAddr The validator address to delegate from.
    /// @param _amount The amount to undelegate.
    /// @param _calltype The opcode to use.
    function testCallUndelegate(
        address _addr,
        string memory _validatorAddr,
        uint256 _amount,
        string memory _calltype
    ) public {
        address calledContractAddress = staking.STAKING_PRECOMPILE_ADDRESS;
        bytes memory payload = abi.encodeWithSignature(
            "undelegate(address,string,uint256)",
            _addr,
            _validatorAddr,
            _amount
        );
        bytes32 calltypeHash = keccak256(abi.encodePacked(_calltype));

        if (calltypeHash == keccak256(abi.encodePacked("delegatecall"))) {
            (bool success, ) = calledContractAddress.delegatecall(payload);
            require(success, "failed delegatecall to precompile");
        } else if (calltypeHash == keccak256(abi.encodePacked("staticcall"))) {
            (bool success, ) = calledContractAddress.staticcall(payload);
            require(success, "failed staticcall to precompile");
        } else if (calltypeHash == keccak256(abi.encodePacked("call"))) {
            (bool success, ) = calledContractAddress.call(payload);
            require(success, "failed call to precompile");
        } else if (calltypeHash == keccak256(abi.encodePacked("callcode"))) {
            // NOTE: callcode is deprecated and now only available via inline assembly
            assembly {
                // Load the function signature and argument data onto the stack
                let ptr := add(payload, 0x20)
                let len := mload(payload)

                // Invoke the contract at calledContractAddress in the context of the current contract
                // using CALLCODE opcode and the loaded function signature and argument data
                let success := callcode(
                    gas(),
                    calledContractAddress,
                    0,
                    ptr,
                    len,
                    0,
                    0
                )

                // Check if the call was successful and revert the transaction if it failed
                if iszero(success) {
                    revert(0, 0)
                }
            }
        } else {
            revert("invalid calltype");
        }
    }

    /// @dev This function is used to test the behaviour when executing queries using special function calling opcodes,
    /// like call, delegatecall, staticcall, and callcode.
    /// @param _addr The address of the delegator.
    /// @param _validatorAddr The validator address to query for.
    /// @param _calltype The opcode to use.
    function testCallDelegation(
        address _addr,
        string memory _validatorAddr,
        string memory _calltype
    ) public returns (uint256 shares, staking.Coin memory coin) {
        address calledContractAddress = staking.STAKING_PRECOMPILE_ADDRESS;
        bytes memory payload = abi.encodeWithSignature(
            "delegation(address,string)",
            _addr,
            _validatorAddr
        );
        bytes32 calltypeHash = keccak256(abi.encodePacked(_calltype));

        if (calltypeHash == keccak256(abi.encodePacked("delegatecall"))) {
            (bool success, bytes memory data) = calledContractAddress
                .delegatecall(payload);
            require(success, "failed delegatecall to precompile");
            (shares, coin) = abi.decode(data, (uint256, staking.Coin));
        } else if (calltypeHash == keccak256(abi.encodePacked("staticcall"))) {
            (bool success, bytes memory data) = calledContractAddress
                .staticcall(payload);
            require(success, "failed staticcall to precompile");
            (shares, coin) = abi.decode(data, (uint256, staking.Coin));
        } else if (calltypeHash == keccak256(abi.encodePacked("call"))) {
            (bool success, bytes memory data) = calledContractAddress.call(
                payload
            );
            require(success, "failed call to precompile");
            (shares, coin) = abi.decode(data, (uint256, staking.Coin));
        } else if (calltypeHash == keccak256(abi.encodePacked("callcode"))) {
            //Function signature
            bytes4 sig = bytes4(keccak256(bytes("delegation(address,string)")));
            // Length of the input data is 164 bytes on 32bytes chunks:
            //                          Memory location
            // 0 - 4 byte signature     x
            // 1 - 0x0000..address		x + 0x04
            // 2 - 0x0000..00			x + 0x24
            // 3 - 0x40..0000			x + 0x44
            // 4 - val_addr_chunk1		x + 0x64
            // 5 - val_addr_chunk2..000	x + 0x84
            uint256 len = 164;
            // Coin type includes denom & amount
            // need to get these separately from the bytes response
            string memory denom;
            uint256 amt;

            // NOTE: callcode is deprecated and now only available via inline assembly
            assembly {
                let chunk1 := mload(add(_validatorAddr, 32)) // first 32 bytes of validator address string
                let chunk2 := mload(add(add(_validatorAddr, 32), 32)) // remaining 19 bytes of val address string

                // Load the function signature and argument data onto the stack
                let x := mload(0x40) // Find empty storage location using "free memory pointer"
                mstore(x, sig) // Place function signature at begining of empty storage
                mstore(add(x, 0x04), _addr) // Place the address (input param) after the function sig
                mstore(add(x, 0x24), 0x40) // These are needed for
                mstore(add(x, 0x44), 0x33) // bytes unpacking
                mstore(add(x, 0x64), chunk1) // Place the validator address in 2 chunks (input param)
                mstore(add(x, 0x84), chunk2) // because mstore stores 32bytes

                // Invoke the contract at calledContractAddress in the context of the current contract
                // using CALLCODE opcode and the loaded function signature and argument data
                let success := callcode(
                    gas(),
                    calledContractAddress, // to addr
                    0, // no value
                    x, // inputs are stored at location x
                    len, // inputs length
                    x, //store output over input (saves space)
                    0xC0 // output length for this call
                )

                // output length for this call is 192 bytes splitted on these 32 bytes chunks:
                // 1 - 0x00..amt   -> @ 0x40
                // 2 - 0x000..00   -> @ 0x60
                // 3 - 0x40..000   -> @ 0x80
                // 4 - 0x00..amt    -> @ 0xC0
                // 5 - 0x00..denom  -> @ 0xE0   TODO: cannot get the return value

                shares := mload(x) // Assign shares output value - 32 bytes long
                amt := mload(add(x, 0x60)) // Assign output value to c - 64 bytes long (string & uint256)

                mstore(0x40, add(x, 0x100)) // Set storage pointer to empty space

                // Check if the call was successful and revert the transaction if it failed
                if iszero(success) {
                    revert(0, 0)
                }
            }
            coin = staking.Coin(denom, amt); // NOTE: this is returning a blank denom because unpacking the denom is not straightforward and hasn't been solved, which is okay for this generic test case
        } else {
            revert("invalid calltype");
        }

        return (shares, coin);
    }

    /// @dev This function showcased, that there was a bug in the EVM implementation, that occured when
    /// Cosmos state is modified in the same transaction as state information inside
    /// the EVM.
    /// @param _validatorAddr Address of the validator to delegate to
    /// @param _amount Amount to delegate
    function testDelegateIncrementCounter(
        string memory _validatorAddr,
        uint256 _amount
    ) public {
        bool successStk = staking.STAKING_CONTRACT.approve(
            address(this),
            _amount,
            delegateMethod
        );
        require(successStk, "Staking Approve failed");
        staking.STAKING_CONTRACT.delegate(
            address(this),
            _validatorAddr,
            _amount
        );
        counter += 1;
    }

    /// @dev This function showcases the possibility to deposit into the contract
    /// and immediately delegate to a validator using the same balance in the same transaction.
    function approveDepositAndDelegate(string memory _validatorAddr) payable public {
        bool successTx = staking.STAKING_CONTRACT.approve(
            address(this),
            msg.value,
            delegateMethod
        );
        require(successTx, "Delegate Approve failed");
        staking.STAKING_CONTRACT.delegate(
            address(this),
            _validatorAddr,
            msg.value
        );
    }

    /// @dev This function is suppose to fail because the amount to delegate is
    /// higher than the amount approved.
    function approveDepositAndDelegateExceedingAllowance(string memory _validatorAddr) payable public {
        bool successTx = staking.STAKING_CONTRACT.approve(
            tx.origin,
            msg.value,
            delegateMethod
        );
        require(successTx, "Delegate Approve failed");
        staking.STAKING_CONTRACT.delegate(
            address(this),
            _validatorAddr,
            msg.value + 1
        );
    }

    /// @dev This function is suppose to fail because the amount to delegate is
    /// higher than the amount approved.
    function approveDepositDelegateAndFailCustomLogic(string memory _validatorAddr) payable public {
        bool successTx = staking.STAKING_CONTRACT.approve(
            tx.origin,
            msg.value,
            delegateMethod
        );
        require(successTx, "Delegate Approve failed");
        staking.STAKING_CONTRACT.delegate(
            address(this),
            _validatorAddr,
            msg.value
        );
        // This should fail since the balance is already spent in the previous call
        payable(msg.sender).transfer(msg.value);
    }
}
