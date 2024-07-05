// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.17;

import "../../staking/StakingI.sol" as staking;
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract FlashLoan {
    address public owner;

    constructor() {
        owner = msg.sender;
    }

    function flashLoan(
        address _token,
        string memory _validator,
        uint256 _amount
    ) public returns (bool) {
        require(msg.sender == owner, "Only owner can call this function");

        // Get some tokens to initiate the flash loan
        IERC20 token = IERC20(_token);
        require(
            token.allowance(msg.sender, address(this)) >= _amount,
            "Insufficient allowance"
        );

        uint256 balancePre = token.balanceOf(address(this));
        bool success = token.transferFrom(msg.sender, address(this), _amount);
        require(success, "Failed to transfer tokens for flash loan");
        require(
            token.balanceOf(address(this)) == balancePre + _amount,
            "Flash loan failed"
        );

        // Execute some precompile logic (e.g. staking)
        success = staking.STAKING_CONTRACT.delegate(
            msg.sender,
            _validator,
            _amount
        );
        require(success, "failed to delegate");

        // Transfer tokens back to end the flash loan
        balancePre = token.balanceOf(address(this));
        token.transfer(msg.sender, _amount);
        require(
            token.balanceOf(address(this)) == balancePre - _amount,
            "Flash loan repayment failed"
        );

        return true;
    }

    function flashLoanWithRevert(
        address _token,
        string memory _validator,
        uint256 _amount
    ) public returns (bool) {
        require(msg.sender == owner, "Only owner can call this function");

        // Get some tokens to initiate the flash loan
        IERC20 token = IERC20(_token);
        require(
            token.allowance(msg.sender, address(this)) >= _amount,
            "Insufficient allowance"
        );

        uint256 balancePre = token.balanceOf(address(this));
        bool success = token.transferFrom(msg.sender, address(this), _amount);
        require(success, "Failed to transfer tokens for flash loan");
        require(
            token.balanceOf(address(this)) == balancePre + _amount,
            "Flash loan failed"
        );

        try
            FlashLoan(address(this)).delegateWithRevert(
                msg.sender,
                _validator,
                _amount
            )
        {} catch {}

        return true;
    }

    function delegateWithRevert(
        address _delegator,
        string memory _validator,
        uint256 _amount
    ) external {
        // Execute some precompile logic and revert (e.g. staking)
        bool success = staking.STAKING_CONTRACT.delegate(
            _delegator,
            _validator,
            _amount
        );
        require(success, "failed to delegate");
        revert();
    }
}
