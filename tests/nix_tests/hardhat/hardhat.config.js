require('@nomicfoundation/hardhat-toolbox')

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [
      {
        // NOTE: changing compiler version may break tests,
        // as the expected gas and bytecodes may be different
        version: '0.8.18'
      }
    ]
  },
  typechain: {
    outDir: 'typechain',
    target: 'ethers-v6'
  }
}
