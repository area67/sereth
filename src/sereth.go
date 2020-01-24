package main

import (
	"sync/atomic"
	"bytes"
	"common"
	"encoding/hex"
)

type TransactionObject struct {
	fromAddress, inputAddress, functionName, mark, val, oldMark []byte
	nextTxn                                             []*TransactionObject
	prevTxn                                             *TransactionObject
	hash												atomic.Value
}

func parseTxPoolSeq(txns []*Transaction) []TransactionObject{
	inputArray := make([]TransactionObject, 0)

    //Decode transaction payloads
	for i := 0; i < len(txns); i++ {
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
            var txnObj = TransactionObject{hash: txn.hash, val: val, oldMark: oldMark, mark: Keccak256(oldMark, val), inputAddress: addr, nextTxn: make([]*TransactionObject, 0, 20)}

            inputArray = append(inputArray, txnObj)
		}
	}
	return inputArray
}

func sliceDelete(slice []*TransactionObject) []*TransactionObject {
	slice[len(slice)-1] = nil
	slice = slice[0 : len(slice)-1]
	return slice
}

func findOrder(txns []TransactionObject) (*TransactionObject, []*TransactionObject, int) {
	//Create doubly linked graph
	for i := 0; i < len(txns); i++ {
		for k := 0; k < len(txns); k++ {
			if bytes.Equal(txns[i].oldMark, txns[k].mark) && i != k {
				txns[i].nextTxn = append(txns[i].nextTxn, &txns[k])
				txns[k].prevTxn = &txns[i]
				break
			}
		}
	}

	//Final all potential head values
	var candidateHeads = make([]TransactionObject, len(txns))
	var k int = 0
	for i := 0; i < len(txns); i++ {
		if txns[i].hash.Load() ==  0 {
			candidateHeads[k] = txns[i]
			k = k + 1
		}
	}
	candidateHeads = candidateHeads[0:k]

	var longestChainDepth int = 0
	var seriesHead TransactionObject
	var depth int
	var path []*TransactionObject
	var maxDepth int = 0
	var maxDepthPath []*TransactionObject
	for i := 0; i < len(candidateHeads); i++ {
		depth = 1
		path = make([]*TransactionObject, 0, 1000)
		maxDepth = 0
		maxDepthPath = make([]*TransactionObject,0,1000)
		path = append(path, &candidateHeads[i])

		findDeepestBranch(&candidateHeads[i], depth, &maxDepth, path, maxDepthPath[0:0])
		maxDepthPath = maxDepthPath[0:maxDepth]

		if depth > longestChainDepth {
			seriesHead = candidateHeads[i]
		}
	}

	return &seriesHead, maxDepthPath, maxDepth
}

func findDeepestBranch(head *TransactionObject, depth int, maxDepth *int, path, maxDepthPath []*TransactionObject) {
	if len(head.nextTxn) == 0 {
		if depth > *maxDepth {
			*maxDepth = depth
			maxDepthPath = maxDepthPath[0:*maxDepth]
			copy(maxDepthPath, path)
		}
		return
	}

	for i := 0; i < len(head.nextTxn); i++ {
		path = append(path, head.nextTxn[i])
		findDeepestBranch(head.nextTxn[i], depth+1, maxDepth, path, maxDepthPath)
		path = sliceDelete(path)
	}
	return
}

func isInSeries(txn *TransactionObject, series []*TransactionObject) bool {
	for i := 0; i < len(series); i++ {
		if bytes.Equal(txn.fromAddress, series[i].fromAddress) && bytes.Equal(txn.mark, series[i].mark) && bytes.Equal(txn.oldMark, series[i].oldMark) && bytes.Equal(txn.val, series[i].val) {
			return true
		}
	}
	return false
}

func findRecentSet(series []*TransactionObject) *TransactionObject {
	for i:= len(series)-1; i >= 0; i-- {
		if hex.EncodeToString(series[i].functionName) == "d1602737" {
			return series[i]
		}
	}

	return nil
}

func createSeries(parsedList []TransactionObject) []*TransactionObject {
	var _, series, _ = findOrder(parsedList)
	return series
}