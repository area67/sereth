pragma solidity ^0.4.20;

// Sereth contract provides a consistent intra-block view of protected state variable p and serializes operations on it.

contract Sereth {

    bytes32 s0 = 'raaAddress';
    bytes32 s1 = 'raaMark';
    bytes32 s2 = 'raaValue';
    bytes32[3] p = [s0, s1, s2];
    uint256 sold = 0;

// Mark, Set and Get are methods in the Hash-Mark-Set transactional data structure.
/*
    function mark(bytes32[3] raa) private pure returns(bytes32) {
        return raa[1];
    }
    function set(bytes32[3] amv, bytes32[3] raa) public returns (bytes32[3]) {
        // verify the mark matches the longest tail from RAA, or the transaction is a head candidate
        if ((keccak256(amv[1]) == keccak256(raa[1])) || 
            ((keccak256(amv[1]) == keccak256(p[1])) && keccak256(raa[1]) == keccak256(s1))) { 
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
*/
    function get(bytes32[3] raa) public pure returns(bytes32) {
        return raa[2];
    }
    function getAMV(bytes32[3] raa) public pure returns(bytes32[3]) {
        return raa;
    }

// Get and Set used to change block state of descriptor object p for tests or to reinitialize.
// Use contract owner only modifier to restrict access in production deployments.

    function getAMV_blk() public view returns(bytes32[3]) {
        return p;
    }
    function setAMV_blk(bytes32[3] raa) public returns (bytes32[3]) {
        p = raa;
        return raa;
    }

// This section demonstrates a dynamic pricing use case for the Hash-Mark-Set transactional data structure

    function buy(bytes32[3] offer) public { 
        if ((keccak256(offer[1]) == keccak256(p[1])) && (keccak256(offer[2]) == keccak256(p[2]))) {
            // mark and price are valid: execute buy order and set new values for price descriptor object
            sold++;
            p[0] = bytes32(msg.sender);
            p[1] = keccak256(offer[1], offer[2]);
            p[2] = bytes32(uint256(offer[2]) + uint256(1));
        }
    }
    function qytSold() public view returns(uint256) {
        return sold;
    }
}

