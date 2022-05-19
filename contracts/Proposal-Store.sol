// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

// This is an evil token. Whenever an A -> B transfer is called, half of the amount goes to B
// and half to a predefined C
contract ProposalStore {
    address private UniGovModAcct;
        /// @notice Ballot receipt record for a voter
    struct Receipt {
        /// @notice Whether or not a vote has been cast
        bool hasVoted;

        /// @notice Whether or not the voter supports the proposal or abstains
        uint8 support;

        /// @notice The number of votes the voter had, which were cast
        uint96 votes;
    }
    
    mapping(uint => Proposal) private proposals;
    struct Proposal {
        /// @notice Unique id for looking up a proposal
        uint id;

        /// @notice Creator of the proposal
        address proposer;

        /// @notice The timestamp that the proposal will be available for execution, set once the vote succeeds
        uint eta;

        /// @notice the ordered list of target addresses for calls to be made
        address[] targets;

        /// @notice The ordered list of values (i.e. msg.value) to be passed to the calls to be made
        uint[] values;

        /// @notice The ordered list of function signatures to be called
        string[] signatures;

        /// @notice The ordered list of calldata to be passed to each call
        bytes[] calldatas;

        /// @notice The block at which voting begins: holders must delegate their votes prior to this block
        uint startBlock;

        /// @notice The block at which voting ends: votes must be cast prior to this block
        uint endBlock;

        /// @notice Current number of votes in favor of this proposal
        uint forVotes;

        /// @notice Current number of votes in opposition to this proposal
        uint againstVotes;

        /// @notice Current number of votes for abstaining for this proposal
        uint abstainVotes;

        /// @notice Flag marking whether the proposal has been canceled
        bool canceled;

        /// @notice Flag marking whether the proposal has been executed
        bool executed;

        /// @notice Receipts of ballots for the entire set of voters
        mapping (address => Receipt) receipts;
    }
	modifier OnlyUniGov {
	require (msg.sender == UniGovModAcct);
	_;
    }

    constructor() {
	UniGovModAcct == msg.sender;
    }
    
    function AddProposal(Proposal prop) OnlyUniGov external {
	proposals[prop.id] = prop;
    }

    function QueryProp(uint propId) external returns(Proposal){
	if (proposals[propId] == propId) {
	    return proposals[propId];
	}
	return 0;
    }
}
