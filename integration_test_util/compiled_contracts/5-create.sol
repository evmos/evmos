// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

contract Foo {
    event ConstructorCall();

    constructor() {
        emit ConstructorCall();
    }
}

pragma solidity >=0.7.0 <0.9.0;

contract Bar {
    event Deployed();

    function deploy() public returns(Foo) {
        Foo ret = new Foo();

        emit Deployed();
        return ret;
    }

    function deploy2(bytes32 salt) public returns(Foo) {
        Foo ret = new Foo{ salt: salt }();

        emit Deployed();
        return ret;
    }
}

contract BarInteraction {
    address barAddr;

    function setBarAddr(address _bar) public payable {
        barAddr = _bar;
    }

    function deploy() public returns(Foo) {
        return Bar(barAddr).deploy();
    }

    function deploy2(bytes32 salt) public returns(Foo) {
        return Bar(barAddr).deploy2(salt);
    }
}