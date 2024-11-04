require('@nomicfoundation/hardhat-toolbox')

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [{ version: '0.8.18' }]
  },
  networks: {
    evmos: {
      url: 'http://127.0.0.1:8545',
      chainId: 9002,
      accounts: [
        '0x88CBEAD91AEE890D27BF06E003ADE3D4E952427E88F88D31D61D3EF5E5D54305',
        '0x3B7955D25189C99A7468192FCBC6429205C158834053EBE3F78F4512AB432DB9'
      ]
    }
  }
}
