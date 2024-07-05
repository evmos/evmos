pragma solidity ^0.8.17;

interface ICounter {
    function add() external;
    function subtract() external;
    function getCounter() external view returns (uint256);
    event Changed(uint256 counter);
    event Added(uint256 counter);
}
