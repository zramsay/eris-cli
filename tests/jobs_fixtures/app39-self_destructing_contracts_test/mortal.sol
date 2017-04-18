pragma solidity >=0.0.0;

import "./owned.sol";

contract mortal is owned {
    function destroy() onlyOwner {
      selfdestruct(owner);
    }
}
