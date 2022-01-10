// SPDX-License-Identifier: MIT
// OpenZeppelin Contracts v4.3.2 (token/ERC20/presets/ERC20PresetMinterPauser.sol)

pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import "@openzeppelin/contracts/access/AccessControlEnumerable.sol";
import "@openzeppelin/contracts/utils/Context.sol";

/**
 * @dev {ERC20} token, including:
 *
 *  - ability for holders to burn (destroy) their tokens
 *  - a minter role that allows for token minting (creation)
 *  - a pauser role that allows to stop all token transfers
 *
 * This contract uses {AccessControl} to lock permissioned functions using the
 * different roles - head to its documentation for details.
 *
 * The account that deploys the contract will be granted the minter and pauser
 * roles, as well as the default admin role, which will let it grant both minter
 * and pauser roles to other accounts.
 */
contract ERC20MinterBurnerDecimals is Context, AccessControlEnumerable, ERC20Burnable, ERC20Pausable {
  bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
  bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
  bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");
  uint8 private _decimals;

  /**
    * @dev Grants `DEFAULT_ADMIN_ROLE`, `MINTER_ROLE` and `PAUSER_ROLE` to the
    * account that deploys the contract and customizes tokens decimals
    *
    * See {ERC20-constructor}.
    */
  constructor(string memory name, string memory symbol, uint8 decimals_)
    ERC20(name, symbol) {
      _setupRole(DEFAULT_ADMIN_ROLE, _msgSender());

      _setupRole(MINTER_ROLE, _msgSender());
      _setupRole(PAUSER_ROLE, _msgSender());
      _setupRole(BURNER_ROLE, _msgSender());
      _setupDecimals(decimals_);
  }

  /**
    * @dev Sets `_decimals` as `decimals_ once at Deployment'
    */
  function _setupDecimals(uint8 decimals_) private {
    _decimals = decimals_;
  }

  /**
    * @dev Overrides the `decimals()` method with custom `_decimals`
    */
  function decimals() public view virtual override returns (uint8) {
    return _decimals;
  }

  /**
    * @dev Creates `amount` new tokens for `to`.
    *
    * See {ERC20-_mint}.
    *
    * Requirements:
    *
    * - the caller must have the `MINTER_ROLE`.
    */
  function mint(address to, uint256 amount) public virtual {
      require(hasRole(MINTER_ROLE, _msgSender()), "ERC20MinterBurnerDecimals: must have minter role to mint");
      _mint(to, amount);
  }

      /**
    * @dev Destroys `amount` new tokens for `to`.
    *
    * See {ERC20-_burn}.
    *
    * Requirements:
    *
    * - the caller must have the `BURNER_ROLE`.
    */
  function burnCoins(address from, uint256 amount) public virtual {
      require(hasRole(BURNER_ROLE, _msgSender()), "ERC20MinterBurnerDecimals: must have burner role to burn");
      _burn(from, amount);
  }

  /**
    * @dev Pauses all token transfers.
    *
    * See {ERC20Pausable} and {Pausable-_pause}.
    *
    * Requirements:
    *
    * - the caller must have the `PAUSER_ROLE`.
    */
  function pause() public virtual {
      require(hasRole(PAUSER_ROLE, _msgSender()), "ERC20MinterBurnerDecimals: must have pauser role to pause");
      _pause();
  }

  /**
    * @dev Unpauses all token transfers.
    *
    * See {ERC20Pausable} and {Pausable-_unpause}.
    *
    * Requirements:
    *
    * - the caller must have the `PAUSER_ROLE`.
    */
  function unpause() public virtual {
      require(hasRole(PAUSER_ROLE, _msgSender()), "ERC20MinterBurnerDecimals: must have pauser role to unpause");
      _unpause();
  }

  function _beforeTokenTransfer(
      address from,
      address to,
      uint256 amount
  ) internal virtual override(ERC20, ERC20Pausable) {
      super._beforeTokenTransfer(from, to, amount);
  }
}