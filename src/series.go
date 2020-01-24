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

package main
//package main

import (
    "bytes"
    "fmt"
    "sync/atomic"
    "unsafe"
    "common"
)

//Geth Transaction data type
type Transaction struct {
    data txdata
    hash atomic.Value
}

type txdata struct {
	Payload      []byte
}

//Definition of seriesNode
type seriesNode struct {
    //Data fields
    fromAddress, inputAddress, functionName, mark, val,oldMark []byte
    //Array of any susequent transactions
    nextTxn                                             []*seriesNode
    //Pointer to previous transactions
    prevTxn                                             *seriesNode
    //Node depth
    depth                                               int32
    //Hash
    hash                                                atomic.Value
}

//Constructor for seriesNode Objects
func NewSeriesNode() seriesNode {
	n := seriesNode {
		fromAddress:	nil,
		inputAddress:	nil,
		functionName:	nil,
		mark:			nil,
		val:			nil,
		oldMark:		nil,
		nextTxn:		make([]*seriesNode, 20, 20),
        prevTxn:		nil,
        depth:          -1,
	}
	return n
}

//Definition of Series
type Series struct {
    Head *seriesNode
    Tail *seriesNode
    RawPool [][]*seriesNode
}

//Constructor for Series Objects
func newSeries(numThreads int) Series {
	s := Series{nil, nil, make([][]*seriesNode, numThreads)}
	return s
}

func (s *Series) parseTxPool(txns []*Transaction, tid int, num_threads int) {
    interval := len(txns)/num_threads
    start_index := interval * tid
    end_index := start_index+interval

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

        var name []byte = data[0:4]
        var addr []byte = data[4:36]
		var oldMark []byte = data[36:68]
        var val []byte = data[68:100]
        
		//Check function signature
		if bytes.Equal(name, common.FromHex("d1602737")) || bytes.Equal(name, common.FromHex("3f91e238")) || bytes.Equal(name, common.FromHex("c32bc356")) {
            var txnObj = NewSeriesNode()
            txnObj.hash = txn.hash
            txnObj.val = val
            txnObj.oldMark = oldMark
            txnObj.mark = Keccak256(oldMark, val)
            txnObj.inputAddress = addr

            s.RawPool[tid] = append(s.RawPool[tid], &txnObj)
		}
    }
}

/*Insert should add a seriesNode to the series at the appropriate
 *location in the tree
 */
func (s *Series) InsertTxn(n *seriesNode) bool {

    //parent := s.Head.findParent(n)
    parent := s.findParent(n)
    n.prevTxn = parent

    if parent == nil {
        return false;
    }

    for {
        for i, _ := range parent.nextTxn {
            item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&parent.nextTxn[i])))
            if item != nil && (*seriesNode)(item).hash == n.hash {
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

func (s *Series) finishInserting(n *seriesNode, d int32) {
    leaf := true
    depth := atomic.LoadInt32(&n.depth)
    if depth == -1 {
        atomic.StoreInt32(&n.depth, d)
        for i, _ := range n.nextTxn {
            item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.nextTxn[i])))
            if (item != nil) {
                leaf = false
                s.finishInserting((*seriesNode)(item), d+1)
            }
        }

        if leaf {
            tail := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.Tail)))
            if tail == nil || (*seriesNode)(tail).depth < d {
                atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.Tail)), tail, unsafe.Pointer(n))
            }
        }
    }
}

func (s *Series) findParent(new_node *seriesNode) (*seriesNode) {
    node := s.findParentFromPool(new_node)

    if (node == nil) {
        node = s.findParentFromDag(new_node, s.Head)
    }

    return node
}

func (s *Series) findParentFromDag(new_node *seriesNode, candidate *seriesNode) (*seriesNode) {

    if bytes.Equal(new_node.oldMark, candidate.mark) {
        return candidate
    }

    for i, _ := range candidate.nextTxn {
        item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&candidate.nextTxn[i])))
        if item != nil {
            parent := s.findParentFromDag(new_node, (*seriesNode)(item))

            if parent != nil {
                return parent
            }
        }
    }

    return nil
}

func (s *Series) findParentFromPool(new_node *seriesNode) (*seriesNode) {
    for i,_ := range s.RawPool {
        for k,_ := range s.RawPool[i] {
            item := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.RawPool[i][k])))
            
            if bytes.Equal(new_node.oldMark, (*seriesNode)(item).mark) {
                return (*seriesNode)(item)
            }
        }
    }

    return nil
}

func (s *Series) findHash(target int, curr *seriesNode) (bool) {
    if curr.hash.Load() == target{
        return true
    }

    for i, _ := range curr.nextTxn {
        item := curr.nextTxn[i]
        if item != nil {
            ret := s.findHash(target, (*seriesNode)(item))
            if ret {
                return true
            }
        }
    }

    return false
}

func (s *Series) verifyTree(curr *seriesNode, count *int) {
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

/*(The default heuristic is to return the transaction at the end of
 *the longest branch)
 */
func (s *seriesNode) getTailOfSeries(currentDepth int, maxDepth int, result []*seriesNode) []*seriesNode {

    // fmt.Println(currentDepth, s.nextTxn)

    if currentDepth == maxDepth {
        result = append(result, s)

        return result
    }

    if len(s.nextTxn) == 0 {
        return result
    }

    for _, n := range s.nextTxn {
        result = n.getTailOfSeries(currentDepth + 1, maxDepth, result)
    }

    return result
}

func printSlice(s []*seriesNode) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
