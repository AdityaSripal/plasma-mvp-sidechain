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

// The AnteHandler is responsible for checking the validity of the SpendMsg
// given the current state of the blockchain
// NewAnteHandler returns an AnteHandler that checks signatures, adds in deposit UTXOs,
// and checks balances.
func NewAnteHandler(utxoMapper utxo.Mapper, plasmaStore kvstore.KVStore, plasmaClient *eth.Plasma) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (_ sdk.Context, _ sdk.Result, abort bool) {

		baseTx, ok := tx.(types.BaseTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be in form of BaseTx").Result(), true
		}

		// TODO: Check signatures
		// Verify the first input signature

		// TODO: Implement CheckUTXO
		res := checkUTXO(ctx, plasmaClient, utxoMapper, position0, addr0)
		if !res.IsOK() {
			return ctx, res, true
		}

		// Check that UTXO has not exitted
		exitErr := hasTXExited(plasmaClient, position0)
		if exitErr != nil {
			return ctx, exitErr.Result(), true
		}
		// Add in deposit UTXO to store if it doesn't already exist.
		if position0.IsDeposit() {
			deposit, _ := DepositExists(position0.DepositNum, plasmaClient)
			inputUTXO := utxo.NewUTXO(deposit.Owner.Bytes(), deposit.Amount.Uint64(), types.Denom, position0)
			utxoMapper.ReceiveUTXO(ctx, inputUTXO)
		}

		// Verify the second input
		if utils.ValidAddress(spendMsg.Owner1) {
			// TODO: Check signatures

			addr1 := common.BytesToAddress(signerAddrs[1].Bytes())
			position1 := types.PlasmaPosition{spendMsg.Blknum1, spendMsg.Txindex1, spendMsg.Oindex1, spendMsg.DepositNum1}
			// Check that second input has not exitted
			exitErr := hasTXExited(plasmaClient, position1)
			if exitErr != nil {
				return ctx, exitErr.Result(), true
			}

			// Check that second input exists in UTXO store and is unspent
			res := checkUTXO(ctx, plasmaClient, utxoMapper, position1, addr1)
			if !res.IsOK() {
				return ctx, res, true
			}
			// Add in second input to UTXO store if it is a deposit.
			if position1.IsDeposit() {
				deposit, _ := DepositExists(position1.DepositNum, plasmaClient)
				inputUTXO := utxo.NewUTXO(deposit.Owner.Bytes(), deposit.Amount.Uint64(), types.Denom, position1)
				utxoMapper.ReceiveUTXO(ctx, inputUTXO)
			}

		}

		// TODO: Check that total balance of Inputs == total balance of Outputs

		return ctx, sdk.Result{}, false // continue...
	}
}

func processSig(
	addr common.Address, sig [65]byte, signBytes []byte) (
	res sdk.Result) {

	// Check signatures the way that Ethereum does
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

	// TODO: Check that either UTXO exists in the sidechain's UTXO store and is unspent
	// OR if input is a deposit, then check that the deposit with corresponding
	// nonce has been finalized already on rootchain

	// Then check that the address trying to spend UTXO is same address that owns UTXO

	return sdk.Result{}
}

// Queries plasmaClient to see if a finalized deposit exists with given nonce
func DepositExists(nonce uint64, plasmaClient *eth.Plasma) (types.Deposit, bool) {
	deposit, err := plasmaClient.GetDeposit(big.NewInt(int64(nonce)))

	if err != nil {
		return types.Deposit{}, false
	}
	return *deposit, true
}

// Queries plasmaClient to see if a UTXO position has exitted/is exitting.
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
