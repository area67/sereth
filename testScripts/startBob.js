// Start Bob and warm up wiht some ether transactions
//
//miner.start()
eth.defaultAccount=bob
personal.unlockAccount(bob, "foobar123", 10000);
eth.sendTransaction({from:bob, to:lily, value: web3.toWei(0.01, "ether")})
eth.sendTransaction({from:bob, to:lily, value: web3.toWei(0.009, "ether")})
eth.sendTransaction({from:bob, to:alice, value: web3.toWei(0.008, "ether")})
eth.sendTransaction({from:bob, to:alice, value: web3.toWei(0.007, "ether")})
