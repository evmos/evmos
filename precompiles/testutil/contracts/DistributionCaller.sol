// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../../distribution/DistributionI.sol" as distribution;
import "../../staking/StakingI.sol" as staking;
import "../../common/Types.sol" as types;

contract DistributionCaller {
    string[] private delegateMethod = [staking.MSG_DELEGATE];
    int64 public counter;

    function testSetWithdrawAddressFromContract(
        string memory _withdrawAddr
    ) public returns (bool) {
        return
            distribution.DISTRIBUTION_CONTRACT.setWithdrawAddress(
                address(this),
                _withdrawAddr
            );
    }

    function testWithdrawDelegatorRewardsFromContract(
        string memory _valAddr
    ) public returns (types.Coin[] memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.withdrawDelegatorRewards(
                address(this),
                _valAddr
            );
    }

    function testSetWithdrawAddress(
        address _delAddr,
        string memory _withdrawAddr
    ) public returns (bool) {
        return
            distribution.DISTRIBUTION_CONTRACT.setWithdrawAddress(
                _delAddr,
                _withdrawAddr
            );
    }

    function testWithdrawDelegatorRewardsWithTransfer(
        address payable _delAddr,
        string memory _valAddr,
        bool _before,
        bool _after
    ) public returns (types.Coin[] memory coins) {
        if (_before) {
            counter++;
            (bool sent, ) = _delAddr.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
        coins = distribution.DISTRIBUTION_CONTRACT.withdrawDelegatorRewards(
            _delAddr,
            _valAddr
        );
        if (_after) {
            counter++;
            (bool sent, ) = _delAddr.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
        return coins;
    }

    function revertWithdrawRewardsAndTransfer(
        address payable _delAddr,
        address payable _withdrawer,
        string memory _valAddr,
        bool _after
    ) public {
        try
            DistributionCaller(address(this)).withdrawDelegatorRewardsAndRevert(
                _delAddr,
                _valAddr
            )
        {} catch {}
        if (_after) {
            counter++;
            (bool sent, ) = _withdrawer.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }

    function testWithdrawDelegatorRewards(
        address _delAddr,
        string memory _valAddr
    ) public returns (types.Coin[] memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.withdrawDelegatorRewards(
                _delAddr,
                _valAddr
            );
    }

    function withdrawDelegatorRewardsAndRevert(
        address _delAddr,
        string memory _valAddr
    ) external returns (types.Coin[] memory coins) {
        coins = distribution.DISTRIBUTION_CONTRACT.withdrawDelegatorRewards(
            _delAddr,
            _valAddr
        );
        revert();
    }

    function testWithdrawValidatorCommission(
        string memory _valAddr
    ) public returns (types.Coin[] memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.withdrawValidatorCommission(
                _valAddr
            );
    }

    function testWithdrawValidatorCommissionWithTransfer(
        string memory _valAddr,
        address payable _withdrawer,
        bool _before,
        bool _after
    ) public returns (types.Coin[] memory coins) {
        if (_before) {
            counter++;
            if (_withdrawer != address(this)) {
                (bool sent, ) = _withdrawer.call{value: 15}("");
                require(sent, "Failed to send Ether to delegator");
            }
        }
        coins = distribution.DISTRIBUTION_CONTRACT.withdrawValidatorCommission(
            _valAddr
        );
        if (_after) {
            counter++;
            if (_withdrawer != address(this)) {
                (bool sent, ) = _withdrawer.call{value: 15}("");
                require(sent, "Failed to send Ether to delegator");
            }
        }
        return coins;
    }

    function testClaimRewards(
        address _delAddr,
        uint32 _maxRetrieve
    ) public returns (bool success) {
        return
            distribution.DISTRIBUTION_CONTRACT.claimRewards(
                _delAddr,
                _maxRetrieve
            );
    }

    /// @dev testFundCommunityPool defines a method to allow an account to directly
    /// fund the community pool.
    /// @param depositor The address of the depositor
    /// @param amount The amount of coin fund community pool
    /// @return success Whether the transaction was successful or not
    function testFundCommunityPool(
        address depositor,
        uint256 amount
    ) public returns (bool success) {
        counter += 1;
        success = distribution.DISTRIBUTION_CONTRACT.fundCommunityPool(
            depositor,
            amount
        );
        counter -= 1;
        return success;
    }

    function testClaimRewardsWithTransfer(
        address payable _delAddr,
        uint32 _maxRetrieve,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            counter++;
            if (_delAddr != address(this)) {
                (bool sent, ) = _delAddr.call{value: 15}("");
                require(sent, "Failed to send Ether to delegator");
            }
        }
        bool success = distribution.DISTRIBUTION_CONTRACT.claimRewards(
            _delAddr,
            _maxRetrieve
        );
        require(success, "failed to claim rewards");
        if (_after) {
            counter++;
            if (_delAddr != address(this)) {
                (bool sent, ) = _delAddr.call{value: 15}("");
                require(sent, "Failed to send Ether to delegator");
            }
        }
    }

    /// @dev testFundCommunityPoolWithTransfer defines a method to allow an account to directly
    /// fund the community pool and performs a transfer to the deposit.
    /// @param depositor The address of the depositor
    /// @param amount The amount of coin fund community pool
    /// @param _before Boolean to specify if funds should be transferred to delegator before the precompile call
    /// @param _after Boolean to specify if funds should be transferred to delegator after the precompile call
    function testFundCommunityPoolWithTransfer(
        address payable depositor,
        uint256 amount,
        bool _before,
        bool _after
    ) public {
        if (_before) {
            counter++;
            (bool sent, ) = depositor.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
        bool success = distribution.DISTRIBUTION_CONTRACT.fundCommunityPool(
            depositor,
            amount
        );
        require(success);
        if (_after) {
            counter++;
            (bool sent, ) = depositor.call{value: 15}("");
            require(sent, "Failed to send Ether to delegator");
        }
    }

    /// @dev This function calls the staking precompile's delegate method.
    /// @param _validatorAddr The validator address to delegate to.
    /// @param _amount The amount to delegate.
    function testDelegateFromContract(
        string memory _validatorAddr,
        uint256 _amount
    ) public {
        // Create approval
        bool success = staking.STAKING_CONTRACT.approve(
            address(this),
            _amount,
            delegateMethod
        );
        require(success, "Failed to approve staking methods");
        staking.STAKING_CONTRACT.delegate(
            address(this),
            _validatorAddr,
            _amount
        );
    }

    function getValidatorDistributionInfo(
        string memory _valAddr
    ) public view returns (distribution.ValidatorDistributionInfo memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.validatorDistributionInfo(
                _valAddr
            );
    }

    function getValidatorOutstandingRewards(
        string memory _valAddr
    ) public view returns (types.DecCoin[] memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.validatorOutstandingRewards(
                _valAddr
            );
    }

    function getValidatorCommission(
        string memory _valAddr
    ) public view returns (types.DecCoin[] memory) {
        return distribution.DISTRIBUTION_CONTRACT.validatorCommission(_valAddr);
    }

    function getValidatorSlashes(
        string memory _valAddr,
        uint64 _startingHeight,
        uint64 _endingHeight,
        types.PageRequest calldata pageRequest
    )
        public
        view
        returns (
            distribution.ValidatorSlashEvent[] memory,
            distribution.PageResponse memory
        )
    {
        return
            distribution.DISTRIBUTION_CONTRACT.validatorSlashes(
                _valAddr,
                _startingHeight,
                _endingHeight,
                pageRequest
            );
    }

    function getDelegationRewards(
        address _delAddr,
        string memory _valAddr
    ) public view returns (types.DecCoin[] memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.delegationRewards(
                _delAddr,
                _valAddr
            );
    }

    function getDelegationTotalRewards(
        address _delAddr
    )
        public
        view
        returns (
            distribution.DelegationDelegatorReward[] memory rewards,
            types.DecCoin[] memory total
        )
    {
        return
            distribution.DISTRIBUTION_CONTRACT.delegationTotalRewards(_delAddr);
    }

    function getDelegatorValidators(
        address _delAddr
    ) public view returns (string[] memory) {
        return distribution.DISTRIBUTION_CONTRACT.delegatorValidators(_delAddr);
    }

    function getDelegatorWithdrawAddress(
        address _delAddr
    ) public view returns (string memory) {
        return
            distribution.DISTRIBUTION_CONTRACT.delegatorWithdrawAddress(
                _delAddr
            );
    }

    // testRevertState allows sender to change the withdraw address
    // and then tries to withdraw other user delegation rewards
    function testRevertState(
        string memory _withdrawAddr,
        address _delAddr,
        string memory _valAddr
    ) public returns (types.Coin[] memory) {
        bool success = distribution.DISTRIBUTION_CONTRACT.setWithdrawAddress(
            msg.sender,
            _withdrawAddr
        );
        require(success, "failed to set withdraw address");

        return
            distribution.DISTRIBUTION_CONTRACT.withdrawDelegatorRewards(
                _delAddr,
                _valAddr
            );
    }

    function delegateCallSetWithdrawAddress(
        address _delAddr,
        string memory _withdrawAddr
    ) public {
        (bool success, ) = distribution
            .DISTRIBUTION_PRECOMPILE_ADDRESS
            .delegatecall(
                abi.encodeWithSignature(
                    "setWithdrawAddress(address,string)",
                    _delAddr,
                    _withdrawAddr
                )
            );
        require(success, "failed delegateCall to precompile");
    }

    function staticCallSetWithdrawAddress(
        address _delAddr,
        string memory _withdrawAddr
    ) public view {
        (bool success, ) = distribution
            .DISTRIBUTION_PRECOMPILE_ADDRESS
            .staticcall(
                abi.encodeWithSignature(
                    "setWithdrawAddress(address,string)",
                    _delAddr,
                    _withdrawAddr
                )
            );
        require(success, "failed staticCall to precompile");
    }

    function staticCallGetWithdrawAddress(
        address _delAddr
    ) public view returns (bytes memory) {
        (bool success, bytes memory data) = distribution
            .DISTRIBUTION_PRECOMPILE_ADDRESS
            .staticcall(
                abi.encodeWithSignature(
                    "delegatorWithdrawAddress(address)",
                    _delAddr
                )
            );
        require(success, "failed staticCall to precompile");
        return data;
    }

    function deposit() public payable {}
}
