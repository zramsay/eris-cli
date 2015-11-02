contract SimpleStorage {
    string storedData;
    bool called;

    function get() constant returns (string retVal) {
        if (called == true) {
            storedData = "blue";
        } else {
            storedData = "red";
            called = true;
        }

        return storedData;
    }
}