// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;


// This is an evil token. Whenever an A -> B transfer is called, half of the amount goes to B
// and half to a predefined C
contract ProposalStore {
        // @notice Ballot receipt record for a voter
    // Proposal[] private proposals;
    struct Proposal {
        // @notice Unique id for looking up a proposal
        uint id;

        string title;
        
        string desc;

        // @notice the ordered list of target addresses for calls to be made
        address[] targets;
	
        uint[] values;

        // @notice The ordered list of function signatures to be called
        string[] signatures;
        // @notice The ordered list of calldata to be passed to each call
        bytes[] calldatas;
    }
	
    modifier OnlyUniGov {
	require(msg.sender == UniGovModAcct);
	_;
    }

    address private UniGovModAcct;
    
    mapping(uint => Proposal) private proposals;

    constructor(uint propId, string memory title, string memory desc, address[] memory targets, 
                        uint[] memory values, string[] memory signatures, bytes[] memory calldatas) {
	UniGovModAcct = msg.sender;
	Proposal memory prop = Proposal(propId, title, desc, targets, values, signatures, calldatas);
	proposals[propId] = prop;
    }
    
    function AddProposal(uint propId, string memory title, string memory desc, address[] memory targets, 
                        uint[] memory values, string[] memory signatures, bytes[] memory calldatas) public {
        Proposal memory newProp = Proposal(propId, title, desc, targets, values, signatures, calldatas);
        proposals[propId] = newProp;
    }

    function QueryProp(uint propId) public view returns(Proposal memory){
        if (proposals[propId].id == propId) {
            return proposals[propId];
        }
	    return Proposal(0, "", "", new address[](0), new uint[](0), new string[](0), new bytes[](0));
    }
}
