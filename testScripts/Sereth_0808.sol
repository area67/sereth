pragma solidity ^0.4.20;

contract Sereth {

    bytes32[3] p;

    function get(bytes32[3] raa) public pure returns(bytes32) {
        return raa[2];
    }
    function set(bytes32[3] amv, bytes32[3] raa) public returns (bytes32[3]) {
        // if (keccak256(amv[1]) == keccak256(getMark([0,0,0]))) {
        if (keccak256(amv[1]) == keccak256(raa[1])) {
            // mark valid, set new values; return: address, new mark, new price
            p[0] = bytes32(msg.sender);
            amv[0] = bytes32(msg.sender);
            p[1] = keccak256(amv[1], amv[2]);
            amv[1] = keccak256(amv[1], amv[2]);
            p[2] = amv[2];
            return amv; 
        }
        else {
            // mark not valid, no state change; return: address, valid mark, current price
            return raa;
        }
    }
    function getMark(bytes32[3] raa) public pure returns(bytes32) {
        return raa[1];
    }

    function get_blk() public view returns (bytes32) {
        return p[2];
    }
    function set_blk(uint256 v) public returns (bytes32) {
        p[2] = bytes32(v);
        return bytes32(v);
    }
    function getMark_blk() public view returns (bytes32) {
        return p[1];
    }


    function buy(bytes32[3] offer) public returns (bytes32[3]) { 
        return offer
    }


}

