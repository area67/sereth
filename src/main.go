
package main

import (
	"common"
	"math/rand"
	"strconv"
	"fmt"
	"time"
	"os"
    "sync"

	"golang.org/x/crypto/sha3"
)

var rawPool []*Transaction
var TxPool []*seriesNode
var blockSize = 100;
var txPoolSize = 16384
var started = true
var s Series
var mutex = &sync.Mutex{}
var numThreads = 4
var ops []int


func main() {
	numThreads, _ = strconv.Atoi(os.Args[1])
	ops = make([]int, numThreads, numThreads)
	alg := os.Args[2]
	fillTransactionPool(txPoolSize)

	if alg == "LF" {
		testLFHMS()
	} else if alg == "BL" {
		testBLHMS()
	} else {
		fmt.Println("Error: no valid algorithm selected")
		return
	}
}

func testBLHMS() {
	f, _ := os.OpenFile("BLHMS.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	start := time.Now()
	var parsedList = parseTxPoolSeq(rawPool)
	var _,_,_ = findOrder(parsedList)
	totalTime := time.Since(start)

	fmt.Println(totalTime)
	f.WriteString(strconv.Itoa(1) + "\t" + totalTime.String() + "\n")
}

func testLFHMS() {
	//Open output file
	f, _ := os.OpenFile("LFHMS.out", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	//Create data structure
	s = newSeries(numThreads)

	//Create waitgroup
	var wg sync.WaitGroup
	for w := 0; w < numThreads; w++ {
		wg.Add(1)
	}

	//begin parseTx, start timer
	start := time.Now()
	for w := 0; w < numThreads; w++ {
		go doParseTxPool(w, &wg)
	}
	wg.Wait()

	//Stop timer
	totalTime := time.Since(start)

	//Verify
	j := 0
	for k:=0; k < numThreads; k++ {
		for i:=0; i < len(s.RawPool[k]); i++ {
			if s.RawPool[k][i].hash != TxPool[j].hash {
				fmt.Println("Txpool Parsing fail ", s.RawPool[k][i].hash, " ", TxPool[j].hash)
				return
			}
			j++
			//fmt.Println("p:",s.RawPool[k][i].hash)
			//fmt.Println("s:",TxPool[(k*txPoolSize/numThreads) + i].hash)
		}
	}
	//fmt.Println("Tx parsing success")

	//Setup Head
	s.Head = s.RawPool[0][0]
	s.Head.depth = 0
	s.RawPool[0] = s.RawPool[0][1:]

	//Reset Waitgroup
	for w := 0; w < numThreads; w++ {
		wg.Add(1)
	}

	//build DAG, resume timer
	start = time.Now()
	for w := 0; w < numThreads; w++ {
		go buildDag(w, &wg)
	}
	wg.Wait()

	//Final Timer
	totalTime = totalTime + time.Since(start)

	//Tally final stats
	totalOps := 0
	for _, val := range ops {
		totalOps += val
	}

	//Verify Tree
	count := 0
	s.verifyTree(s.Head, &count)

	//fmt.Println("Size of Dag: ", count)
	//fmt.Printf("ops completed: %d\n", totalOps+1)
	//fmt.Println("tail: ", s.Tail.hash)
	fmt.Println(totalTime)
	f.WriteString(strconv.Itoa(numThreads) + "\t" + totalTime.String() + "\n")
}

func doParseTxPool(id int, wg *sync.WaitGroup) {
	s.parseTxPool(rawPool, id, numThreads)
	wg.Done()
}

func buildDag(id int, wg *sync.WaitGroup) {
	var localOps = 0
	//fmt.Printf("Worker %d entering\n", id)
	
    for i := 0; i < len(s.RawPool[id]); i++ {

		new_node := s.RawPool[id][i]
		if !s.InsertTxn(new_node) {
			fmt.Println("Error");
		} else {
			localOps++
		}
	}

	wg.Done()
	ops[id] = localOps
}
func Keccak256(data ...[]byte) []byte {
	d := sha3.New256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func SliceDelete(s []seriesNode, index int) []seriesNode {
    ret := make([]seriesNode, 0)
    ret = append(ret, s[:index]...)
    return append(ret, s[index+1:]...)
}

//Create x sereth transactions and fill the pool with them
//Function will ensure that each transaction has at least some parent transaction so that it will contribute
//meaningfully to the sereth DAG
func fillTransactionPool(x int) {
	signature := common.FromHex("d1602737")

	//Create transacitions and transaction parentage
	for i:=0; i < x; i++{
		n := NewSeriesNode()
		n.hash.Store(i)
		n.val = common.FromHexFixed(strconv.Itoa(rand.Int()%10000), 32)
		n.inputAddress = common.FromHexFixed("0x0000000000000000000000000000000000000000000000000000000000077777", 32)

		if i == 0 {
			n.oldMark = common.FromHexFixed("0x0", 32)
		} else {
			n.oldMark = TxPool[rand.Int()%len(TxPool)].mark
		}

		n.mark = Keccak256(n.oldMark, n.val)

		TxPool = append(TxPool, &n)
	}


	//Create simulated geth txpool
	for i:=0; i < x; i++{
		d := txdata { Payload: nil}
		n := Transaction { data: d }

		payload := make([]byte, 0)

		payload = append(payload, signature...)
		payload = append(payload, TxPool[i].inputAddress...)
		payload = append(payload, TxPool[i].oldMark...)
		payload = append(payload, TxPool[i].val...)

		d.Payload = payload;
		n.data = d
		n.hash.Store(i)
		rawPool = append(rawPool, &n);
	}
	
	return
}
