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
//package main

import (
    "bytes"
    "strconv"
    "fmt"
)

//Definition of seriesNode
type seriesNode struct {
	//Data fields
	hash, fromAddress, inputAddress, functionName, mark, val,oldMark []byte
    //Array of any susequent transactions
    nextTxn                                             []*seriesNode
    //Pointer to previous transactions
    prevTxn                                             *seriesNode
}

//Constructor for seriesNode Objects
func NewSeriesNode() seriesNode {
	n := seriesNode {
		hash:			nil,
		fromAddress:	nil,
		inputAddress:	nil,
		functionName:	nil,
		mark:			nil,
		val:			nil,
		oldMark:		nil,
		nextTxn:		nil,
		prevTxn:		nil,
	}
	return n
}

//Definition of Series
type Series struct {
	Head *seriesNode
}

//Constructor for Series Objects
func newSeries() Series {
	s := Series{nil}
	return s
}

/*Insert should add a seriesNode to the series at the appropriate
 *location in the tree
 */
func (s Series) InsertTxn(n seriesNode) bool {
	return false
}

// Insert a node in the nextTxn
func (s seriesNode) Insert(n *seriesNode) bool {
    s.nextTxn = append(s.nextTxn, n);
    n.prevTxn = &s
    return true
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func (s seriesNode) findParent(data []byte, parent []*seriesNode) ([]*seriesNode) {

    if bytes.Equal(data, s.hash) {
        parent = append(parent, s.prevTxn)

        return parent
    }

    for _, n := range s.nextTxn {
        if len(parent) == 0 {
             parent = n.findParent(data, parent)
        } else {
            return parent
        }

    }

    return parent
}


func (s seriesNode) findMaxDepth() int {

    if len(s.nextTxn) == 0 {
        return 0
    }

    maxDepth := 0
    for _, n := range s.nextTxn {
        maxDepth = max(maxDepth, n.findMaxDepth())
    }

    return maxDepth + 1
}

/*Returns the transaction at the end of the series according to the 
 *desired heuristic
 */

/*(The default heuristic is to return the transaction at the end of
 *the longest branch)
 */
func (s seriesNode) getTailOfSeries(currentDepth int, maxDepth int, result []*seriesNode) []*seriesNode {

    // fmt.Println(currentDepth, s.nextTxn)

    if currentDepth == maxDepth {
        result = append(result, &s)

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

func main() {
    fmt.Println("in main")

    var nodes []*seriesNode
    for i:=0; i < 10; i++ {
        n := NewSeriesNode();
        n.hash = append(n.hash, []byte(strconv.Itoa(i))[0])
        nodes = append(nodes, &n)
        fmt.Printf("%d:%p\n", i, &n)
    }

    // fmt.Println("nodes", nodes)

    // build the tree
    node1 := NewSeriesNode()
    fmt.Println("node 1", nodes[0])
    node1.nextTxn = append(node1.nextTxn, nodes[0])
    node1.nextTxn = append(node1.nextTxn, nodes[1])
    node1.nextTxn = append(node1.nextTxn, nodes[2])
    nodes[0].prevTxn = &node1
    nodes[1].prevTxn = &node1
    nodes[2].prevTxn = &node1
    printSlice(node1.nextTxn)

    node2 := node1.nextTxn[1]
    node2.nextTxn = append(node2.nextTxn, nodes[3])
    node2.nextTxn = append(node2.nextTxn, nodes[4])
    nodes[3].prevTxn = node2
    nodes[4].prevTxn = node2
    printSlice(node2.nextTxn)

    node3 := node2.nextTxn[0]
    node3.nextTxn = append(node3.nextTxn, nodes[5])
    node3.nextTxn = append(node3.nextTxn, nodes[6])
    nodes[5].prevTxn = node3
    nodes[6].prevTxn = node3
    printSlice(node3.nextTxn)

    node4 := node2.nextTxn[1]
    node4.nextTxn = append(node4.nextTxn, nodes[7])
    nodes[7].prevTxn = node4
    printSlice(node4.nextTxn)

    node5 := node4.nextTxn[0]
    node5.nextTxn = append(node5.nextTxn, nodes[8])
    nodes[8].prevTxn = node5
    printSlice(node5.nextTxn)

   fmt.Println("node5", &nodes[5])
   fmt.Println("node6", &nodes[6])
   fmt.Println("node7", &nodes[7])

    fmt.Println("node8", &nodes[8])
    series_ := newSeries()

    series_.Head = &node1
    fmt.Println("series", series_)

    m := node1.findMaxDepth();
    fmt.Println("max depth", m)
    var tails []*seriesNode
    tails = node1.getTailOfSeries(0, 3, tails)
    fmt.Println("tails", tails)

    d := node4.hash
    var parent []*seriesNode
    parent = node1.findParent(d, parent)
    fmt.Println("parent", parent)
    fmt.Println("node2", node4)

    m1 := parent[0].findMaxDepth();
    fmt.Println("max depth", m1)
    var tails1 []*seriesNode
    tails1 = parent[0].getTailOfSeries(0, 3, tails1)
    fmt.Println("tails1", tails1)

    //node9 := nodes[9]
}
