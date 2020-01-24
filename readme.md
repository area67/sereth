To Execute:
	From src directory:
	$ go run *.go <number_of_threads>

Dependecies:
	golang.org/x/crypto/sha3 ($go get golang.org/x/crypto/sha3)

Parameters:
	<number_of_threads> - The number of threads to run the benchmark with

Be sure to configure your GOPATH environment variable correctly as per: https://golang.org/doc/code.html. 
Some output data will be written to file, if you want to plot it, execute: $gnuplot plot (requires gnuplot)
