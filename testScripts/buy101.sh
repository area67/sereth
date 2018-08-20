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

# initialize actual arguments
getFuncArg=0000000000000000000000000000000000000000000000000000000000022222
inAddrRC=0000000000000000000000000000000000000000000000000000000000077777
inAddrRU=0000000000000000000000000000000000000000000000000000000000088888

# set the RPC data input for get and grc 
getFuncInput=$getFuncSig$getFuncArg$getFuncArg$getFuncArg
#echo $getFuncInput
grcFuncInput=$grcFuncSig$getFuncArg$getFuncArg$getFuncArg
#echo $grcFuncInput

# Loop to send many transactions
for i in {1..100}
do
    echo "TX" $i
    #read -rsp $'Hit any key to prepare offer \n' -n1 key
    inAddr=$inAddrRU
    AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
    inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
    echo $inMark
    inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
    echo $inVal

    if [ "$inMark" != "$raaMark" ]
    then 
        if [ "$inVal" != "$raaVal" ]
        then # both inMark and inVal are Read Uncommitted values from the algorithm
            echo "intra-block transaction with intra block set -- keep Read Uncommitted mark and value; submit as RU"
            echo $inMark
            echo $inVal
        else # mark does not match special but the value does, this is a multiple buy case on the committed value, Read Committed value but keep the RU mark
            echo " multiple buys and value set previous to this block -- Read committed value but keep dynamic mark; submit as RU"
            AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1012}'`
            inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
            echo $inMark
            echo $inVal
        fi
    else # mark indicates nothing has happened since block was published, read committed mark and value
        echo "head candidate -- Read Committed mark and value; submit as RC"
        AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1013}'`
        inAddr=$inAddrRC
        inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
        inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
        echo $inMark
        echo $inVal
    fi

    read -rsp $'Enter "o" to send offer at current price, or "p" to change the price instead... \n' -n1 key
    if [ "$key" == "o" ]
    then 
        # send the offer
        buyFuncInput=$buyFuncSig$inAddr$inMark$inVal
        echo $buyFuncInput | cut -c 3-
        curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$buyFuncInput'"}],"id":1012}'
    fi
    if [ "$key" == "p" ]
    then
        read -rsp $'Enter a 2 digit hex value that will be the new price (like 2A): \n' -n2 num
        # change the price
        inVal=00000000000000000000000000000000000000000000000000000000000000
        inVal=$inVal$num
        setFuncInput=$setFuncSig$inAddr$inMark$inVal
        echo $setFuncInput | cut -c 3-
        curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$setFuncInput'"}],"id":1010}'
    fi
    sleep 0.25
done

