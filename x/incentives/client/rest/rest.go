package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/tharsis/evmos/v3/x/incentives/types"
)

// RegisterIncentiveProposalRequest defines a request for a new register a
// contract incentive.
type RegisterIncentiveProposalRequest struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title           string       `json:"title" yaml:"title"`
	Description     string       `json:"description" yaml:"description"`
	Deposit         sdk.Coins    `json:"deposit" yaml:"deposit"`
	ContractAddress string       `json:"contract_address" yaml:"contract_address"`
	Allocation      sdk.DecCoins `json:"allocation" yaml:"allocation"`
	Epochs          uint32       `json:"epochs" yaml:"epochs"`
}

// CancelIncentiveProposalRequest defines a request for a new register a
// contract incentive.
type CancelIncentiveProposalRequest struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title           string       `json:"title" yaml:"title"`
	Description     string       `json:"description" yaml:"description"`
	Deposit         sdk.Coins    `json:"deposit" yaml:"deposit"`
	ContractAddress string       `json:"contract_address" yaml:"contract_address"`
}

func RegisterIncentiveProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ModuleName,
		Handler:  newRegisterIncentiveProposalHandler(clientCtx),
	}
}

func CancelIncentiveProposalRequestRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ModuleName,
		Handler:  newCancelIncentiveProposalRequestHandler(clientCtx),
	}
}

func newRegisterIncentiveProposalHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterIncentiveProposalRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		contract := req.ContractAddress

		content := types.NewRegisterIncentiveProposal(req.Title, req.Description, contract, req.Allocation, req.Epochs)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

func newCancelIncentiveProposalRequestHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CancelIncentiveProposalRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		contract := req.ContractAddress

		content := types.NewCancelIncentiveProposal(req.Title, req.Description, contract)
		msg, err := govtypes.NewMsgSubmitProposal(content, req.Deposit, fromAddr)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
