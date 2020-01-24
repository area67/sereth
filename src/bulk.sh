#!/bin/bash

for ((k=1; k<32; k++))
do
    echo $k
    for i in {1..10}
    do
        go run *.go $k $LF
    done
done