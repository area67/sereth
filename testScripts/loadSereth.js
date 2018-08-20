// JavaScript to automate contract deployment
// Sereth copyright 2018 
// V. Cook, Z. Painter and D. Dechev

console.log('Start miner to write new version of contract to chain');
miner.start()
console.log('Load and unlock');
loadScript('solidity/Sereth.sol.js');
eth.defaultAccount=bob
personal.unlockAccount(bob, "foobar123", 10000);
console.log('Deploy contract');
var serethDeploy = { from:bob data:sereth_BIN, gas:800000 };
var serethContract = eth.contract(sereth_ABI);
var serethGas = eth.estimateGas(serethDeploy);
var serethInstance = serethContract.new(serethDeploy);
si = serethInstance;
//console.log(Date.now()/100);
console.log('Block '+eth.blockNumber);
console.log('Please wait for contract deploy transaction to be mined');
console.log('Block '+eth.blockNumber);
personal.unlockAccount(bob, "foobar123", 10000);
console.log('Block '+eth.blockNumber);
personal.unlockAccount(bob, "foobar123", 10000);
console.log('Test function call to match result and see if load is ok (also necessary to setup txpool address)');
console.log('["0x0000000000000000000000008691bf25ce4a56b15c1f99c944dc948269031801", "0x71ca0f9b204c6ee53c8072aa1cb65d88f2afbdb1ce10bfa96d8107f119cc863b", "0x000000000000000000000000000000000000000000000000000000000000003c"]');
console.log('si.set.call([0,"0x7374617274484d53000000000000000000000000000000000000000000000000","0x000000000000000000000000000000000000000000000000000000000000003C"], ["address","0x7374617274484d53000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000037"])');

