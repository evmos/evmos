// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../common/Types.sol";

/// @dev The IGov contract's address.
address constant GOV_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000805;

/// @dev Define all the available gov methods.
string constant MSG_VOTE = "/cosmos.gov.v1.MsgVote";

/// @dev The IGov contract's instance.
IGov constant GOV_CONTRACT = IGov(
    GOV_PRECOMPILE_ADDRESS
);

/**
 * @dev VoteOption enumerates the valid vote options for a given governance proposal.
 */
enum VoteOption {
    // Unspecified defines a no-op vote option.
    Unspecified,
    // Yes defines a yes vote option.
    Yes,
    // Abstain defines an abstain vote option.
    Abstain,
    // No defines a no vote option.
    No,
    // NoWithWeto defines a no with veto vote option.
    NoWithWeto
}
/// @dev Vote represents a vote on a governance proposal
struct SingleVote {
    uint64 proposalId;
    address voter;
    WeightedVoteOption[] options;
    string metadata;
}

/// @dev WeightedVoteOption represents a weighted vote option
struct WeightedVoteOption {
    VoteOption option;
    string weight;
}

/// @author Luke
/// @title Gov Precompile Contract
/// @dev The interface through which solidity contracts will interact with Gov
/// @custom:address 0x0000000000000000000000000000000000000805
interface IGov {
    /// @dev Vote defines an Event emitted when a proposal voted.
    /// @param voter the address of the voter
    /// @param proposalId the proposal of id
    /// @param option the option for voter
    event Vote(
        address indexed voter,
        uint64 proposalId,
        uint8 option
    );

    /// TRANSACTIONS

    /// @dev vote defines a method to add a vote on a specific proposal.
    /// @param voter The address of the voter
    /// @param proposalId the proposal of id
    /// @param option the option for voter
    /// @param metadata the metadata for voter send
    /// @return success Whether the transaction was successful or not
    function vote(
        address voter,
        uint64 proposalId,
        VoteOption option,
        string memory metadata
    ) external returns (bool success);

    /// QUERIES

    /// @dev votes returns the votes for a specific proposal.
    /// @param proposalId the proposal id
    /// @param pagination the pagination options
    /// @return votes The votes for the proposal
    /// @return pageResponse The pagination information
    function votes(
        uint64 proposalId,
        PageRequest calldata pagination
    ) external view returns (SingleVote[] memory votes, PageResponse memory pageResponse);
}
