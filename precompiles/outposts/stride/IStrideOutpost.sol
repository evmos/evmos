// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../../common/Types.sol";

/// @dev The Stride Outpost contract's address.
address constant STRIDE_OUTPOST_ADDRESS = 0x0000000000000000000000000000000000000900;

/// @dev The Stride Outpost contract's instance.
IStrideOutpost constant STRIDE_OUTPOST_CONTRACT = IStrideOutpost(
    STRIDE_OUTPOST_ADDRESS
);

/// @dev AutopilotParams is a struct containing the parameters for a liquid stake and redeem transactions.
/// @param receiver - The address on the Evmos chain that will redeem or receive LSD.
/// @param strideForwarder - The bech32-formatted address on the Stride chain that will be used to execute
/// LiquidStake or Redeem transactions.
struct AutopilotParams {
    string channelID;
    address sender;
    address receiver;
    address token;
    uint256 amount;
    string strideForwarder;
}


/// @author Evmos Team
/// @title StrideOutpost Precompiled Contract
/// @dev The interface through which solidity contracts will interact with Stride Outpost that uses ICS20 under the hood
/// @custom:address 0x0000000000000000000000000000000000000900
interface IStrideOutpost {
    /// @dev Emitted when an ICS-20 transfer is executed.
    /// @param sender The address of the sender.
    /// @param receiver The address of the receiver.
    /// @param sourcePort The source port of the IBC transaction.
    /// @param sourceChannel The source channel of the IBC transaction.
    /// @param denom The denomination of the tokens transferred.
    /// @param amount The amount of tokens transferred.
    /// @param memo The IBC transaction memo.
    event IBCTransfer(
        address indexed sender,
        string indexed receiver,
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
    /// @param receiver The address of the receiver on the Evmos chain.
    /// @param token The token to be un-luquid staked.
    /// @param strideForwarder The bech32-formatted address of the receiver on the Stride chain.
    /// @param amount The amount of tokens to unstake.
    event RedeemStake(
        address indexed sender,
        address indexed receiver,
        address indexed token,
        string strideForwarder,
        uint256 amount
    );

    /// @dev Liquid stake a native Coin on the Stride chain and return it to the Evmos chain.
    /// @param payload The AutopilotParams struct containing the parameters for the liquid stake transaction.
    function liquidStake(AutopilotParams calldata payload) external returns (bool success);

    /// @dev This method unstakes the LSD Coin (ex. stEvmos, stAtom) and redeems
    /// the native Coin by sending an ICS20 Transfer to the specified chain.
    /// @param payload The AutopilotParams struct containing the parameters for the redeem stake transaction.
    function redeemStake(AutopilotParams calldata payload) external returns (bool success);
}