// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evidence

import (
	"fmt"
	"time"

	evidencetypes "cosmossdk.io/x/evidence/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

const (
	// SubmitEvidenceMethod defines the ABI method name for the evidence SubmitEvidence
	// transaction.
	SubmitEvidenceMethod = "submitEvidence"
	// EvidenceMethod defines the ABI method name for the Evidence query.
	EvidenceMethod = "evidence"
	// GetAllEvidenceMethod defines the ABI method name for the GetAllEvidence query.
	GetAllEvidenceMethod = "getAllEvidence"
)

// EventSubmitEvidence defines the event data for the SubmitEvidence transaction.
type EventSubmitEvidence struct {
	Submitter common.Address
	Hash      []byte
}

// SingleEvidenceOutput defines the output for the Evidence query.
type SingleEvidenceOutput struct {
	Evidence EquivocationData
}

// AllEvidenceOutput defines the output for the GetAllEvidence query.
type AllEvidenceOutput struct {
	Evidence     []EquivocationData
	PageResponse query.PageResponse
}

// EquivocationData represents the Solidity Equivocation struct
type EquivocationData struct {
	Height           int64  `abi:"height"`
	Time             uint64 `abi:"time"`
	Power            int64  `abi:"power"`
	ConsensusAddress string `abi:"consensusAddress"`
}

// ToEquivocation converts the EquivocationData to a types.Equivocation
func (e EquivocationData) ToEquivocation() *evidencetypes.Equivocation {
	return &evidencetypes.Equivocation{
		Height:           e.Height,
		Time:             time.Unix(int64(e.Time), 0).UTC(), //nolint:gosec // G115
		Power:            e.Power,
		ConsensusAddress: e.ConsensusAddress,
	}
}

// NewMsgSubmitEvidence creates a new MsgSubmitEvidence instance.
func NewMsgSubmitEvidence(origin common.Address, args []interface{}) (*evidencetypes.MsgSubmitEvidence, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	equivocation, ok := args[0].(EquivocationData)
	if !ok {
		return nil, fmt.Errorf("invalid equivocation evidence")
	}

	// Convert the EquivocationData to a types.Equivocation
	evidence := equivocation.ToEquivocation()

	// Create the MsgSubmitEvidence using the SDK msg builder
	msg, err := evidencetypes.NewMsgSubmitEvidence(
		sdk.AccAddress(origin.Bytes()),
		evidence,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create evidence message: %w", err)
	}

	return msg, nil
}
