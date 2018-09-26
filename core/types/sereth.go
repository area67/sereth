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
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"os"
)

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

type TransactionObject struct {
	hash, fromAddress, inputAddress, functionName, mark, val, oldMark []byte
	rpc						    *RPCTransaction
	nextTxn                                             []*TransactionObject
	prevTxn                                             *TransactionObject
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *Transaction, blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	var signer Signer = FrontierSigner{}
	if tx.Protected() {
		signer = NewEIP155Signer(tx.ChainId())
	}
	from, _ := Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}
	return result
}

/*var head = TransactionObject{
mark: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
fromAddress: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
val: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
nextTxn: make([]*TransactionObject, 0, 100)}*/

var inAddrRC []byte = common.FromHex("0x0000000000000000000000000000000000000000000000000000000000077777")
var inAddrRU []byte = common.FromHex("0x0000000000000000000000000000000000000000000000000000000000088888")
var specialMark []byte = common.FromHex("0x7261614d61726b00000000000000000000000000000000000000000000000000")
var specialVal []byte = common.FromHex("0x72616156616c7565000000000000000000000000000000000000000000000000")

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *Transaction) *RPCTransaction {
	return newRPCTransaction(tx, common.Hash{}, 0, 0)
}

func sliceAppend(slice []*TransactionObject, val *TransactionObject) []*TransactionObject {
	//ToDo: Increase size of slice if out of space
	slice = slice[0 : len(slice)+1]
	slice[len(slice)-1] = val
	return slice
}

func sliceDelete(slice []*TransactionObject) []*TransactionObject {
	slice[len(slice)-1] = nil
	slice = slice[0 : len(slice)-1]
	return slice
}

//Parse the payload and filter out unrelated transactions
func parseTransactions(txns []*RPCTransaction) ([]TransactionObject) {
	//Create a slice that will contain all filtered transactions
	f, ferr := os.OpenFile("/home/bitnami/interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		log.Fatal("Cannot open file", ferr)
	}

	var inputArray = make([]TransactionObject, len(txns))
	var k int = 0

	//Decode transaction payloads
	for i := 0; i < len(txns); i++ {
		var txn = txns[i]
		var data = txn.Input

		if len(data) < 100 {
			continue
		}

		var name []byte = data[0:4]
		var oldMark []byte = data[36:68]
		var val []byte = data[68:100]

		sig := hex.EncodeToString(name)
		_, ferr = f.WriteString(fmt.Sprintf("Parsing transaction\nVal: %x,\noldMark: %x\n", val, oldMark))

		//Filter transactions and add them to our slice
		// ourAddress := common.HexToAddress("0x48c1bdb954c945a57459286719e1a3c86305fd9e")
		// var to = *txn.To
		if sig == "d1602737" || sig == "c32bc356" {
			// && bytes.Compare(to.Bytes(), ourAddress.Bytes()) == 0 {
			var mark []byte = crypto.Keccak256(oldMark, val)
			var address = txn.From.Bytes()
			var paddedAddress = make([]byte, 32)
			copy(paddedAddress[32-len(address):], address)

			var txnObj = TransactionObject{nil, paddedAddress, data[4:36], name, mark, val, oldMark, txn, make([]*TransactionObject, 0, 100), nil}

			/*Filter out transactions that the anaylzer previously rejected
			if bytes.Equal(oldMark, raaMark) && bytes.Equal(val, raaVal) || (bytes.Equal(specialMark, raaMark) && bytes.Equal(specialVal, raaVal)) {
				inputArray[k] = txnObj
				k = k + 1
			}*/

			//Filter out rejects
			if bytes.Equal(data[4:36], inAddrRC) || bytes.Equal(data[4:36], inAddrRU) {
				inputArray[k] = txnObj
				k = k + 1
			}
		}
	}

	f.Close()

	//Return a slice the length of filtered values we obtained
	return inputArray[0:k]
}

func findOrder(txns []TransactionObject) (*TransactionObject, []*TransactionObject, int) {
	f, ferr := os.OpenFile("/home/bitnami/interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		log.Fatal("Cannot open file", ferr)
	}
	defer f.Close()

	//Create doubly linked graph
	for i := 0; i < len(txns); i++ {
		for k := 0; k < len(txns); k++ {
			if bytes.Equal(txns[i].mark, txns[k].oldMark) && i != k {
				txns[i].nextTxn = sliceAppend(txns[i].nextTxn, &txns[k])
				txns[k].prevTxn = &txns[i]
			}
		}
	}

	//Final all potential head values
	var candidateHeads = make([]TransactionObject, 25)
	var k int = 0
	for i := 0; i < len(txns); i++ {
		if bytes.Equal(txns[i].inputAddress, inAddrRC) {
			candidateHeads[k] = txns[i]
			k = k + 1
		}
	}
	candidateHeads = candidateHeads[0:k]

	_, ferr = f.WriteString(fmt.Sprintf("Number of candidate heads found: %d\n", k))

	var longestChainDepth int = 0
	var seriesHead TransactionObject
	var depth int = 1
        var path = make([]*TransactionObject, 0, 1000)
        var maxDepth int = 0
        var maxDepthPath = make([]*TransactionObject,0,1000)
	for i := 0; i < len(candidateHeads); i++ {
		depth = 1
		path = make([]*TransactionObject, 0, 1000)
		maxDepth = 0
		maxDepthPath = make([]*TransactionObject,0,1000)
		path = sliceAppend(path, &candidateHeads[i])

		findDeepestBranch(&candidateHeads[i], depth, &maxDepth, path, maxDepthPath[0:0])
		maxDepthPath = maxDepthPath[0:maxDepth]

		if depth > longestChainDepth {
			seriesHead = candidateHeads[i]
		}
	}

	f.Close()

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
		path = sliceAppend(path, head.nextTxn[i])
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

func IsRAA(input []byte) bool {
	f, ferr := os.OpenFile("/home/bitnami/interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
		   //     set                 buy
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

func DoRAA(input []byte, txnList []*RPCTransaction) []byte {
	f, ferr := os.OpenFile("/home/bitnami/interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		log.Fatal("Cannot open file", ferr)
	}

	var parsedList = parseTransactions(txnList)
	_, ferr = f.WriteString(fmt.Sprintf("Parsed list length: %d\n", len(parsedList)))
	if len(parsedList) == 0 {
		_, ferr = f.WriteString(fmt.Sprintf("Parsed list length: %d, returning 0 RAA\n\n", len(parsedList)))
		var defaults = [][]byte{common.FromHex("0x7261614164647265737300000000000000000000000000000000000000000000"), common.FromHex("0x7261614d61726b00000000000000000000000000000000000000000000000000"), common.FromHex("0x72616156616c7565000000000000000000000000000000000000000000000000")}
		_, ferr = f.WriteString(fmt.Sprintf("defaults: %v\n", defaults))
		for i := 0; i < 3; i++ {
			for k := 0; k < 32; k++ {
				input[(len(input)-96)+(i*32)+k] = defaults[i][k]
			}
		}
		return input
	}

	var head, series, seriesDepth = findOrder(parsedList)

	_, ferr = f.WriteString(fmt.Sprintf("Head node:\nmark: %x\nval: %x\n", head.mark, head.val))

	//Get last touple from series
	var n = series[seriesDepth-1]
	var m = findRecentSet(series)
	var array [][]byte

	if m == nil {
		array = [][]byte{n.fromAddress, n.mark, specialVal}
	} else {
		array = [][]byte{n.fromAddress, n.mark, m.val}
		_, ferr = f.WriteString(fmt.Sprintf("Deepest set node:\nval: %x\n", m.val))
	}

	_, ferr = f.WriteString(fmt.Sprintf("Deepest node:\nmark: %x\nval: %x\n", n.mark, n.val))

	for i := 0; i < 3; i++ {
		for k := 0; k < 32; k++ {
			input[(len(input)-96)+(i*32)+k] = array[i][k]
		}
	}

	_, ferr = f.WriteString(fmt.Sprintf("RAA: %x\n\n\n", input))

	return input

	//End Anaylzer
}

//Return only the series as a list of RPCTransactions
func series(txnList []*RPCTransaction) []*RPCTransaction {
	var parsedList = parseTransactions(txnList)
        var _, series, _ = findOrder(parsedList)
	var rpcList = make([]*RPCTransaction,len(series))

	for i := 0; i < len(series); i++ {
		rpcList[i] = series[i].rpc
	}

	return rpcList

}
