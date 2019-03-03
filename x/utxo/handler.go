package utxo

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Return the next position for handler to store newly created UTXOs
// Secondary is true if NextPosition is meant to return secondary output positions for a single multioutput transaction
// If false, NextPosition will increment position to accomadate outputs for a new transaction
type NextPosition func(ctx sdk.Context, secondary bool) Position

// User-defined fee update function
type FeeUpdater func([]Output) sdk.Error

// Handler handles spends of arbitrary utxo implementation
func NewSpendHandler(um Mapper, nextPos NextPosition) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		spendMsg, ok := msg.(SpendMsg)
		if !ok {
			panic("Msg does not implement SpendMsg")
		}

		// Add outputs from store
		for i, o := range spendMsg.Outputs() {
			var next Position
			if i == 0 {
				next = nextPos(ctx, false)
			} else {
				next = nextPos(ctx, true)
			}
			utxo := NewUTXO(o.Owner, o.Amount, o.Denom, next)
			um.ReceiveUTXO(ctx, utxo)
		}

		// Spend inputs from store
		for _, i := range spendMsg.Inputs() {
			err := um.SpendUTXO(ctx, i.Owner, i.Position)
			if err != nil {
				return err.Result()
			}
		}

		return sdk.Result{}
	}
}
