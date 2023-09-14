// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../../common/Types.sol";

/// @dev The Stride Outpost contract's address.
address constant STRIDE_OUTPOST_ADDRESS = 0x0000000000000000000000000000000000000900;

/// @dev The Stride Outpost contract's instance.
StrideOutpostI constant STRIDE_OUTPOST_CONTRACT = StrideOutpostI(STRIDE_OUTPOST_ADDRESS);

/// @author Evmos Team
/// @title StrideOutpost Precompiled Contract
/// @dev The interface through which solidity contracts will interact with Stride Outpost that uses ICS20 under the hood
/// @custom:address 0x0000000000000000000000000000000000000900
interface StrideOutpostI {

    /// TODO
    function claimAirdrop(string calldata receiver) external returns (bool);

    function unstakeLiquidEvmos(uint256 amount, string calldata receiver) external returns (bool);

    /// @dev Liquid stake evmos on the Stride chain and return to the Evmos chain
    /// @param amount the coin that will be liquid staked (only supports Evmos)
    /// @param receiver the bech32 address of the receiver
    /// @return true if the ICS20 transfer was successful
    function liquidStakeEvmos(
        Coin memory amount,
        string calldata receiver
    ) external returns (bool);


    /// @dev Emitted when a user calls Autopilot to liquid stake Evmos
    /// @param sender the hex address of the sender
    /// @param coin the coin that will be liquid staked (only supports Evmos)
    event LiquidStakeEvmos(
        address indexed sender,
        Coin coin
    );
}