/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [
      {
        version: "0.8.20",
      },
      {
        version: "0.4.22",
      },
    ],
  },
  paths: {
    sources: "./solidity",
  },
};
