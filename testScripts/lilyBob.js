console.log('g')
eth.defaultAccount=lily
personal.unlockAccount(lily, "foobar123", 10000)

console.log('g')
loadScript('Sereth.sol.js')
console.log('f')
console.log(sereth_ABI)
var serethContract = eth.contract(sereth_ABI)
var si
function csi(address)
{
	si = serethContract.at(address)
}
