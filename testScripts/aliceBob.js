eth.defaultAccount=alice
personal.unlockAccount(alice, "foobar123", 10000)

loadScript('../sereth/testScripts/Sereth.sol.js')
var serethContract = eth.contract(sereth_ABI)
var si

function csi(address)
{
	si = serethContract.at(address)
}
