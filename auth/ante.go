package auth

import (
	"fmt"
	"github.com/AdityaSripal/plasma-mvp-sidechain/eth"
	types "github.com/AdityaSripal/plasma-mvp-sidechain/types"
	utils "github.com/AdityaSripal/plasma-mvp-sidechain/utils"
	"github.com/AdityaSripal/plasma-mvp-sidechain/x/kvstore"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"reflect"

	"github.com/AdityaSripal/plasma-mvp-sidechain/x/utxo"
)

// NewAnteHandler returns an AnteHandler that checks signatures,
// confirm signatures, and increments the feeAmount
func NewAnteHandler(utxoMapper utxo.Mapper, plasmaStore kvstore.KVStore, plasmaClient *eth.Plasma) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (_ sdk.Context, _ sdk.Result, abort bool) {

		baseTx, ok := tx.(types.BaseTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be in form of BaseTx").Result(), true
		}

		sigs := baseTx.GetSignatures()

		// Base Tx must have only one msg
		msg := baseTx.GetMsgs()[0]

		// Assert that number of signatures is correct.
		var signerAddrs = msg.GetSigners()

		spendMsg, ok := msg.(types.SpendMsg)
		if !ok {
			return ctx, sdk.ErrInternal("msg must be of type SpendMsg").Result(), true
		}
		signBytes := spendMsg.GetSignBytes()

		// Verify the first input signature
		addr0 := common.BytesToAddress(signerAddrs[0].Bytes())
		position0 := types.PlasmaPosition{spendMsg.Blknum0, spendMsg.Txindex0, spendMsg.Oindex0, spendMsg.DepositNum0}

		res := checkUTXO(ctx, plasmaClient, utxoMapper, position0, addr0)
		if !res.IsOK() {
			return ctx, res, true
		}
		exitErr := hasTXExited(plasmaClient, position0)
		if exitErr != nil {
			return ctx, exitErr.Result(), true
		}
		if position0.IsDeposit() {
			deposit, _ := DepositExists(position0.DepositNum, plasmaClient)
			inputUTXO := utxo.NewUTXO(deposit.Owner.Bytes(), deposit.Amount.Uint64(), types.Denom, position0)
			utxoMapper.ReceiveUTXO(ctx, inputUTXO)
		}

		res = processSig(addr0, sigs[0], signBytes)
		if !res.IsOK() {
			return ctx, res, true
		}

		// Verify the second input
		if utils.ValidAddress(spendMsg.Owner1) {
			addr1 := common.BytesToAddress(signerAddrs[1].Bytes())
			position1 := types.PlasmaPosition{spendMsg.Blknum1, spendMsg.Txindex1, spendMsg.Oindex1, spendMsg.DepositNum1}

			exitErr := hasTXExited(plasmaClient, position1)
			if exitErr != nil {
				return ctx, exitErr.Result(), true
			}

			// second input can be less than fee amount
			res := checkUTXO(ctx, plasmaClient, utxoMapper, position1, addr1)
			if !res.IsOK() {
				return ctx, res, true
			}
			if position1.IsDeposit() {
				deposit, _ := DepositExists(position1.DepositNum, plasmaClient)
				inputUTXO := utxo.NewUTXO(deposit.Owner.Bytes(), deposit.Amount.Uint64(), types.Denom, position1)
				utxoMapper.ReceiveUTXO(ctx, inputUTXO)
			}

			res = processSig(addr1, sigs[1], signBytes)

			if !res.IsOK() {
				return ctx, res, true
			}
		}

		/* Check that balances add up */
		// Add up all inputs
		totalInput := map[string]uint64{}
		for _, i := range spendMsg.Inputs() {
			utxo := utxoMapper.GetUTXO(ctx, i.Owner, i.Position)
			totalInput[utxo.Denom] += utxo.Amount
		}

		// Add up all outputs and fee
		totalOutput := map[string]uint64{}
		for _, o := range spendMsg.Outputs() {
			totalOutput[o.Denom] += o.Amount
		}

		for denom, _ := range totalInput {
			if totalInput[denom] != totalOutput[denom] {
				return ctx, utxo.ErrInvalidTransaction(2, "Inputs do not equal Outputs").Result(), true
			}
		}

		// TODO: tx tags (?)
		return ctx, sdk.Result{}, false // continue...
	}
}

func processSig(
	addr common.Address, sig [65]byte, signBytes []byte) (
	res sdk.Result) {

	hash := ethcrypto.Keccak256(signBytes)
	signHash := utils.SignHash(hash)
	pubKey, err := ethcrypto.SigToPub(signHash, sig[:])

	if err != nil || !reflect.DeepEqual(ethcrypto.PubkeyToAddress(*pubKey).Bytes(), addr.Bytes()) {
		return sdk.ErrUnauthorized(fmt.Sprintf("signature verification failed for: %X", addr.Bytes())).Result()
	}

	return sdk.Result{}
}

// Checks that utxo at the position specified exists, matches the address in the SpendMsg
// and returns the denomination associated with the utxo
func checkUTXO(ctx sdk.Context, plasmaClient *eth.Plasma, mapper utxo.Mapper, position types.PlasmaPosition, addr common.Address) sdk.Result {
	var inputAddress []byte
	input := mapper.GetUTXO(ctx, addr.Bytes(), &position)
	if position.IsDeposit() && reflect.DeepEqual(input, utxo.UTXO{}) {
		deposit, ok := DepositExists(position.DepositNum, plasmaClient)
		if !ok {
			return utxo.ErrInvalidUTXO(2, "Deposit UTXO does not exist yet").Result()
		}
		inputAddress = deposit.Owner.Bytes()
	} else {
		if !input.Valid {
			return sdk.ErrUnknownRequest(fmt.Sprintf("UTXO trying to be spent, is not valid: %v.", position)).Result()
		}
		inputAddress = input.Address
	}

	// Verify that utxo owner equals input address in the transaction
	if !reflect.DeepEqual(inputAddress, addr.Bytes()) {
		return sdk.ErrUnauthorized(fmt.Sprintf("signer does not match utxo owner, signer: %X  owner: %X", addr.Bytes(), inputAddress)).Result()
	}
	return sdk.Result{}
}

func DepositExists(nonce uint64, plasmaClient *eth.Plasma) (types.Deposit, bool) {
	deposit, err := plasmaClient.GetDeposit(big.NewInt(int64(nonce)))

	if err != nil {
		return types.Deposit{}, false
	}
	return *deposit, true
}

func hasTXExited(plasmaClient *eth.Plasma, pos types.PlasmaPosition) sdk.Error {
	if plasmaClient == nil {
		return nil
	}

	var positions [4]*big.Int
	for i, num := range pos.Get() {
		positions[i] = big.NewInt(int64(num.Uint64()))
	}
	exited := plasmaClient.HasTXBeenExited(positions)
	if exited {
		return types.ErrInvalidTransaction(types.DefaultCodespace, "Input UTXO has already exited")
	}
	return nil
}
