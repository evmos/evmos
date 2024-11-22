// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../common/Types.sol";

/// @dev The ISlashing contract's address.
address constant SLASHING_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000000806;

/// @dev The ISlashing contract's instance.
ISlashing constant SLASHING_CONTRACT = ISlashing(SLASHING_PRECOMPILE_ADDRESS);

/// @dev SigningInfo defines a validator's signing info for monitoring their
/// liveness activity.
struct SigningInfo {
    /// @dev Address of the validator
    address validatorAddress;
    /// @dev Height at which validator was first a candidate OR was unjailed
    uint64 startHeight;
    /// @dev Index offset into signed block bit array
    uint64 indexOffset;
    /// @dev Timestamp until which validator is jailed due to liveness downtime
    uint64 jailedUntil;
    /// @dev Whether or not a validator has been tombstoned (killed out of validator set)
    bool tombstoned;
    /// @dev Missed blocks counter (to avoid scanning the array every time)
    uint64 missedBlocksCounter;
}

/// @dev Params defines the parameters for the slashing module.
struct Params {
    /// @dev SignedBlocksWindow defines how many blocks the validator should have signed
    uint64 signedBlocksWindow;
    /// @dev MinSignedPerWindow defines the minimum blocks signed per window to avoid slashing
    string minSignedPerWindow;
    /// @dev DowntimeJailDuration defines how long the validator will be jailed for downtime
    uint64 downtimeJailDuration;
    /// @dev SlashFractionDoubleSign defines the percentage of slash for double sign
    string slashFractionDoubleSign;
    /// @dev SlashFractionDowntime defines the percentage of slash for downtime
    string slashFractionDowntime;
}

/// @author Evmos Team
/// @title Slashing Precompiled Contract
/// @dev The interface through which solidity contracts will interact with slashing.
/// We follow this same interface including four-byte function selectors, in the precompile that
/// wraps the pallet.
/// @custom:address 0x0000000000000000000000000000000000000806
interface ISlashing {
    /// @dev Emitted when a validator is unjailed
    /// @param validator The address of the validator
    event ValidatorUnjailed(address indexed validator);

    /// @dev GetSigningInfo returns the signing info for a specific validator.
    /// @param consAddress The validator consensus address
    /// @return signingInfo The validator signing info
    function getSigningInfo(
        address consAddress
    ) external view returns (SigningInfo memory signingInfo);

    /// @dev GetSigningInfos returns the signing info for all validators.
    /// @param pagination Pagination configuration for the query
    /// @return signingInfos The list of validator signing info
    /// @return pageResponse Pagination information for the response
    function getSigningInfos(
        PageRequest calldata pagination
    ) external view returns (SigningInfo[] memory signingInfos, PageResponse memory pageResponse);

    /// @dev Unjail allows validators to unjail themselves after being jailed for downtime
    /// @param validatorAddress The validator operator address to unjail
    /// @return success true if the unjail operation was successful
    function unjail(address validatorAddress) external returns (bool success);

    /// @dev GetParams returns the slashing module parameters
    /// @return params The slashing module parameters
    function getParams() external view returns (Params memory params);
}
