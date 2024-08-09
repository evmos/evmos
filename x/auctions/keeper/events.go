// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

//// EmitRoundFinished emits an event for the RoundFinished event.
//func EmitRoundFinished(ctx sdk.Context, evmKeeper evmkeeper.Keeper, stateDB vm.StateDB, roundID *big.Int) error {
//	// Get the precompile instance
//	evmParams := evmKeeper.GetParams(ctx)
//	precompile, found, err := evmKeeper.GetStaticPrecompileInstance(&evmParams, common.HexToAddress(auctionsprecompile.PrecompileAddress))
//	if err != nil || !found {
//		return err
//	}
//
//	// Prepare the events
//	p := precompile.(auctionsprecompile.Precompile)
//	event := p.Events[auctionsprecompile.EventTypeRoundFinished]
//	topics := make([]common.Hash, 1)
//
//	// The first topic is always the signature of the event.
//	topics[0] = event.ID
//
//	// Pack the arguments to be used as the Data field
//	arguments := abi.Arguments{event.Inputs[0]}
//	packed, err := arguments.Pack(roundID)
//	if err != nil {
//		return err
//	}
//
//	stateDB.AddLog(&ethtypes.Log{
//		Address:     p.Address(),
//		Topics:      topics,
//		Data:        packed,
//		BlockNumber: uint64(ctx.BlockHeight()),
//	})
//
//	return nil
//}
