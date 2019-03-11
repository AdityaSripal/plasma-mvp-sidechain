package types

import (
	// "fmt"
	// utils "github.com/AdityaSripal/plasma-mvp-sidechain/utils"
	// utxo "github.com/AdityaSripal/plasma-mvp-sidechain/x/utxo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/ethereum/go-ethereum/common"
	rlp "github.com/ethereum/go-ethereum/rlp"
)

// var _ utxo.SpendMsg = SpendMsg{}

/*
 * SpendMsg contains all the data that a user submits to the Plasma state in order to spend his/her UTXOs
 * No authentication fields are enclosed within
 * In the FourthState implementation, a SpendMsg can spend a maximum of 2 UTXOs and create a maximum of 2 UTXOs
 * The first input must always be filled in.
 * Each UTXO is specified by its position (BlockNum, TxIndex, OIndex, DepositNum). Where BlockNum is the block at which
 * UTXO was created. TxIndex is the index inside the block's transaction list where UTXO was created. OIndex is
 * either 0 or 1 to denote whether the UTXO specified was the first or second output of the transaction.
 * In regular transaction UTXOs, the first 2 fields are filled in and DepositNum is 0.
 * If the UTXO is a deposit UTXO, then first 3 fields are 0 and DepositNum corresponds to the deposit nonce on rootchain
 * SpendMsg implements sdk.Msg and utxo.SpendMsg
 */
type SpendMsg struct {
	// TODO: fill in fields
}

// Implements sdk.Msg. Improve later
func (msg SpendMsg) Type() string { return "spend_utxo" }

// Implements sdk.Msg.
func (msg SpendMsg) Route() string { return "spend" }

/*
 * Implements sdk.Msg
 * Performs Basic stateless validation of our message
 * What would constitute an invalid (or malformed) message?
 */
func (msg SpendMsg) ValidateBasic() sdk.Error {
	// TODO: Implement ValidateBasic to reject malformed messages
	return nil
}

// Implements Msg.
func (msg SpendMsg) GetSignBytes() []byte {
	b, err := rlp.EncodeToBytes(msg)
	if err != nil {
		panic(err)
	}
	return b
}

/*
 * Implements sdk.Msg
 * Who should sign and authenticate this SpendMsg?
 */
func (msg SpendMsg) GetSigners() []sdk.AccAddress {
	// TODO: Implement GetSigners to return the addresses that need to sign this msg.
	return nil
}

/*
func (msg SpendMsg) Inputs() []utxo.Input {
	inputs := []utxo.Input{utxo.Input{
		Owner:    msg.Owner0.Bytes(),
		Position: NewPlasmaPosition(msg.Blknum0, msg.Txindex0, msg.Oindex0, msg.DepositNum0),
	}}
	if NewPlasmaPosition(msg.Blknum1, msg.Txindex1, msg.Oindex1, msg.DepositNum1).IsValid() {
		// Add valid second input
		inputs = append(inputs, utxo.Input{
			Owner:    msg.Owner1.Bytes(),
			Position: NewPlasmaPosition(msg.Blknum1, msg.Txindex1, msg.Oindex1, msg.DepositNum1),
		})
	}
	return inputs
}

func (msg SpendMsg) Outputs() []utxo.Output {
	outputs := []utxo.Output{utxo.Output{msg.Newowner0.Bytes(), Denom, msg.Amount0}}
	if msg.Amount1 != 0 {
		outputs = append(outputs, utxo.Output{msg.Newowner1.Bytes(), Denom, msg.Amount1})
	}
	return outputs
}
*/

//----------------------------------------
// BaseTx
var _ sdk.Tx = BaseTx{}

// What additional fields are necessary to authenticate a SpendMsg
type BaseTx struct {
	SpendMsg
	// TODO: Add additional authentication fields
}

// TODO: Create tx constructor
func NewBaseTx() BaseTx {
	return BaseTx{}
}

// Implements sdk.Tx. Since BaseTx has only one message we return in in array
func (tx BaseTx) GetMsgs() []sdk.Msg { return []sdk.Msg{tx.SpendMsg} }

// TODO: Implement GetSignatures
func (tx BaseTx) GetSignatures() {}
