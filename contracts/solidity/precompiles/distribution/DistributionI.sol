// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../common/Types.sol";

/// @dev The DistributionI contract's address.
address constant DISTRIBUTION_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000801;

/// @dev Define all the available distribution methods.
string constant MSG_SET_WITHDRAWER_ADDRESS = "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress";
string constant MSG_WITHDRAW_DELEGATOR_REWARD = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward";
string constant MSG_WITHDRAW_VALIDATOR_COMMISSION = "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission";

/// @dev The DistributionI contract's instance.
DistributionI constant DISTRIBUTION_CONTRACT = DistributionI(
    DISTRIBUTION_PRECOMPILE_ADDRESS
);

struct ValidatorSlashEvent {
    uint64 validatorPeriod;
    Dec fraction;
}

struct ValidatorDistributionInfo {
    string operatorAddress;
    DecCoin[] selfBondRewards;
    DecCoin[] commission;
}

struct DelegationDelegatorReward {
    string validatorAddress;
    DecCoin[] reward;
}

/// @author Evmos Team
/// @title Distribution Precompile Contract
/// @dev The interface through which solidity contracts will interact with Distribution
/// @custom:address 0x0000000000000000000000000000000000000801
interface DistributionI {
    /// @dev ClaimRewards defines an Event emitted when rewards are claimed
    /// @param delegatorAddress the address of the delegator
    /// @param amount the amount being claimed
    event ClaimRewards(
        address indexed delegatorAddress,
        uint256 amount
    );

    /// @dev SetWithdrawerAddress defines an Event emitted when a new withdrawer address is being set
    /// @param caller the caller of the transaction
    /// @param withdrawerAddress the newly set withdrawer address
    event SetWithdrawerAddress(
        address indexed caller,
        string withdrawerAddress
    );

    /// @dev WithdrawDelegatorRewards defines an Event emitted when rewards from a delegation are withdrawn
    /// @param delegatorAddress the address of the delegator
    /// @param validatorAddress the address of the validator
    /// @param amount the amount being withdrawn from the delegation
    event WithdrawDelegatorRewards(
        address indexed delegatorAddress,
        address indexed validatorAddress,
        uint256 amount
    );

    /// @dev WithdrawValidatorCommission defines an Event emitted when validator commissions are being withdrawn
    /// @param validatorAddress is the address of the validator
    /// @param commission is the total commission earned by the validator
    event WithdrawValidatorCommission(
        string indexed validatorAddress,
        uint256 commission
    );

    /// TRANSACTIONS

    /// @dev Claims all rewards from a select set of validators or all of them for a delegator.
    /// @param delegatorAddress The address of the delegator
    /// @param maxRetrieve The maximum number of validators to claim rewards from
    /// @return success Whether the transaction was successful or not
    function claimRewards(
        address delegatorAddress,
        uint32 maxRetrieve
    ) external returns (bool success);

    /// @dev Change the address, that can withdraw the rewards of a delegator.
    /// Note that this address cannot be a module account.
    /// @param delegatorAddress The address of the delegator
    /// @param withdrawerAddress The address that will be capable of withdrawing rewards for
    /// the given delegator address
    function setWithdrawAddress(
        address delegatorAddress,
        string memory withdrawerAddress
    ) external returns (bool success);

    /// @dev Withdraw the rewards of a delegator from a validator
    /// @param delegatorAddress The address of the delegator
    /// @param validatorAddress The address of the validator
    /// @return amount The amount of Coin withdrawn
    function withdrawDelegatorRewards(
        address delegatorAddress,
        string memory validatorAddress
    ) external returns (Coin[] calldata amount);

    /// @dev Withdraws the rewards commission of a validator.
    /// @param validatorAddress The address of the validator
    /// @return amount The amount of Coin withdrawn
    function withdrawValidatorCommission(
        string memory validatorAddress
    ) external returns (Coin[] calldata amount);

    /// QUERIES
    /// @dev Queries validator commission and self-delegation rewards for validator.
    /// @param validatorAddress The address of the validator
    /// @return distributionInfo The validator's distribution info
    function validatorDistributionInfo(
        string memory validatorAddress
    )
    external
    view
    returns (
        ValidatorDistributionInfo calldata distributionInfo
    );

    /// @dev Queries the outstanding rewards of a validator address.
    /// @param validatorAddress The address of the validator
    /// @return rewards The validator's outstanding rewards
    function validatorOutstandingRewards(
        string memory validatorAddress
    ) external view returns (DecCoin[] calldata rewards);

    /// @dev Queries the accumulated commission for a validator.
    /// @param validatorAddress The address of the validator
    /// @return commission The validator's commission
    function validatorCommission(
        string memory validatorAddress
    ) external view returns (DecCoin[] calldata commission);

    /// @dev Queries the slashing events for a validator in a given height interval
    /// defined by the starting and ending height.
    /// @param validatorAddress The address of the validator
    /// @param startingHeight The starting height
    /// @param endingHeight The ending height
    /// @param pageRequest Defines a pagination for the request.
    /// @return slashes The validator's slash events
    /// @return pageResponse The pagination response for the query
    function validatorSlashes(
        string memory validatorAddress,
        uint64 startingHeight,
        uint64 endingHeight,
        PageRequest calldata pageRequest
    )
    external
    view
    returns (
        ValidatorSlashEvent[] calldata slashes,
        PageResponse calldata pageResponse
    );

    /// @dev Queries the total rewards accrued by a delegation from a specific address to a given validator.
    /// @param delegatorAddress The address of the delegator
    /// @param validatorAddress The address of the validator
    /// @return rewards The total rewards accrued by a delegation.
    function delegationRewards(
        address delegatorAddress,
        string memory validatorAddress
    ) external view returns (DecCoin[] calldata rewards);

    /// @dev Queries the total rewards accrued by each validator, that a given
    /// address has delegated to.
    /// @param delegatorAddress The address of the delegator
    /// @return rewards The total rewards accrued by each validator for a delegator.
    /// @return total The total rewards accrued by a delegator.
    function delegationTotalRewards(
        address delegatorAddress
    )
    external
    view
    returns (
        DelegationDelegatorReward[] calldata rewards,
        DecCoin[] calldata total
    );

    /// @dev Queries all validators, that a given address has delegated to.
    /// @param delegatorAddress The address of the delegator
    /// @return validators The addresses of all validators, that were delegated to by the given address.
    function delegatorValidators(
        address delegatorAddress
    ) external view returns (string[] calldata validators);

    /// @dev Queries the address capable of withdrawing rewards for a given delegator.
    /// @param delegatorAddress The address of the delegator
    /// @return withdrawAddress The address capable of withdrawing rewards for the delegator.
    function delegatorWithdrawAddress(
        address delegatorAddress
    ) external view returns (string memory withdrawAddress);

}
