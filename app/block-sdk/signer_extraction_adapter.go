// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package blocksdk

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	signer_adapter "github.com/skip-mev/block-sdk/adapters/signer_extraction_adapter"
)

const (
	// ethereumSignatureExtensionOption is the expected extension option for ethereum msgs signed with via EIP-712
	ethereumSignatureExtensionOption = "/ethermint.evm.v1.ExtensionOptionsEthereumTx"
)

var _ signer_adapter.Adapter = (*SignerExtractionAdapter)(nil)

// SignerExtractionAdapter is an adapter for extracting signers from txs. The SignerExtractor is responsible for extracting the signers for all
// txs (i.e cosmos, ethereum, EIP-712 signed cosmos-txs).
type SignerExtractionAdapter struct {
	signer_adapter.DefaultAdapter
}

func NewSignerExtractorAdapter() SignerExtractionAdapter {
	return SignerExtractionAdapter{}
}

func (sea SignerExtractionAdapter) GetSigners(tx sdk.Tx) ([]signer_adapter.SignerData, error) {
	// attempt to get the signers from the signature as a normal cosmos-sdk signed tx (handle as normal first)
	signers, err := sea.DefaultAdapter.GetSigners(tx)
	if err != nil {
		return nil, err
	}

	// if there are no signers, then check that the tx is an ethereum tx (via extension options)
	if len(signers) == 0 && checkEthereumExtensionOptions(tx) {
		signers = make([]signer_adapter.SignerData, 0)

		// get the signers from each ethereum msg in the tx
		for _, msg := range tx.GetMsgs() {
			// get the signer
			cosmosAddr, nonce, err := getEthereumSignerFromSDKMsg(msg)
			if err != nil {
				return nil, err
			}

			signers = append(signers, signer_adapter.SignerData{
				Signer:   cosmosAddr,
				Sequence: nonce,
			})
		}
	}

	return signers, nil
}
