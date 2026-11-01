// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract Counter {
    uint256 public count;
    address public immutable owner;
    event CountChanged(address indexed operator, uint256 oldValue, uint256 newValue);

    constructor() {
        owner = msg.sender;
    }

    function setNumber(uint256 newValue) public returns (uint256) {
        uint256 oldValue = count;
        count = newValue;
        emit CountChanged(msg.sender, oldValue, newValue);
        return newValue;
    }

    function increment() public returns (uint256) {
        return setNumber(count + 1);
    }
}