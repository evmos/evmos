// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.0;

import "../common/Types.sol";

/// @dev The IAuctions contract's address.
address constant AUCTIONS_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000805;

/// @dev The ITokenFactory contract's instance.
IAuctions constant AUCTIONS_CONTRACT = IAuctions(AUCTIONS_PRECOMPILE_ADDRESS);

// AuctionInfo is a struct representing the information about the currently ongoing auction.
struct AuctionInfo {
    Coin[] tokens;
    Coin highestBid;
    uint64 currentRound;
    address bidderAddress;
}

/// @author Evmos Team
/// @title Auctions Precompiled Contract
/// @dev The interface through which solidity contracts interacts with the burn auctions module
/// @custom:address 0x0000000000000000000000000000000000000805
interface IAuctions {
    // @dev Event emitted when a new bid is made for the basket of Coins.
    // @param sender - the hex address of the sender of the bid.
    // @param amount - the amount in Evmos tokens for the bid.
    event Bid(address indexed sender, uint256 amount);

    // @dev Event emitted when a new Coin deposit is made for the following burn auction.
    // @param sender - the hex address of the sender of the Coins.
    // @param denom - the denom of the Coin being sent.
    // @param amount - the amount of the Coin being sent.
    event CoinDeposit(address indexed sender, string denom, uint256 amount);

    // @dev Event emitted when a burn auction ends.
    // @param winner - the hex address of the auction winner.
    // @param coins - the Coins won in the auction.
    // @param burned - the amount of tokens burned on the auction.
    event AuctionEnd(address indexed winner, Coin[] coins, uint256 burned);

    // @dev Creates a bid for the basket of assets in the current auction.
    // @dev param sender - the hex address of the sender.
    // @dev amount - the amount of Evmos used to bid.
    function bid(
        address sender,
        uint256 amount
    ) external returns (bool success);

    // @dev Deposits Coins for the following auction.
    // @param sender - The sender of the deposited Coin.
    // @param denom - The denom of the deposited Coin.
    // @param amount - The amount to deposit.
    function depositCoin(
        address sender,
        string memory denom,
        uint256 amount
    ) external returns (bool success);

    // @dev Gets information about the current auction.
    function auctionInfo() external view returns (AuctionInfo memory info);
}
