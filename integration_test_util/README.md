
### Integration Test Suite utility
was created with aims to:

- Provide a simple way to set up and run integration tests
- Able to run integration tests in parallel
- Environment nearest to E2E testing
- Initialize a chain with CometBFT node to fully functional testing
- Init chain with pre-defined set of validators and wallets, easier to trace and debug
- Able to set up test for Json-RPC server

Notes:

- To get historical data correctly, need to use query clients/rpc backend/... at corresponding height

Weak points:

- Only support Linux & MacOS, possible to make it compatible with Windows by enhance the TemporaryHolder functionality
- Easy to get import circle error, need a dedicated folder for integration test