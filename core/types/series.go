// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
    "bytes"
    "fmt"
    "sync/atomic"
    "unsafe"
    "os"
    "log"
    "encoding/hex"
    //"strconv"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto/sha3"
)

var specialMark []byte = common.FromHex("0x7261614d61726b00000000000000000000000000000000000000000000000000")
var specialVal []byte = common.FromHex("0x72616156616c7565000000000000000000000000000000000000000000000000")

//Definition of SeriesNode
type SeriesNode struct {
    //Data fields
    fromAddress, inputAddress, functionName, mark, val,oldMark []byte
    //Array of any susequent transactions
    nextTxn                                             []*SeriesNode
    //Pointer to previous transactions
    prevTxn                                             *SeriesNode
    //Node depth
    depth                                               int32
    hash						[]byte
}

//Constructor for SeriesNode Objects
func NewSeriesNode() SeriesNode {
	n := SeriesNode {
		fromAddress:	nil,
		inputAddress:	nil,
		functionName:	nil,
		mark:		nil,
		val:		nil,
		oldMark:	nil,
		nextTxn:	make([]*SeriesNode, 100, 100),
	        prevTxn:	nil,
	        depth:          -1,
	}
	return n
}

//Configures the head node for initialization
func MakeHead() SeriesNode {
    n := SeriesNode {
		fromAddress:	nil,
		inputAddress:	common.FromHex("0x0"),
		functionName:	nil,
		mark:		common.FromHex("0x7261614d61726b00000000000000000000000000000000000000000000000000"),
		val:		common.FromHex("0x72616156616c7565000000000000000000000000000000000000000000000000"),
		oldMark:	nil,
		nextTxn:	make([]*SeriesNode, 100, 100),
		prevTxn:	nil,
	        depth:          0,
	        hash:           common.FromHex("0x0"),
    }
    return n
}

//Definition of Series
type Series struct {
    Head *SeriesNode
    Tail *SeriesNode
    RawPool [][]*SeriesNode
}

//Constructor for Series Objects
func NewSeries(numThreads int) Series {
	s := Series{nil, nil, make([][]*SeriesNode, numThreads)}
	return s
}

func (s *Series) parseTxPool(txns []*Transaction, tid int, num_threads int) {
    interval := len(txns)/num_threads
    start_index := interval * tid
    end_index := start_index+interval
    s.RawPool[tid] = s.RawPool[tid][:0]

    if tid == num_threads - 1 {
	    end_index = len(txns)
    }

	//Decode transaction payloads
	for i := start_index; i < end_index; i++ {
		txn := txns[i]

		data := txn.data.Payload

		if len(data) < 100 {
			continue
        }
        
        if s.findHash(txn.hash.Load().(common.Hash).Bytes(), s.Head) {
            continue
        }

        var name []byte = data[0:4]
        var addr []byte = data[4:36]
		var oldMark []byte = data[36:68]
        var val []byte = data[68:100]

		//Check function signature
        if bytes.Equal(name, common.FromHex("d1602737")) || bytes.Equal(name, common.FromHex("3f91e238")) || bytes.Equal(name, common.FromHex("c32bc356")) {
            var txnObj = NewSeriesNode()
            txnObj.hash = txn.hash.Load().(common.Hash).Bytes() //common.FromHex(strconv.Itoa(i))
            txnObj.val = val
            txnObj.oldMark = oldMark
            txnObj.mark = Keccak256(oldMark, val)
            txnObj.inputAddress = addr
            txnObj.functionName = name

            s.RawPool[tid] = append(s.RawPool[tid], &txnObj)
        }
    }
}

/*Insert should add a SeriesNode to the series at the appropriate
 *location in the tree
 */

func (s *Series) InsertTxn(n *SeriesNode) bool {

    //parent := s.Head.findParent(n)
    parent := s.findParent(n)
    n.prevTxn = parent

    if parent == nil {
        return false;
    }

    for {
        for i, _ := range parent.nextTxn {
            item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&parent.nextTxn[i])))
            if item != nil && bytes.Equal((*SeriesNode)(item).hash, n.hash) {
                fmt.Println("Returning after dupe txn")
                return false
            }
            if item == nil {
                ret := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&parent.nextTxn[i])), item, unsafe.Pointer(n))
                if ret {
                    parent_depth := atomic.LoadInt32(&parent.depth)
                    if parent_depth != -1 {
                        s.finishInserting(n, parent_depth+1)
                    }
                    return true;
                }
            }
        }
    }

    return false
}

func (s *Series) finishInserting(n *SeriesNode, d int32) {
    leaf := true
    depth := atomic.LoadInt32(&n.depth)
    if depth == -1 {
        atomic.StoreInt32(&n.depth, d)
        for i, _ := range n.nextTxn {
            item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.nextTxn[i])))
            if (item != nil) {
                leaf = false
                s.finishInserting((*SeriesNode)(item), d+1)
            }
        }

        if leaf {
            tail := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.Tail)))
            if tail == nil || (*SeriesNode)(tail).depth < d {
                atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.Tail)), tail, unsafe.Pointer(n))
            }
        }
    }
}


func (s *Series) findParent(new_node *SeriesNode) (*SeriesNode) {
    node := s.findParentFromPool(new_node)

    if (node == nil) {
        node = s.findParentFromDag(new_node, s.Head)
    }

    return node
}

func (s *Series) findParentFromDag(new_node *SeriesNode, candidate *SeriesNode) (*SeriesNode) {

    if bytes.Equal(new_node.oldMark, candidate.mark) {
        return candidate
    }

    for i, _ := range candidate.nextTxn {
        item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&candidate.nextTxn[i])))
        if item != nil {
            parent := s.findParentFromDag(new_node, (*SeriesNode)(item))

            if parent != nil {
                return parent
            }
        }
    }

    return nil
}

func (s *Series) findParentFromPool(new_node *SeriesNode) (*SeriesNode) {
    for i,_ := range s.RawPool {
        for k,_ := range s.RawPool[i] {
            item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.RawPool[i][k])))

            if bytes.Equal(new_node.oldMark, (*SeriesNode)(item).mark) {
                return (*SeriesNode)(item)
            }
        }
    }

    return nil
}

//Finds the most recent set in the series
func (s *Series) findRecentSet() *SeriesNode {
    curr := s.Tail
    for curr != nil {
        if bytes.Equal(curr.functionName,common.FromHex("d1602737")) {
            return curr
        }

        curr = curr.prevTxn
    }
    
    return nil
}

func (s *Series) findHash(target []byte, curr *SeriesNode) (bool) {
    if bytes.Equal(curr.hash,target) {
        return true
    }

    for i, _ := range curr.nextTxn {
        item := curr.nextTxn[i]
        if item != nil {
            ret := s.findHash(target, (*SeriesNode)(item))
            if ret {
                return true
            }
        }
    }

    return false
}

func (s *Series) IsRAA(input []byte) bool {
	f, ferr := os.OpenFile("../../interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		log.Fatal("Cannot open file", ferr)
	}
	defer f.Close()

	_, ferr = f.WriteString("\n\nRAA check with input\n")
	_, ferr = f.WriteString(hex.EncodeToString(input))

	if len(input) >= 100 {
		sig := hex.EncodeToString(input[0:4])
		_, ferr = f.WriteString("\nFunction Signature: ")
		_, ferr = f.WriteString(sig)
		if sig == "6c58228a" || sig == "07173de5" || sig == "152227ad" || sig == "19608715" || sig == "dcfef6fb" || sig == "e4472525" {
			// <deprecated> ||       getMark(raa) ||      <deprecated> ||     set(amv, raa) ||       getAMV(raa) ||       mark(p)
			_, ferr = f.WriteString(", RAA Requested!\n")
			return true
		}
	}

	return false
}

func IsHMS (input []byte) bool {
	if len(input) >= 100 {
        sig := hex.EncodeToString(input[0:4])
        if sig == "d1602737" || sig == "c32bc356" {
        //         set                 buy
            return true
        }
    }

    return false
}

func (s *Series) DoRAA(input []byte, txnList []*Transaction) []byte {
	f, ferr := os.OpenFile("../../interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		log.Fatal("Cannot open file", ferr)
	}

	_, ferr = f.WriteString(fmt.Sprintf("txnList size : %d\n", len(txnList)))

	s.parseTxPool(txnList, 0, 1)

	for i := 0; i < len(s.RawPool[0]); i++ {
		var res = s.InsertTxn(s.RawPool[0][i])
		if(res) {
			_, ferr = f.WriteString(fmt.Sprintf("Transaction inserted %x\n", s.RawPool[0][i].mark))
		} else {
			_, ferr = f.WriteString(fmt.Sprintf("Transaction insertion failed %x\n", s.RawPool[0][i].mark))
		}
    }

    //Rewrite the input (raa step)
    var amv [][]byte

    //EMPTY DAG CASE
    if s.Tail == nil {
        amv = [][]byte{common.FromHex("0x7261614164647265737300000000000000000000000000000000000000000000"), common.FromHex("0x7261614d61726b00000000000000000000000000000000000000000000000000"), common.FromHex("0x72616156616c7565000000000000000000000000000000000000000000000000")}
    } else {
        rSet := s.findRecentSet()
        //NO SETS CASE (Only buys)
        if rSet == nil {
            amv = [][]byte{s.Tail.inputAddress, s.Tail.mark, specialVal}
        } else { //SET EXISTS CASE
            amv = [][]byte{s.Tail.inputAddress, s.Tail.mark, rSet.val}
        }
    }

    //Rewrite input using the AMV
	for i := 0; i < 3; i++ {
		for k := 0; k < 32; k++ {
			input[(len(input)-96)+(i*32)+k] = amv[i][k]
		}
	}

	_, ferr = f.WriteString(fmt.Sprintf("RAA: %x\n\n\n", input))

	return input
}

func (s *Series) verifyTree(curr *SeriesNode, count *int) {
    if curr.prevTxn != nil && !bytes.Equal(curr.prevTxn.mark, curr.oldMark) {
        fmt.Println("Error, unmatching nodes")
        return
    }

    (*count)++

    for i, _ := range curr.nextTxn {
        item := curr.nextTxn[i]
        if item != nil {
            s.verifyTree(item, count)
        }
    }
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.New256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func searchtxns(hash []byte, txns []*Transaction) bool {
    for i, _ := range txns {
        if bytes.Equal(txns[i].Hash().Bytes(),hash) {
            return true
        }
    }

    return false
}

//Return only the series as a list of Transactions for semantic mining
func (s *Series) series(txns []*Transaction) []*SeriesNode {
    var seriesList = make([]*SeriesNode,0,len(txns))
    
    curr := s.Tail
    for curr != nil {
        if searchtxns(curr.hash, txns) == true {
            seriesList = append(seriesList, curr)
        } else {
            break
        }
        curr = curr.prevTxn
    }

	return seriesList
}

