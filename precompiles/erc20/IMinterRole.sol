// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IMinterRole {

    event MinterAdded(address indexed account);
    event MinterRemoved(address indexed account);

    function isMinter(address account) external view returns (bool);

    function addMinter(address account) external;

    function renounceMinter() external;

}