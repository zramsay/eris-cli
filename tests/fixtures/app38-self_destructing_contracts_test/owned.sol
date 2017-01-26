contract owned {
    address owner;

    modifier onlyOwner() {
        if (msg.sender == owner) {
            _
        }
    }

    function owned() {
        owner = msg.sender;
    }

}