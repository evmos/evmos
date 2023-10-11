// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

/// @dev The Stride Outpost contract's address.
address constant STRIDE_OUTPOST_ADDRESS = 0x0000000000000000000000000000000000000900;

/// @dev The Stride Outpost contract's instance.
IStrideOutpost constant STRIDE_OUTPOST_CONTRACT = IStrideOutpost(STRIDE_OUTPOST_ADDRESS);

/// @author Evmos Team
/// @title StrideOutpost Precompiled Contract
/// @dev The interface through which solidity contracts will interact with Stride Outpost that uses ICS20 under the hood
/// @custom:address 0x0000000000000000000000000000000000000900
interface IStrideOutpost {
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

    /// @dev Liquid stake a native Coin on the Stride chain and return it to the Evmos chain.
    /// @param token The hex ERC20 address of the token pair.
    /// @param amount The amount that will be liquid staked.
    /// @param receiver The bech32 address of the receiver.
    /// @return success True if the ICS20 transfer was successful.
    function liquidStake(
        address token,
        uint256 amount,
        string calldata receiver
    ) external returns (bool success);

    /// @dev This method unstakes the LSD Coin (ex. stEvmos, stAtom) and redeems
    /// the native Coin by sending an ICS20 Transfer to the specified chain.
    /// @param token The hex address of the token to be redeemed.
    /// @param amount The amount of tokens unstaked.
    /// @param receiver The bech32-formatted address of the receiver on Stride.
    /// @return success The boolean value indicating whether the operation succeeded.
    function redeem(
        address token,
        uint256 amount,
        string calldata receiver
    ) external returns (bool success);
}