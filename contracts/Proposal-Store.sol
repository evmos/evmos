// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

// This is an evil token. Whenever an A -> B transfer is called, half of the amount goes to B
// and half to a predefined C
contract ProposalStore {
    address private UniGovModAcct;
        /// @notice Ballot receipt record for a voter
    
    mapping(uint => Proposal) private proposals;
    struct Proposal {
        /// @notice Unique id for looking up a proposal
        uint id;

	string title;
	
	string desc;

        // @notice the ordered list of target addresses for calls to be made
	address[] targets;
	
        uint[] values;

        /// @notice The ordered list of function signatures to be called
        string[] signatures;
        /// @notice The ordered list of calldata to be passed to each call
        bytes[] calldatas;
    }
	modifier OnlyUniGov {
	require (msg.sender == UniGovModAcct);
	_;
    }

    constructor() {
	UniGovModAcct == msg.sender;
    }
    
    function AddProposal(uint propId, string memory title, string memory desc, address[] memory targets,
			 uint[] memory values, string[] memory signatures, bytes[] memory calldatas) OnlyUniGov external {
	Proposal memory prop = Proposal(propId, title, desc, targets, values, signatures, calldatas);
	proposals[prop.id] = prop;
    }

    function QueryProp(uint propId) external view returns(Proposal memory){
	if (proposals[propId].id == propId) {
	    return proposals[propId];
	}
	return Proposal(0, "", "", new address[](0), new uint[](0), new string[](0), new bytes[](0));
    } 
}
