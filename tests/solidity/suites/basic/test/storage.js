/* eslint-disable no-undef */

const Storage = artifacts.require('Storage')

contract('Test Storage Contract', async function (accounts) {
  let storageInstance

  it('should deploy Storage contract', async function () {
    storageInstance = await Storage.new()
    /* eslint-disable no-unused-expressions */
    expect(storageInstance.address).not.to.be.undefined
  })

  it('should succesfully store a value', async function () {
    const tx = await storageInstance.store(888)
    /* eslint-disable no-unused-expressions */
    expect(tx.tx).not.to.be.undefined
  })

  it('should succesfully retrieve a value', async function () {
    const value = await storageInstance.retrieve()
    expect(value.toString()).to.equal('888')
  })
})
