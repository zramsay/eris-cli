pragma solidity >=0.0.0;

import "./mortal.sol";

contract Dummy is mortal {}

contract DummyFactory {

  Dummy _dummy;

  // create a new contract, the factory is the owner
  function createADummy(){
    _dummy = (new Dummy).value(msg.value)();
  }

  // destroy: should be allowed since the factory is the owner
  function destroyADummy() returns (bool) {
    _dummy.destroy();
    return true;
  }
}
