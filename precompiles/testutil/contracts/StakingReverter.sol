// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../../staking/StakingI.sol";

contract StakingReverter {
    uint counter = 0;

    constructor() payable {}

    function run(uint numTimes, string calldata validatorAddress) external {
        counter++;

        for (uint i = 0; i < numTimes; i++) {
            try
                StakingReverter(address(this)).performDelegation(
                    validatorAddress
                )
            {} catch {}
        }
    }

    function multipleDelegations(
        uint numTimes,
        string calldata validatorAddress
    ) external {
        counter++;

        for (uint i = 0; i < numTimes; i++) {
            StakingReverter(address(this)).performDelegation(validatorAddress);
        }
    }

    function performDelegation(string calldata validatorAddress) external {
        STAKING_CONTRACT.delegate(address(this), validatorAddress, 10);
        revert();
    }

    function getCurrentStake(
        string calldata validatorAddress
    ) external view returns (uint256 shares, Coin memory balance) {
        return STAKING_CONTRACT.delegation(address(this), validatorAddress);
    }

    function multipleQueries(
        uint numTimes,
        address validatorAddress
    ) external view returns (Validator memory validator) {
        for (uint i = 0; i < numTimes; i++) {
            validator = STAKING_CONTRACT.validator(validatorAddress);
        }
        return validator;
    }
}
