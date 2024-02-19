// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

contract Counter {
  uint256 counter = 0;
  string internal constant ERROR_TOO_LOW = "COUNTER_TOO_LOW";
  event Changed(uint256 counter);
  event Added(uint256 counter);

  function add() public {
    counter++;
    emit Added(counter);
    emit Changed(counter);
  }

  function subtract() public {
    require(counter > 0, ERROR_TOO_LOW);
    counter--;
    emit Changed(counter);
  }

  function getCounter() public view returns (uint256) {
    return counter;
  }
}
