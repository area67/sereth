pragma solidity ^0.4.20;

library Sereth {

    struct HMS_uint {
        bytes32 addr; // = 'address';
        bytes32 mark; // = 'startHMS';
        uint256 value; // = 0;
    }

    function getMark(HMS_uint raa) internal pure returns(bytes32) {
        return raa.mark;
    }
    function get(HMS_uint raa) internal pure returns(uint256) {
        return raa.value;
    }
    function set(HMS_uint storage self, bytes32 m, uint v, HMS_uint raa) internal returns (bytes32, bytes32, bytes32) {
        if (keccak256(raa.mark) == keccak256(m)) {
            // mark valid, set new values; return: address, new mark, new price
            self.mark = keccak256(m, v);
            self.value = v;
            return (bytes32(msg.sender), keccak256(m,v), bytes32(v));
        }
        else {
            // mark not valid, no state change; return: address, valid mark, old price
            return (bytes32(msg.sender), self.mark, bytes32(self.value));
        }
    }

    function getMark_old(HMS_uint storage self) internal view returns(bytes32) {
        return self.mark;
    }
    function get_old(HMS_uint storage self) internal view returns(uint256) {
        return self.value;
    }
    function set_old(HMS_uint storage self, uint256 v) internal returns(uint256) {
        self.value = v;
        return v;
    }

}

