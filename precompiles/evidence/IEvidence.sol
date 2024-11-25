// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "../common/Types.sol";

/// @dev The IEvidence contract's address.
address constant EVIDENCE_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000807;

/// @dev The IEvidence contract's instance.
IEvidence constant EVIDENCE_CONTRACT = IEvidence(EVIDENCE_PRECOMPILE_ADDRESS);

/// @dev The Equivocation struct contains information about a validator's equivocation
struct Equivocation {
    // height is the equivocation height
    int64 height;
    // time is the equivocation time
    uint64 time;
    // power is the validator's power at the time of the equivocation
    int64 power;
    // consensusAddress is the validator's consensus address
    string consensusAddress;
}

/// @author The Evmos Core Team
/// @title Evidence Precompile Contract
/// @dev The interface through which solidity contracts will interact with the x/evidence module
interface IEvidence {
    /// @dev Event emitted when evidence is submitted
    /// @param submitter The address of the submitter
    /// @param hash The hash of the submitted evidence
    event SubmitEvidence(address indexed submitter, bytes hash);

    /// @dev Submit evidence of misbehavior (equivocation)
    /// @param evidence The evidence of misbehavior
    /// @return success True if the evidence was submitted successfully
    function submitEvidence(Equivocation calldata evidence) external returns (bool success);

    /// @dev Query evidence by hash
    /// @param evidenceHash The hash of the evidence to query
    /// @return evidence The equivocation evidence data
    function evidence(bytes memory evidenceHash) external view returns (Equivocation memory evidence);

    /// @dev Query all evidence with pagination
    /// @param pageRequest Pagination request
    /// @return evidence List of equivocation evidence
    /// @return pageResponse Pagination response
    function getAllEvidence(PageRequest calldata pageRequest)
        external
        view
        returns (Equivocation[] memory evidence, PageResponse memory pageResponse);
}
