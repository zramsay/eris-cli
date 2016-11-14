contract GSContract {
  uint storedData;

  function GSContract(uint initialValue) {
    storedData = initialValue;
  }

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}
