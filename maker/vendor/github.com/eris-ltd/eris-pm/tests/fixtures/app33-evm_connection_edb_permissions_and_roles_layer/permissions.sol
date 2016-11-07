contract Permissions {
  function has_base(address addr, int permFlag) constant returns (bool value) {}
  function set_base(address addr, int permFlag, bool value) constant returns (bool val) {}
  function unset_base(address addr, int permFlag) constant returns (int pf) {}
  function set_global(address addr, int permFlag, bool value) constant returns (int pf) {}
  function has_role(address addr, bytes32 role) constant returns (bool val) {}
  function add_role(address addr, bytes32 role) constant returns (bool added) {}
  function rm_role(address addr, bytes32 role) constant returns (bool removed) {}
}

contract permSNative {
  // github.com/eris-ltd/eris-db/manager/eris-mint/evm/snative.go#L17
  Permissions perm = Permissions(address(bytes20("permissions_contract")));

  function has_base(address addr, int permFlag) constant returns (bool value) {
    return perm.has_base(addr, permFlag);
  }

  function set_base(address addr, int permFlag, bool value) constant returns (bool val) {
    return perm.set_base(addr, permFlag, value);
  }

  function unset_base(address addr, int permFlag) constant returns (int pf) {
    return perm.unset_base(addr, permFlag);
  }

  // not currently tested
  function set_global(address addr, int permFlag, bool value) constant returns (int pf) {
    return perm.set_global(addr, permFlag, value);
  }

  function has_role(address addr, bytes32 role) constant returns (bool val) {
    return perm.has_role(addr, role);
  }

  function add_role(address addr, bytes32 role) constant returns (bool added) {
    return perm.add_role(addr, role);
  }

  function rm_role(address addr, bytes32 role) constant returns (bool removed) {
    return perm.rm_role(addr, role);
  }
}

