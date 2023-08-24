/* eslint-disable no-undef */
/* eslint-disable no-unused-expressions */

const TodoList = artifacts.require('./OpCodes.sol')
let contractInstance

contract('OpCodes', () => {
  beforeEach(async () => {
    contractInstance = await TodoList.deployed()
  })
  it('Should run the majority of opcodes without errors', async () => {
    let error
    try {
      await contractInstance.test()
      await contractInstance.test_stop()
    } catch (err) {
      error = err
    }
    expect(error).to.be.undefined
  })

  it('Should throw invalid op code', async () => {
    let error
    try {
      await contractInstance.test_invalid()
    } catch (err) {
      error = err
    }
    expect(error).not.to.be.undefined
  })

  it('Should revert', async () => {
    let error
    try {
      await contractInstance.test_revert()
    } catch (err) {
      error = err
    }
    expect(error).not.to.be.undefined
  })
})
