package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/tharsis/evmos/x/intrarelayer/types"
)

// RegisterCoinProposalRequest defines a request for a new register coin proposal.
type RegisterCoinProposalRequest struct {
	BaseReq     rest.BaseReq       `json:"base_req" yaml:"base_req"`
	Title       string             `json:"title" yaml:"title"`
	Description string             `json:"description" yaml:"description"`
	Deposit     sdk.Coins          `json:"deposit" yaml:"deposit"`
	Metadata    banktypes.Metadata `json:"metadata" yaml:"metadata"`
}

// RegisterERC20ProposalRequest defines a request for a new register ERC20 proposal.
type RegisterERC20ProposalRequest struct {
	BaseReq      rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title        string       `json:"title" yaml:"title"`
	Description  string       `json:"description" yaml:"description"`
	Deposit      sdk.Coins    `json:"deposit" yaml:"deposit"`
	ERC20Address string       `json:"erc20_address" yaml:"erc20_address"`
}

// ToggleTokenRelayProposalRequest defines a request for a toggle token relay proposal.
type ToggleTokenRelayProposalRequest struct {
	BaseReq     rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title       string       `json:"title" yaml:"title"`
	Description string       `json:"description" yaml:"description"`
	Deposit     sdk.Coins    `json:"deposit" yaml:"deposit"`
	Token       string       `json:"token" yaml:"token"`
}

// UpdateTokenPairERC20ProposalRequest defines a request for a update token pair ERC20 proposal.
type UpdateTokenPairERC20ProposalRequest struct {
	BaseReq         rest.BaseReq `json:"base_req" yaml:"base_req"`
	Title           string       `json:"title" yaml:"title"`
	Description     string       `json:"description" yaml:"description"`
	Deposit         sdk.Coins    `json:"deposit" yaml:"deposit"`
	ERC20Address    string       `json:"erc20_address" yaml:"erc20_address"`
	NewERC20Address string       `json:"new_erc20_address" yaml:"new_erc20_address"`
}

func RegisterCoinProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ModuleName,
		Handler:  newRegisterCoinProposalHandler(clientCtx),
	}
}

func RegisterERC20ProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ModuleName,
		Handler:  newRegisterERC20ProposalHandler(clientCtx),
	}
}

func ToggleTokenRelayRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ModuleName,
		Handler:  newToggleTokenRelayHandler(clientCtx),
	}
}

func UpdateTokenPairERC20ProposalRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
	return govrest.ProposalRESTHandler{
		SubRoute: types.ModuleName,
		Handler:  newUpdateTokenPairERC20ProposalHandler(clientCtx),
	}
}

// nolint: dupl
func newRegisterCoinProposalHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterCoinProposalRequest

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

		content := types.NewRegisterCoinProposal(req.Title, req.Description, req.Metadata)
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

// nolint: dupl
func newRegisterERC20ProposalHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterERC20ProposalRequest

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

		content := types.NewRegisterERC20Proposal(req.Title, req.Description, req.ERC20Address)
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

// nolint: dupl
func newToggleTokenRelayHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ToggleTokenRelayProposalRequest

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

		content := types.NewToggleTokenRelayProposal(req.Title, req.Description, req.Token)
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

func newUpdateTokenPairERC20ProposalHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UpdateTokenPairERC20ProposalRequest

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

		content := types.NewUpdateTokenPairERC20Proposal(req.Title, req.Description, req.ERC20Address, req.NewERC20Address)
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
