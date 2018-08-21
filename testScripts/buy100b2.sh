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

# Loop to send many transactions
echo -e "\nLoop to send transactions from Peer 2 (Bob)"
for i in {1..5}
do
    echo -e "\nTX $i Time `date +%s`"
    #read -rsp $'Hit any key to prepare offer \n' -n1 key
    AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
    while [[ ${#AMV} -lt 50 ]]
    do
        sleep 0.5
        AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
    #echo $AMV
    done

    inAddr=$inAddrRU
    inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
    # echo $inMark
    inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
    # echo $inVal

    if [[ "$inMark" != "$raaMark" && "$inMark" != "$getFuncArg" ]]
    then 
        if [ "$inVal" != "$raaVal" ]
        then # both inMark and inVal are Read Uncommitted values from the algorithm
            echo "IS intra-block transaction with intra block set -- keep Read Uncommitted mark and value; submit as RU"
            echo $inMark
            echo $inVal
        else # mark does not match special but the value does, this is a multiple buy case on the committed value, Read Committed value but keep the RU mark
            echo "MB multiple buys wiht value set previous to this block -- Read committed value but keep dynamic mark; submit as RU"
            AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1012}'`
            inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
            echo $inMark
            echo $inVal
        fi
    else # mark indicates nothing has happened since block was published, read committed mark and value
        echo "HC head candidate -- Read Committed mark and value; submit as RC"
        AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1013}'`
        inAddr=$inAddrRC
        inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
        inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
        echo $inMark
        echo $inVal
    fi

    tryBuy=`expr $tryBuy + 1`
    echo $tryBuy
    if [ "$inMark" == "$getFuncArg" ]
    then
        # send the offer using the linearized buy
        brcFuncInput=$brcFuncSig$inAddr$inMark$inVal
        echo $brcFuncInput | cut -c 3-
        curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$brcFuncInput'"}],"id":1021}' 
    else  
        # send the offer using serialized buy
        buyFuncInput=$buyFuncSig$inAddr$inMark$inVal
        echo $buyFuncInput | cut -c 3-
        curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$buyFuncInput'"}],"id":1022}' 
    fi
    sleep 0.5
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

