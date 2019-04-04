// JavaScript to automate contract deployment
// Sereth copyright 2018 
// V. Cook, Z. Painter and D. Dechev

console.log('Start miner to write new version of contract to chain');
miner.start()
console.log('Load and unlock');
loadScript('../Sereth/testScripts/Sereth.sol.js');
eth.defaultAccount=bob
personal.unlockAccount(bob, "foobar123", 10000);
console.log('Deploy contract');
var serethDeploy = { from:bob data:sereth_BIN, gas:900000 };
var serethContract = eth.contract(sereth_ABI);
var serethGas = eth.estimateGas(serethDeploy);
var serethInstance = serethContract.new(serethDeploy);
si = serethInstance;
//console.log(Date.now()/100);
console.log('Block: '+eth.blockNumber);
console.log('Please wait for contract deploy transaction to be mined');
console.log('Block: '+eth.blockNumber);
personal.unlockAccount(bob, "foobar123", 10000);
console.log('Block: '+eth.blockNumber);
personal.unlockAccount(bob, "foobar123", 10000);
console.log('Test RAA to match result and see if load is ok (also necessary to setup txpool address)');
console.log('Use: si.get([2,2,2])');
console.log('Also: si.address eth.blockNumber');

