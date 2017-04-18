import "./owned.sol";

contract mortal is owned {
    function destroy() onlyOwner {
      selfdestruct(owner);
    }
}