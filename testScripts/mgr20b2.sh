#!/bin/bash

# blockchain address of sender and contract
contractAddress=$1
fromAddress=0x8691bf25ce4a56b15c1f99c944dc948269031801

# save the special values
raaAddr=7261614164647265737300000000000000000000000000000000000000000000
raaMark=7261614d61726b00000000000000000000000000000000000000000000000000
raaVal=72616156616c7565000000000000000000000000000000000000000000000000

# function signatures for input
setFuncSig=0xd1602737
grcFuncSig=0x1ae3304a
getFuncSig=0xdcfef6fb
buyFuncSig=0x3f91e238
brcFuncSig=0xc32bc356
qtyFuncSig=0xd59fe939

# initialize actual arguments
getFuncArg=0000000000000000000000000000000000000000000000000000000000022222
inAddrRC=0000000000000000000000000000000000000000000000000000000000077777
inAddrRU=0000000000000000000000000000000000000000000000000000000000088888

# set the RPC data input for get and grc 
getFuncInput=$getFuncSig$getFuncArg$getFuncArg$getFuncArg
#echo $getFuncInput
grcFuncInput=$grcFuncSig$getFuncArg$getFuncArg$getFuncArg
#echo $grcFuncInput
qtyFuncInput=$qtyFuncSig$getFuncArg$getFuncArg$getFuncArg
echo $qtyFuncInput

# set the initial counters and timers
echo -e "\nInitialize counters and timers"
qtyJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$qtyFuncInput'"}, "latest"], "id":1031}'`
#echo $qtyJSON
startBuy=0x`echo $qtyJSON | jq -r ".result" | cut -c 3-66`
startBuy=$(($startBuy))
#echo $startBuy
startSet=0x`echo $qtyJSON | jq -r ".result" | cut -c 67-130`
startSet=$(($startSet))
#echo $startSet
blockJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber", "params":["pending", true], "id":1}'`
startBlock=`echo $blockJSON | jq -r ".result" | cut -d "\"" -f 2`
startBlock=$(($startBlock))
#echo $startBlock
startTime=`date +%s`
tryBuy=0
#echo $tryBuy
trySet=0
#echo $trySet

# Loop and change the price from time to time
#for orderTarget in {2..100..2}
echo -e "\nLoop to send price changes from Peer 2 (Bob)"
while true
do
    # set the new value
    num=`hexdump -n 1 -e '1/2 "%02X"' /dev/random`
    inVal=00000000000000000000000000000000000000000000000000000000000001
    inVal=$inVal$num
    # get the latest mark
    AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
    while [[ ${#AMV} -lt 50 ]]
    do
        sleep 0.25
        AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
    echo $AMV
    done
    inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
    # fill in the address, and possibly replace the mark, based on the txpool state
    if [[ "$inMark" != "$raaMark" && "$inMark" != "$getFuncArg" ]]
    then
        inAddr=$inAddrRU
    else
        AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1013}'`
        inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
        inAddr=$inAddrRC
    fi
    read -n 1 -t 0.1 key
    if [[ $key = "q" ]] 
    then
       echo 
       break
    fi
    # send the transaction
    trySet=`expr $trySet + 1`
    echo "Manager trySet $trySet Time `date +%s`"
    setFuncInput=$setFuncSig$inAddr$inMark$inVal
    echo $setFuncInput | cut -c 3-
    curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$setFuncInput'"}],"id":1010}'
    sleep 2
done

# set the final counters and timers
echo -e "\nSet final counters and timers -- waiting for block updates"
read -rsp $'Hit any key to confirm block published \n' -n1 key
#echo "Get nTX"
qtyJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$qtyFuncInput'"}, "latest"], "id":1031}'`
#echo $qtyJSON
endBuy=0x`echo $qtyJSON | jq -r ".result" | cut -c 3-66`
endBuy=$(($endBuy))
#echo $endBuy
endSet=0x`echo $qtyJSON | jq -r ".result" | cut -c 67-130`
endSet=$(($endSet))
#echo $endSet
#echo "Get block"
blockJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber", "params":["pending", true], "id":1}'`
endBlock=`echo $blockJSON | jq -r ".result" | cut -d "\"" -f 2`
endBlock=$(($endBlock))
#echo $endBlock
endTime=`date +%s`

# show stats from this run
elapsedTime=`expr $endTime - $startTime`
elapsedBlocks=`expr $endBlock - $startBlock`
sets=`expr $endSet - $startSet`
buys=`expr $endBuy - $startBuy`
echo "Time"
echo "$startTime $endTime $elapsedTime seconds"
echo "Blocks"
echo "$startBlock $endBlock $elapsedBlocks blocks"
echo "Buys"
echo "$startBuy $endBuy $buys of $tryBuy"
echo "Sets"
echo "$startSet $endSet $sets of $trySet"

