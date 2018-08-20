// Start Alice and warm up wiht some ether transactions
//
//miner.start()
eth.defaultAccount=alice
personal.unlockAccount(alice, "foobar123", 10000);
eth.sendTransaction({from:alice, to:lily, value: web3.toWei(0.020, "ether")})
eth.sendTransaction({from:alice, to:lily, value: web3.toWei(0.019, "ether")})
eth.sendTransaction({from:alice, to:bob, value: web3.toWei(0.018, "ether")})
eth.sendTransaction({from:alice, to:bob, value: web3.toWei(0.017, "ether")})
