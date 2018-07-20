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

package  vm

import (
	"bytes"
	"fmt"
	"math/big"
        "os"
	"log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
	hash, fromAddress, functionName, mark, val, nextMark []byte
	nextTxn                                              []*TransactionObject
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
        var signer types.Signer = types.FrontierSigner{}
        if tx.Protected() {
                signer = types.NewEIP155Signer(tx.ChainId())
        }
        from, _ := types.Sender(signer, tx)
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

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *types.Transaction) *RPCTransaction {
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
func parseTransactions(txns []*RPCTransaction) []TransactionObject {
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

		if len(data) < 68 {
			continue
		}

		var name []byte = data[0:4]
		var mark []byte = data[4:36]
		var val []byte = data[36:68]

		_, ferr = f.WriteString(fmt.Sprintf("Parsing transaction\nMark: %d\nVal: %d\n", mark, val))

		//Filter transactions and add them to our slice
		ourAddress := common.HexToAddress("0x48c1bdb954c945a57459286719e1a3c86305fd9e")
		var to = *txn.To
		if bytes.Equal(name, common.FromHex("6c58228a")) && bytes.Compare(to.Bytes(), ourAddress.Bytes()) == 0 {
			var nextMark []byte = crypto.Keccak256(mark, val)
			inputArray[k] = TransactionObject{nil, nil, name, mark, val, nextMark, make([]*TransactionObject, 0, 100)}
			k = k+1
		}
	}

	f.Close()

	//Return a slice the length of filtered values we obtained
	return inputArray[0:k]
}

func findOrder(txns []TransactionObject) *TransactionObject {
	var head = TransactionObject{nextMark: common.FromHex("0x7374617274484d53000000000000000000000000000000000000000000000000"), nextTxn: make([]*TransactionObject, 0, 100)}

	for i := 0; i < len(txns); i++ {
		if bytes.Equal(head.nextMark, txns[i].mark) {
			head.nextTxn = sliceAppend(head.nextTxn, &txns[i])
		}
	}

	for i := 0; i < len(txns); i++ {
		for k := 0; k < len(txns); k++ {
			if bytes.Equal(txns[i].nextMark, txns[k].mark) && i != k {
				txns[i].nextTxn = sliceAppend(txns[i].nextTxn, &txns[k])
			}
		}
	}

	return &head
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

func Tuple(txP ContentFetcher) [][]byte {
	f, ferr := os.OpenFile("/home/bitnami/interpreter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if ferr != nil {
            log.Fatal("Cannot open file", ferr)
        }

	content := map[string]map[string]map[string]*RPCTransaction{
		"pending": make(map[string]map[string]*RPCTransaction),
		"queued":  make(map[string]map[string]*RPCTransaction),
	}

	pending, _ := txP.Content()

	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["pending"][account.Hex()] = dump
	}
/*
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["queued"][account.Hex()] = dump
	}
*/
        for  _, txs := range pending {
                //dump := content["queued"][account.Hex()]
                for _, tx := range txs {
                         //_, ferr = f.Write([]byte(dump[fmt.Sprintf("%d", tx.Nonce())].Input))
			_, ferr = f.WriteString(fmt.Sprintf("%d\n", tx.Nonce()))
                }
        }

	_, ferr = f.WriteString(fmt.Sprintf("Analyzer Begin\n"))
	//Begin Analyzer

	var txnList = make([]*RPCTransaction, 1000)
	var i int = 0

	for _, txs := range pending {
		for _, tx := range txs {
			txnList[i] = newRPCPendingTransaction(tx)
			i = i + 1
		}
	}

	//Slice such that length is equal to number of transactions in pending
	txnList = txnList[0:i]

	var parsedList = parseTransactions(txnList)
	var head = findOrder(parsedList)

	//Convert linked list into series
	var depth int = 1
	var path = make([]*TransactionObject, 0, 1000)
	var maxDepth int = 0
	var maxDepthPath [1000]*TransactionObject
	path = sliceAppend(path, head)

	findDeepestBranch(head, depth, &maxDepth, path, maxDepthPath[0:0])

	//Get last touple from series
	var n = maxDepthPath[maxDepth-1]

	_, ferr = f.WriteString(fmt.Sprintf("Deepest node:\nmark: %d\nval: %d\nnextmark: %d\n", n.mark, n.val, n.nextMark))

	var array = [][]byte{n.fromAddress, n.mark, n.val}

	return array

	//End Anaylzer
}
