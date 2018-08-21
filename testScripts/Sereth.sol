pragma solidity ^0.4.20;

// Sereth contract provides a consistent intra-block view of protected state variable p and serializes operations on it.

contract Sereth {

    bytes32 s0 = 0x0000000000000000000000000000000000000000000000000000000000077777;
    bytes32 s1 = 'raaMark';
    bytes32 s2 = 0x0000000000000000000000000000000000000000000000000000000000000010;
    bytes32[3] p = [s0, s1, s2];
    uint256 nBuy = 0;
    uint256 nSet = 0;

// Mark, Set and Get are methods in the Hash-Mark-Set transactional data structure.

    function mark(bytes32[3] raa) private pure returns(bytes32) {
        return raa[1];
    }
    function set(bytes32[3] amv) public {
        // mark needs to match the intra-block mark, which is likely if obtained recently 
        if (keccak256(amv[1]) == keccak256(p[1])) { 
            // mark valid, set new mark and value
            nSet++;
            p[0] = bytes32(msg.sender);
            p[1] = keccak256(amv[1], amv[2]);
            p[2] = amv[2];
        }
    }
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
            // mark and price are valid: execute serialized buy order
            nBuy++;
            p[0] = bytes32(msg.sender);
            p[1] = keccak256(offer[1], offer[2]);
        }
    }
    function buyL(bytes32[3] offer) public { 
        if (keccak256(offer[2]) == keccak256(p[2])) {
            // price is valid: execute linearized buy order
            nBuy++;
            p[0] = bytes32(msg.sender);
            p[1] = keccak256(offer[1], offer[2]);
        }
    }
    function nTX() public view returns(uint256, uint256) {
        return (nBuy, nSet);
    }
}

