// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "./IERC20Metadata.sol";

/**
 * @author Evmos Team
 * @title ERC20 Metadata Allowance Interface
 * @dev Interface for the optional metadata and allowance functions from the ERC20 standard.
 */
interface IERC20MetadataAllowance is IERC20Metadata {	
    /** @dev Atomically increases the allowance granted to spender by the caller.
      * This is an alternative to approve that can be used as a mitigation for problems described in
      * IERC20.approve.
      * @param spender The address which will spend the funds.
      * @param addedValue The amount of tokens added to the spender allowance.
      * @return approved Boolean value to indicate if the approval was successful.
    */
    function increaseAllowance(
        address spender,
        uint256 addedValue
    ) external returns (bool approved);

    /** @dev Atomically decreases the allowance granted to spender by the caller.
      * This is an alternative to approve that can be used as a mitigation for problems described in
      * IERC20.approve.
      * @param spender The address which will spend the funds.
      * @param subtractedValue The amount to be substracted from the spender allowance.
      * @return approved Boolean value to indicate if the approval was successful.
    */
    function decreaseAllowance(
        address spender,
        uint256 subtractedValue
    ) external returns (bool approved);
}
