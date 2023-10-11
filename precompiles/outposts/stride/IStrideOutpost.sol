// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../../common/Types.sol";

/// @dev The Stride Outpost contract's address.
address constant STRIDE_OUTPOST_ADDRESS = 0x0000000000000000000000000000000000000900;

/// @dev The Stride Outpost contract's instance.
IStrideOutpost constant STRIDE_OUTPOST_CONTRACT = IStrideOutpost(STRIDE_OUTPOST_ADDRESS);



/// @dev Allocation represents a single allocation for an IBC fungible token transfer.
struct Allocation {
    string sourcePort;
    string sourceChannel;
    Coin[] spendLimit;
    string[] allowList;
}

/// @author Evmos Team
/// @title StrideOutpost Precompiled Contract
/// @dev The interface through which solidity contracts will interact with Stride Outpost that uses ICS20 under the hood
/// @custom:address 0x0000000000000000000000000000000000000900
interface IStrideOutpost {
    /// @dev Emitted when an ICS-20 transfer authorization is granted.
    /// @param grantee The address of the grantee.
    /// @param granter The address of the granter.
    /// @param allocations the Allocations authorized with this grant
    event IBCTransferAuthorization(
        address indexed grantee,
        address indexed granter,
        Allocation[] allocations
    );

    /// @dev This event is emitted when an granter revokes a grantee's allowance.
    /// @param grantee The address of the grantee.
    /// @param granter The address of the granter.
    event RevokeIBCTransferAuthorization(
        address indexed grantee,
        address indexed granter
    );

    // @dev Emitted when an ICS-20 transfer is executed.
    /// @param sender The address of the sender.
    /// @param receiver The address of the receiver.
    /// @param sourcePort The source port of the IBC transaction.
    /// @param sourceChannel The source channel of the IBC transaction.
    /// @param denom The denomination of the tokens transferred.
    /// @param amount The amount of tokens transferred.
    /// @param memo The IBC transaction memo.
    event IBCTransfer(
        address indexed sender,
        address indexed receiver,
        string sourcePort,
        string sourceChannel,
        string denom,
        uint256 amount,
        string memo
    );

    /// @dev Emitted on a LiquidStake transaction.
    /// @param sender The address of the sender.
    /// @param token The address of the ERC-20 token pair.
    /// @param amount The amount of tokens that were liquid staked.
    event LiquidStake(
        address indexed sender,
        address indexed token,
        uint256 amount
    );

    /// @dev Emitted on a Redeem transaction.
    /// @param sender The address of the sender.
    /// @param token The token to be un-luquid staked.
    /// @param receiver The bech32-formatted address of the receiver on the Stride chain.
    /// @param amount The amount of tokens to unstake.
    event Redeem(
        address indexed sender,
        address indexed token,
        string receiver,
        uint256 amount
    );

    /// @dev Approves IBC transfer with a specific amount of tokens to use only with the Osmosis channel.
    /// @param grantee The address for which the transfer authorization is granted.
    /// @param spendLimit The amount of tokens that can be transferred.
    /// @param allowList The list of allowed tokens to be transferred.
    /// @return approved The boolean value indicating whether the operation succeeded.
    function approve(
        address grantee,
        Coin[] calldata spendLimit,
        string[] calldata allowList
    ) external returns (bool approved);

    /// @dev Revokes IBC transfer authorization for a specific grantee.
    /// @param grantee The address for which the transfer authorization will be revoked.
    function revoke(address grantee) external returns (bool revoked);

    /// @dev Returns the remaining number of tokens that a grantee smart contract
    /// will be allowed to spend on behalf of granter through
    /// IBC transfers. This is an empty by array.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param granter The address of the account able to transfer the tokens.
    /// @return allocations The remaining amounts allowed to spend for
    /// corresponding source port and channel.
    function allowance(
        address grantee,
        address granter
    ) external view returns (Allocation[] memory allocations);

    /// @dev Increase the allowance of a given grantee by a specific amount of tokens for IBC transfer methods.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param denom the denomination of the Coin to be transferred to the receiver
    /// @param amount The amount of tokens to be spent.
    /// @return approved is true if the operation ran successfully
    function increaseAllowance(
        address grantee,
        string calldata denom,
        uint256 amount
    ) external returns (bool approved);


    /// @dev Decreases the allowance of a given grantee by a specific amount of tokens for for IBC transfer methods.
    /// @param grantee The address of the contract that is allowed to spend the granter's tokens.
    /// @param denom the denomination of the Coin to be transferred to the receiver
    /// @param amount The amount of tokens to be spent.
    /// @return approved is true if the operation ran successfully
    function decreaseAllowance(
        address grantee,
        string calldata denom,
        uint256 amount
    ) external returns (bool approved);

    /// @dev Liquid stake a native Coin on the Stride chain and return it to the Evmos chain.
    /// @param sender The sender of the liquid stake transaction.
    /// @param token The hex ERC20 address of the token pair.
    /// @param amount The amount that will be liquid staked.
    /// @param receiver The bech32 address of the receiver.
    /// @return success True if the ICS20 transfer was successful.
    function liquidStake(
        address sender,
        address token,
        uint256 amount,
        string calldata receiver
    ) external returns (bool success);

    /// @dev This method unstakes the LSD Coin (ex. stEvmos, stAtom) and redeems
    /// the native Coin by sending an ICS20 Transfer to the specified chain.
    /// @param sender The sender of the redeem transaction.
    /// @param token The hex address of the token to be redeemed.
    /// @param amount The amount of tokens unstaked.
    /// @param receiver The bech32-formatted address of the receiver on Stride.
    /// @return success The boolean value indicating whether the operation succeeded.
    function redeem(
        address sender,
        address token,
        uint256 amount,
        string calldata receiver
    ) external returns (bool success);
}