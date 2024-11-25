// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evidence

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/accounts/abi"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
)

// Evidence implements the query logic for getting evidence by hash.
func (p *Precompile) Evidence(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	evidenceHash, ok := args[0].([]byte)
	if !ok {
		return nil, errors.New(ErrInvalidEvidenceHash)
	}

	queryServer := evidencekeeper.NewQuerier(&p.evidenceKeeper)
	res, err := queryServer.Evidence(ctx, &evidencetypes.QueryEvidenceRequest{
		Hash: strings.ToUpper(hex.EncodeToString(evidenceHash)),
	})
	if err != nil {
		return nil, err
	}

	// Convert the Any type to Equivocation
	equivocation, ok := res.Evidence.GetCachedValue().(*evidencetypes.Equivocation)
	if !ok {
		return nil, errors.New(ErrExpectedEquivocation)
	}

	// Convert to our Equivocation struct
	evidence := EquivocationData{
		Height:           equivocation.Height,
		Time:             uint64(equivocation.Time.Unix()), //nolint:gosec // G115
		Power:            equivocation.Power,
		ConsensusAddress: equivocation.ConsensusAddress,
	}

	return method.Outputs.Pack(evidence)
}

// GetAllEvidence implements the query logic for getting all evidence.
func (p *Precompile) GetAllEvidence(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	pageRequest, ok := args[0].(*query.PageRequest)
	if !ok {
		return nil, fmt.Errorf("invalid page request")
	}

	queryServer := evidencekeeper.NewQuerier(&p.evidenceKeeper)
	res, err := queryServer.AllEvidence(ctx, &evidencetypes.QueryAllEvidenceRequest{
		Pagination: pageRequest,
	})
	if err != nil {
		return nil, err
	}

	evidenceList := make([]EquivocationData, len(res.Evidence))
	for i, evidence := range res.Evidence {
		equivocation, ok := evidence.GetCachedValue().(*evidencetypes.Equivocation)
		if !ok {
			return nil, fmt.Errorf("invalid evidence type at index %d: expected Equivocation", i)
		}

		evidenceList[i] = EquivocationData{
			Height:           equivocation.Height,
			Time:             uint64(equivocation.Time.Unix()), //nolint:gosec // G115
			Power:            equivocation.Power,
			ConsensusAddress: equivocation.ConsensusAddress,
		}
	}

	return method.Outputs.Pack(evidenceList, res.Pagination)
}
