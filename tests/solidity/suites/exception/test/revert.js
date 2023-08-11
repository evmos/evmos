/* eslint-disable no-undef */

const TestRevert = artifacts.require('TestRevert')

contract('TestRevert', (accounts) => {
  let revert

  beforeEach(async () => {
    revert = await TestRevert.new()
  })
  it('should revert', async () => {
    await revert.try_set(10)
    let no = await revert.query_a()
    assert.equal(no, '0', 'The modification on a should be reverted')
    no = await revert.query_b()
    assert.equal(no, '10', 'The modification on b should not be reverted')
    no = await revert.query_c()
    assert.equal(no, '10', 'The modification on c should not be reverted')

    await revert.set(10)
    no = await revert.query_a()
    assert.equal(no, '10', 'The force set should not be reverted')
  })
})
