#!/bin/bash

# blockchain address of sender and contract
contractAddress=$1
fromAddress=0x8691bf25ce4a56b15c1f99c944dc948269031801

bobAddress=0x8691bf25ce4a56b15c1f99c944dc948269031801
aliceAddress=0xdda6ef2ff259928c561b2d30f0cad2c2736ce8b6
lily=0xb1b6a66a410edc72473d92decb3772bad863e243

#Ratio of sets to buy
ratio=5

# save the special values
raaAddr=7261614164647265737300000000000000000000000000000000000000000000
raaMark=7261614d61726b00000000000000000000000000000000000000000000000000
raaVal=72616156616c7565000000000000000000000000000000000000000000000000

# function signatures for input
setFuncSig=0xd1602737
srcFuncSig=0xdde5aece
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

# Loop and change the price from time to time
#for orderTarget in {2..100..2}
echo -e "\nLoop to send price changes from Peer 2 (Bob)"
#while true
retryInterval=0.1
txInterval=1.0

for j in {1..5}
do
    ratio=$j
    echo "Beginning buy:set ratio 1:$j" >> results.dat
    for k in {1..2}
    do
        # set the initial counters and timers
        echo -e "\nInitialize counters and timers"
        qtyJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$qtyFuncInput'"}, "latest"], "id":1031}'`
        echo $qtyJSON
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
        poolCond_IS=0
        poolCond_MB=0
        poolCond_HC=0

        for i in {0..15}
        do
	    #Noise
	    #curl 172.31.13.77:8103 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$lily'","to":"'$bobAddress'","value":"0x100000"}], "id":1011}'
            #curl 172.31.13.77:8103 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$lily'","to":"'$bobAddress'","value":"0x100000"}], "id":1011}'
            
	    #---------------------------------------------------------------------
            #Send set transaction from BOB
            if ! (( i % $ratio ));
            then
                # set the new value
                num=`hexdump -n 1 -e '1/2 "%02X"' /dev/random`
                inVal=00000000000000000000000000000000000000000000000000000000000001
                inVal=$inVal$num
                # get the latest mark
                AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
                while [[ ${#AMV} -lt 50 ]]
                do
                    sleep $retryInterval
                    AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
                #echo $AMV
                done
                inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
                # fill in the address, and possibly replace the mark, based on the txpool state
                if [[ "$inMark" != "$raaMark" && "$inMark" != "$getFuncArg" ]]
                then
                    inAddr=$inAddrRU
                    poolCond=IS
                    (( poolCond_IS++ ))
                else
                    AMV=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1013}'`
                    inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
                    inAddr=$inAddrRC
                    poolCond=HC
                    (( poolCond_HC++ ))
                fi
                read -n 1 -t 0.1 key
                if [[ $key = "q" ]]
                then
                   echo
                   break
                fi
                # send the transaction
                trySet=`expr $trySet + 1`
                echo -e "\nManager trySet $trySet Time `date +%s` $poolCond"
                setFuncInput=$setFuncSig$inAddr$inMark$inVal
                #srcFuncInput=$srcFuncSig$inAddr$inMark$inVal
                #echo $setFuncInput | cut -c 3-
                #curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$setFuncInput'", "gas": "0x138800", "gasPrice": "0x100"}],"id":1010}'
                curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$setFuncInput'"}],"id":1015}'

		# echo 'curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data' '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$setFuncInput'"}],"id":1015}'

		fc -ln -1
                #curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$srcFuncInput'", "gas": "0x138800", "gasPrice": "0x100"}],"id":1017}'
                sleep $txInterval
		echo
	    fi
            #End of Bob
            #---------------------------------------------------------------------
            #Send buy transaction from ALICE

            AMV=`curl -s http://localhost:8101 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$aliceAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`

            while [[ ${#AMV} -lt 50 ]]
            do
                sleep $retryInterval
                AMV=`curl -s http://localhost:8101 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$aliceAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'`
            done

            inAddr=$inAddrRU
            inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
            #echo $inMark
            inMarkAMV=$inMark
            inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
            #echo $inVal

            if [[ "$inMark" != "$raaMark" && "$inMark" != "$getFuncArg" ]]
            then
                if [ "$inVal" != "$raaVal" ]
                then # both inMark and inVal are Read Uncommitted values from the algorithm
                    poolCond=IS
                    (( poolCond_IS++ ))
                    #echo "IS intra-block transaction with intra block set -- keep Read Uncommitted mark and value; submit as RU"
                    #echo $inMark
                    #echo $inVal
                else # mark does not match special but the value does, this is a multiple buy case on the committed value, Read Committed value but keep the RU mark
                    poolCond=MB
                    (( poolCond_MB++ ))
                    echo "MB multiple buys with value set previous to this block -- Read committed value but keep dynamic mark; submit as RU"
                    AMV=`curl -s http://localhost:8101 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$aliceAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1012}'`
                    inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
                fi
            else # mark indicates nothing has happened since block was published, read committed mark and value
            poolCond=HC
                (( poolCond_HC++ ))
                #echo "HC head candidate -- Read Committed mark and value; submit as RC"
                AMV=`curl -s http://localhost:8101 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$aliceAddress'","to":"'$contractAddress'","data":"'$grcFuncInput'"}, "latest"], "id":1013}'`
                inAddr=$inAddrRC
                inMark=`echo $AMV | jq -r ".result" | cut -c 67-130`
                inVal=`echo $AMV | jq -r ".result" | cut -c 131-194`
                #echo $inMark
                #echo $inVal
            fi

            tryBuy=`expr $tryBuy + 1`
            #echo $tryBuy
            #if [ "$inMarkAMV" == "$getFuncArg" ]
            if  [ false ]
            then
                # send the offer using the linearized buy
                #echo "Linearized buy"
                buyMode=L
                brcFuncInput=$brcFuncSig$inAddr$inMark$inVal
                echo -e "Alice TX $i Time `date +%s` $poolCond $buyMode"
                echo $brcFuncInput
                curl http://localhost:8101 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$aliceAddress'","to":"'$contractAddress'","data":"'$brcFuncInput'", "gas": "0x138800"}],"id":1021}'

            else
                # send the offer using serialized buy
                buyMode=S
                buyFuncInput=$buyFuncSig$inAddr$inMark$inVal
                echo -e "TX $i Time `date +%s` $poolCond $buyMode"
                #echo $buyFuncInput | cut -c 3-
                curl http://localhost:8101 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$aliceAddress'","to":"'$contractAddress'","data":"'$buyFuncInput'"}],"id":1022}'
            fi
            #End of Alice
            #---------------------------------------------------------------------
            sleep $txInterval
        done

        # set the final counters and timers
        echo -e "\nBob -- final counters and timers -- waiting for block updates"


        #Wait for txpool to be empty
        # pending=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"txpool_content","params":[4], "id":1011}' | python -c "import json,sys;obj=json.load(sys.stdin);print len(obj['result']['pending']);"`
        # queued=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"txpool_content","params":[4], "id":1011}' | python -c "import json,sys;obj=json.load(sys.stdin);print len(obj['result']['queued']);"`

        # while (($pending > 0)) || (($queued > 0))
        # do
        #         sleep $txInterval
        #         queued=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"txpool_content","params":[4], "id":1011}' | python -c "import json,sys;obj=json.load(sys.stdin);print len(obj['result']['queued']);"`
        #         pending=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"txpool_content","params":[4], "id":1011}' | python -c "import json,sys;obj=json.load(sys.stdin);print len(obj['result']['pending']);"`
        # done

        #echo "Get nTX"
        qtyJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$bobAddress'","to":"'$contractAddress'","data":"'$qtyFuncInput'"}, "latest"], "id":1031}'`
	echo $qtyFuncInput
	echo "here"
        echo $qtyJSON
        endBuy=0x`echo $qtyJSON | jq -r ".result" | cut -c 3-66`
        endBuy=$(($endBuy))
        echo $endBuy
        endSet=0x`echo $qtyJSON | jq -r ".result" | cut -c 67-130`
        endSet=$(($endSet))
        echo $endSet
        echo "Get block"
        blockJSON=`curl -s http://localhost:8102 -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber", "params":["pending", true], "id":1}'`
        endBlock=`echo $blockJSON | jq -r ".result" | cut -d "\"" -f 2`
        endBlock=$(($endBlock))
        echo $endBlock
        endTime=`date +%s`

        # show stats from this run
        elapsedTime=`expr $endTime - $startTime`
        elapsedBlocks=`expr $endBlock - $startBlock`
        sets=`expr $endSet - $startSet`
        buys=`expr $endBuy - $startBuy`
        echo >> results.dat
	echo "Run #$k" >> results.dat
        echo "Blocks" >> results.dat
        echo "$startBlock $endBlock $elapsedBlocks blocks" >> results.dat
        echo "Buys" >> results.dat
        echo "$startBuy $endBuy $buys of $tryBuy" >> results.dat
        echo "Sets" >> results.dat
        echo "$startSet $endSet $sets of $trySet IS $poolCond_IS MB $poolCond_MB HC $poolCond_HC" >> results.dat
        echo "" >> results.dat
	
        echo "txInterval retryInterval ratio  k  startTime  endTime    elapsedTime   startBlock  endBlock  elapsedBlocks  startSet  endSet  sets trySet poolCond_IS  poolCond_MB  poolCond_HC  startBuy  endBuy  buys tryBuy" >> results.dat
        echo "$txInterval        $retryInterval           $ratio      $k  $startTime $endTime $elapsedTime           $startBlock        $endBlock      $elapsedBlocks              $startSet      $endSet    $sets   $trySet     $poolCond_IS           $poolCond_MB            $poolCond_HC            $startBuy     $endBuy   $buys  $tryBuy" >> results.dat
    done
done


