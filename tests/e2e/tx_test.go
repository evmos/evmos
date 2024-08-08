// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package e2e

import (
	"context"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v19/tests/e2e/upgrade"
	"github.com/evmos/evmos/v19/utils"
)

// executeTransactions executes some sample transactions to check they are still working after the upgrade.
func (s *IntegrationTestSuite) executeTransactions() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chainID := utils.TestnetChainID + "-1"

	// TODO: Add more transactions in future (e.g. staking precompile)
	s.sendBankTransfer(ctx, chainID)
}

// SendBankTransfer sends a bank transfer to check that the transactions are still working after the upgrade.
func (s *IntegrationTestSuite) sendBankTransfer(ctx context.Context, chainID string) {
	receiver := "evmos1jcltmuhplrdcwp7stlr4hlhlhgd4htqh3a79sq"
	sentCoins := sdk.Coins{sdk.NewInt64Coin("aevmos", 10000000000)}

	balancePre, err := s.upgradeManager.GetBalance(ctx, s.upgradeParams.ChainID, receiver)
	s.Require().NoError(err, "can't get balance of receiver account")

	// send some tokens between accounts to check transactions are still working
	exec, err := s.upgradeManager.CreateModuleTxExec(upgrade.E2ETxArgs{
		ModuleName: "bank",
		SubCommand: "send",
		Args:       []string{"mykey", receiver, sentCoins.String()},
		ChainID:    chainID,
		From:       "mykey",
	})
	s.Require().NoError(err, "failed to create bank send tx command")

	_, errBuf, err := s.upgradeManager.RunExec(ctx, exec)
	s.Require().NoError(err, "failed to execute bank send tx")
	// NOTE: The only message in the errBuf that is allowed is `gas estimate: ...`
	gasEstimateMatch, err := regexp.MatchString(`^\s*gas estimate: \d+\s*$`, errBuf.String())
	s.Require().NoError(err, "failed to match gas estimate message")
	s.Require().True(
		gasEstimateMatch,
		"expected message in errBuf to be `gas estimate: ...`; got: %q\n",
		errBuf.String(),
	)

	// Wait until the transaction has succeeded and is included in the chain
	err = s.upgradeManager.WaitNBlocks(ctx, 2)
	s.Require().NoError(err, "failed to wait for blocks")

	balancePost, err := s.upgradeManager.GetBalance(ctx, s.upgradeParams.ChainID, receiver)
	s.Require().NoError(err, "can't get balance of receiver account")

	diff := balancePost.Sub(balancePre...)
	s.Require().Equal(diff.String(), sentCoins.String(), "unexpected difference in bank balance")
}
