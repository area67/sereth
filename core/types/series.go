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
)

//Definition of seriesNode
type seriesNode struct {
	hash, fromAddress, inputAddress, functionName, mark, val, oldMark []byte	//Data fields
	nextTxn                                             []*TransactionObject	//Array of any susequent transactions
	prevTxn                                             *TransactionObject		//Pointer to previous transactions
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

//Insert should add a seriesNode to the series at the appropriate location in the tree
func (s Series) Insert() bool {
	return false;
}

//Returns the transaction at the end of the series according to the desired heuristic
//(The default heuristic is to return the transaction at the end of the longest branch)
func (s Series) GetTailOfSeries() {

}