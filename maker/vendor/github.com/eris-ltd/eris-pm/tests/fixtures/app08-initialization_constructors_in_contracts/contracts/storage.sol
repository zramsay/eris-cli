contract SimpleStorage {
  uint storedData;

  function SimpleStorage(uint x) {
    storedData = x;
  }

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}

