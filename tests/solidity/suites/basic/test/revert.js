/* eslint-disable no-undef */

const Storage = artifacts.require('Storage')

async function expectRevert (promise) {
  try {
    await promise
  } catch (error) {
    if (error.message.indexOf('revert') === -1) {
      expect('revert').to.equal(
        error.message,
        'Wrong kind of exception received'
      )
    }
    return
  }
  expect.fail('Expected an exception but none was received')
}

contract('Test EVM Revert', async function (accounts) {
  let storageInstance
  it('should deploy Storage contract', async function () {
    storageInstance = await Storage.new()
    /* eslint-disable no-unused-expressions */
    expect(storageInstance.address).not.to.be.undefined
  })

  it('should revert when call `shouldRevert()`', async function () {
    await expectRevert(storageInstance.shouldRevert())
  })
})
