// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

import "openzeppelin-solidity/contracts/utils/Context.sol"; // _msgSender();
import "openzeppelin-solidity/contracts/access/AccessControl.sol"; // To setup various roles
import "openzeppelin-solidity/contracts/token/ERC20/IERC20.sol"; // 
import "openzeppelin-solidity/contracts/utils/Address.sol";

contract ICS20Bank {

    using Address for address;

    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

   // Mapping from token ID to account balances
    mapping(string => mapping(address => uint256)) private _balances;

    /*
    * @TODO - Discuss Roles w/ Team
    */
    constructor() {
        _setupRole(ADMIN_ROLE, _msgSender());
    }

    // @dev - Setup a role for an account
    function setOperator(address operator) virtual public {
        require(hasRole(ADMIN_ROLE, _msgSender()), "must have admin role to set new operator");
        _setupRole(OPERATOR_ROLE, operator);
    }

    // Get balance from Mapping
    function balanceOf(address account, string calldata id) virtual external view returns (uint256) {
        require(account != address(0), "ICS20Bank: balance query for the zero address");
        return _balances[id][account];
    }

    // Transfer tokens from one account to another
    function transferFrom(address from, address to, string calldata id, uint256 amount) override virtual external {
        require(to != address(0), "ICS20Bank: transfer to the zero address");
        require(
            from == _msgSender() || hasRole(OPERATOR_ROLE, _msgSender()),
            "ICS20Bank: caller is not owner nor approved"
        );

        uint256 fromBalance = _balances[id][from];
        require(fromBalance >= amount, "ICS20Bank: insufficient balance for transfer");
        _balances[id][from] = fromBalance - amount;
        _balances[id][to] += amount;
    }

    function mint(address account, string calldata id, uint256 amount) override virtual external {
        require(hasRole(OPERATOR_ROLE, _msgSender()), "ICS20Bank: must have minter role to mint");
        _mint(account, id, amount);
        // @Note: Possible recursion - have to resolve
    }

    function burn(address account, string calldata id, uint256 amount) override virtual external {
        require(hasRole(OPERATOR_ROLE, _msgSender()), "ICS20Bank: must have minter role to mint");
        _burn(account, id, amount);
        // @Note: Possible recursion - have to resolve
    }


    /*
    * INTERNAL FUNCTIONS
    */

    // Transfer function that calls the external transfer function - returns a boolean upon success
    function _transferFrom(address sender, address receiver, string memory denom, uint256 amount) override internal returns (bool) {
        try transferFrom(sender, receiver, denom, amount) {
            return true;
        } catch (bytes memory) {
            return false;
        }
    }

    // Mint function that calls the external mint function - returns a boolean upon success
    function _mint(address account, string memory denom, uint256 amount) override internal returns (bool) {
        try mint(account, denom, amount) {
            return true;
        } catch (bytes memory) {
            return false;
        }
    }

    // Burn function that calls the external burn function - returns a boolean upon success
    function _burn(address account, string memory denom, uint256 amount) override internal returns (bool) {
        try burn(account, denom, amount) {
            return true;
        } catch (bytes memory) {
            return false;
        }
    }
}
