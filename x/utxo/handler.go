package utxo

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Return the next position for handler to store newly created UTXOs
// Secondary is true if NextPosition is meant to return secondary output positions for a single multioutput transaction
// If false, NextPosition will increment position to accomadate outputs for a new transaction
type NextPosition func(ctx sdk.Context, secondary bool) Position

// Handler handles spends of arbitrary utxo implementation
// The handler will take SpendMsg and apply the appropriate state changes
// No need to have checks here since that is done before in AnteHandler.
func NewSpendHandler(um Mapper, nextPos NextPosition) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		spendMsg, ok := msg.(SpendMsg)
		if !ok {
			panic("Msg does not implement SpendMsg")
		}

		// TODO: Spend all input UTXOs

		// TODO: Create new Unspent Transaction Outputs and save them in store

		return sdk.Result{}
	}
}
