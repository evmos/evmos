const StakingI = artifacts.require("StakingI");

contract("StakingTest", (accounts) => {
  it("should stake to a validator", async () => {
    const valAddr = "evmosvaloper1y036du8szp07wqfep2mzyneygdrvykxv33yesl";
    const stakeAmount = 1000000000000000;

    const staking = await StakingI.at("0x0000000000000000000000000000000000000800");

    const accounts = await web3.eth.getAccounts();
    const tx = await staking.delegate(accounts[0], valAddr, stakeAmount, { from: signer });
    const receipt = await web3.eth.getTransactionReceipt(tx.tx);

    console.log(`Staked 1000000000000000 aevmos with ${valAddr}`);
    console.log("The transaction details are");
    console.log(receipt);
  });
});