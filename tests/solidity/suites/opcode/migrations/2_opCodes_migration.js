/* eslint-disable no-undef */

const OpCodes = artifacts.require('./OpCodes.sol')

module.exports = function (deployer) {
  deployer.deploy(OpCodes)
}
