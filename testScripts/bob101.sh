#!/bin/bash

contractAddress=$1 # 0x199b59896f84c9c2b2d894baf2ca2c203a19c343
fromAddress=0x8691bf25ce4a56b15c1f99c944dc948269031801
setFuncSig=0x19608715

markFuncArg=0000000000000000000000000000000000000000000000000000000000000002
markFuncSig=0xdcfef6fb
markFuncInput=$markFuncSig$markFuncArg$markFuncArg$markFuncArg
#echo $markFuncInput

getFuncArg=0000000000000000000000000000000000000000000000000000000000000001
getFuncSig=0x07173de5
getFuncInput=$getFuncSig$getFuncArg$getFuncArg$getFuncArg
#echo $getFuncInput

# Send the first transaction manually to test connection
echo "TX 0"
read -rsp $'Enter a 4 digit hex value (like A2B3): \n' -n4 num
inAddr=7261614164647265737300000000000000000000000000000000000000000000
inMark=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$markFuncInput'"}, "latest"], "id":1011}' | jq -r ".result" | cut -c 67-130`
inVal=000000000000000000000000000000000000000000000000000000000000
inVal=$inVal$num
raaAddr=7261614164647265737300000000000000000000000000000000000000000000
raaMark=7261614d61726b00000000000000000000000000000000000000000000000000
raaVal=72616156616c7565000000000000000000000000000000000000000000000000
setFuncInput=$setFuncSig$inAddr$inMark$inVal$raaAddr$raaMark$raaVal
echo $setFuncInput | cut -c 3-
read -rsp $'Hit any key to send TX 0\n' -n1 key
curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$setFuncInput'"}],"id":1010}'

# Loop to send many transactions
read -rsp $'Hit any key to start TX loop! \n' -n1 key
for i in {1..100}
do
	echo "TX" $i
#	read -rsp $'Enter a 4 digit hex value (like A1B2): \n' -n4 num  
#        num=222$i
        num=`hexdump -n 2 -e '1/2 "%04X"' /dev/random`
#	curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$markFuncInput'"}, "latest"], "id":1011}'
#	curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}'
	# inMark=c92986bccfcefe626b43484c74d1a8265e80647a8c438a51e2ace689ab58f98f
        read -rsp $'Hit any key to prepare inputs \n' -n1 key
	inMark=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$markFuncInput'"}, "latest"], "id":1011}' | jq -r ".result" | cut -c 67-130`
#	echo $inMark 
#	inMark=`curl -s http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$getFuncInput'"}, "latest"], "id":1011}' | jq -r ".result" | cut -c 3-`
#	echo $inMark 
	inVal=000000000000000000000000000000000000000000000000000000000000
	inVal=$inVal$num
	setFuncInput=$setFuncSig$inAddr$inMark$inVal$raaAddr$raaMark$raaVal
	echo $setFuncInput | cut -c 3-
	read -rsp $'Hit any key to send TX \n' -n1 key
	curl http://localhost:8102 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"'$fromAddress'","to":"'$contractAddress'","data":"'$setFuncInput'"}],"id":1020}'
        sleep 0.25
done

