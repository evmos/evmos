package app

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ccvgov "github.com/cosmos/interchain-security/v4/x/ccv/democracy/governance"
)

// TODO: Whitelist more Evmos specific proposals here ?
var WhiteListModule = map[string]struct{}{
	"/cosmos.gov.v1.MsgUpdateParams":               {},
	"/cosmos.bank.v1beta1.MsgUpdateParams":         {},
	"/cosmos.staking.v1beta1.MsgUpdateParams":      {},
	"/cosmos.distribution.v1beta1.MsgUpdateParams": {},
	"/cosmos.mint.v1beta1.MsgUpdateParams":         {},
	"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade":   {},
	"/cosmos.upgrade.v1beta1.MsgCancelUpgrade":     {},
}

func IsModuleWhiteList(typeUrl string) bool {
	_, found := WhiteListModule[typeUrl]
	return found
}

func IsProposalWhitelisted(content govv1beta1.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isParamChangeWhitelisted(getParamChangesMapFromArray(c.Changes))
	case *upgradetypes.SoftwareUpgradeProposal, //nolint:staticcheck
		*upgradetypes.CancelSoftwareUpgradeProposal: //nolint:staticcheck
		return true

	default:
		return false
	}
}

func getParamChangesMapFromArray(paramChanges []proposal.ParamChange) map[ccvgov.ParamChangeKey]struct{} {
	res := map[ccvgov.ParamChangeKey]struct{}{}
	for _, paramChange := range paramChanges {
		key := ccvgov.ParamChangeKey{
			MsgType: paramChange.Subspace,
			Key:     paramChange.Key,
		}

		res[key] = struct{}{}
	}

	return res
}

func isParamChangeWhitelisted(paramChanges map[ccvgov.ParamChangeKey]struct{}) bool {
	for paramChangeKey := range paramChanges {
		_, found := WhitelistedParams[paramChangeKey]
		if !found {
			return false
		}
	}
	return true
}

var WhitelistedParams = map[ccvgov.ParamChangeKey]struct{}{
	//bank
	{MsgType: banktypes.ModuleName, Key: string(banktypes.KeySendEnabled)}: {},
	//governance
	{MsgType: govtypes.ModuleName, Key: string(govv1.ParamStoreKeyDepositParams)}: {}, //min_deposit, max_deposit_period
	{MsgType: govtypes.ModuleName, Key: string(govv1.ParamStoreKeyVotingParams)}:  {}, //voting_period
	{MsgType: govtypes.ModuleName, Key: string(govv1.ParamStoreKeyTallyParams)}:   {}, //quorum,threshold,veto_threshold
	//staking
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyUnbondingTime)}:     {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxValidators)}:     {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxEntries)}:        {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyHistoricalEntries)}: {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyBondDenom)}:         {},
	//distribution
	{MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyCommunityTax)}:        {},
	{MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyWithdrawAddrEnabled)}: {},
	//mint
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyMintDenom)}: {},
	//ibc transfer
	{MsgType: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeySendEnabled)}:    {},
	{MsgType: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeyReceiveEnabled)}: {},
	// TODO: Add more params here that are Evmos specific and need to be whitelisted ?
	//ica
	{MsgType: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyHostEnabled)}:   {},
	{MsgType: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyAllowMessages)}: {},
}
