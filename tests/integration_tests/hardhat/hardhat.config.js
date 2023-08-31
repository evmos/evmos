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
  },
  networks: {
    evmos: {
      url: 'http://127.0.0.1:26701',
      chainId: 9000,
      accounts: [
        '0x82F33180C13B553AF9046E6D353960165ECBE1C1746B5DB38D4343C2FDB0480E',
        '0xE95790AFDE7A9EAF910419BBDFB7EF8ED93A6570562F19ADC4BB73C29E80CC64'
      ]
    }
  }
}
