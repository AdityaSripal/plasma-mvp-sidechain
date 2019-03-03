package utxo

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
)

// UTXO is a standard unspent transaction output
// When spent, it becomes invalid and spender keys are filled in
type UTXO struct {
	Address  []byte
	Amount   uint64
	Denom    string
	Valid    bool
	Position Position
}

func NewUTXO(owner []byte, amount uint64, denom string, position Position) UTXO {
	return UTXO{
		Address:  owner,
		Amount:   amount,
		Denom:    denom,
		Position: position,
		Valid:    true,
	}
}

func (utxo UTXO) StoreKey(cdc *amino.Codec) []byte {
	encPos := cdc.MustMarshalBinaryBare(utxo.Position)
	return append(utxo.Address, encPos...)
}

// Create a prototype Position.
// Must return pointer to struct implementing Position
type ProtoPosition func() Position

// Positions must be unqiue or a collision may result when using mapper.go
type Position interface {
	// Position is a uint slice
	Get() []sdk.Uint // get position int slice. Return nil if unset.

	// returns true if the position is valid, false otherwise
	IsValid() bool
}

// SpendMsg is an interface that wraps sdk.Msg with additional information
// for the UTXO spend handler.
type SpendMsg interface {
	sdk.Msg

	Inputs() []Input
	Outputs() []Output
}

type Input struct {
	Owner []byte
	Position
}

type Output struct {
	Owner  []byte
	Denom  string
	Amount uint64
}
