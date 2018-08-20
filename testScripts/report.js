function listBTX(n1, n2) {
  for (var i = n1; i <= n2; i++) {
    var block = eth.getBlock(i, true);
    console.log("\nblock " + i + " hash " + block.hash);
    if (block.transactions != null) {
      block.transactions.forEach( function(e) {
        console.log(e.transactionIndex + " " + e.input.slice(226,235) + " " + e.nonce + " " + e.input.slice(2,10) + " " + e.input.slice(198,202) + " " + e.input.slice(390,394));
      })
  }
}
}

listBTX(0,200);

