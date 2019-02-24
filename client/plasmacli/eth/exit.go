package eth

import (
	"fmt"
	"math/big"
	"strconv"

	ks "github.com/FourthState/plasma-mvp-sidechain/client/keystore"
	"github.com/FourthState/plasma-mvp-sidechain/plasma"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tm "github.com/tendermint/tendermint/rpc/core/types"
)

// TODO: Add support for exiting fees

func init() {
	ethCmd.AddCommand(exitCmd)
	exitCmd.Flags().String(feeF, "", "fee committed in an unfinalized spend of the input")
	exitCmd.Flags().BoolP(trustNodeF, "t", false, "trust connected full node")
	exitCmd.Flags().StringP(txBytesF, "b", "", "bytes of the transaction that created the utxo ")
	exitCmd.Flags().StringP(gasLimitF, "g", "21000", "gas limit for ethereum transaction")
	exitCmd.Flags().String(proofF, "", "merkle proof of inclusion")
	exitCmd.Flags().StringP(sigsF, "S", "", "confirmation signatures for the utxo")
	viper.BindPFlags(exitCmd.Flags())
}

var exitCmd = &cobra.Command{
	Use:   "exit <account> <position>",
	Short: "Start an exit for the given position",
	Long: `Starts an exit for the given position. If the trust-node flag is set, 
the necessary information will be retrieved from the connected full node. 
Otherwise, the transaction bytes, merkle proof, and confirmation signatures must be given. 
Usage of flags override information retrieved from full node. 

Usage:
	plasmacli exit <account> <position> -t
	plasmacli exit <account> <position> -t --fee <amount>
	plasmacli exit <account> <position> -b <tx-bytes> --proof <merkle-proof> -S <confirmation-signatures> --fee <amount>`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// parse position
		position, err := plasma.FromPositionString(args[1])
		if err != nil {
			return err
		}

		fee, err := strconv.ParseInt(viper.GetString(feeF), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse fee: { %s }", err)
		}

		gasLimit, err := strconv.ParseUint(viper.GetString(gasLimitF), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse gas limit: { %s }", err)
		}

		// retrieve account key
		key, err := ks.GetKey(args[0])
		if err != nil {
			return fmt.Errorf("failed to retrieve account key: { %s }", err)
		}
		addr := crypto.PubkeyToAddress(key.PublicKey)

		// bind key, generate transact opts
		auth := bind.NewKeyedTransactor(key)
		transactOpts := &bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: gasLimit,
			Value:    big.NewInt(minExitBond),
		}

		// send deposit exit
		if position.IsDeposit() {
			if _, err := rc.contract.StartDepositExit(transactOpts, position.DepositNonce, big.NewInt(fee)); err != nil {
				return fmt.Errorf("failed to start deposit exit: { %s }", err)
			}
			fmt.Println("Started deposit exit")
			return nil
		}

		// retrieve information necessary for transaction exit
		var txBytes, proof, confirmSignatures []byte
		if viper.GetBool(trustNodeF) { // query full node
			var result *tm.ResultTx
			result, confirmSignatures, err = getProof(addr, position)
			if err != nil {
				return fmt.Errorf("failed to retrieve exit information: { %s }", err)
			}

			txBytes = result.Tx

			// flatten proof
			for _, aunt := range result.Proof.Proof.Aunts {
				proof = append(proof, aunt...)
			}
		}

		txBytes, proof, confirmSignatures, err = parseExitFlags(txBytes, proof, confirmSignatures)

		// TODO: Add support for querying for confirm sigs in local storage
		txPos := [3]*big.Int{position.BlockNum, big.NewInt(int64(position.TxIndex)), big.NewInt(int64(position.OutputIndex))}
		if _, err := rc.contract.StartTransactionExit(transactOpts, txPos, txBytes, proof, confirmSignatures, big.NewInt(fee)); err != nil {
			return fmt.Errorf("failed to start transaction exit: { %s }", err)
		}
		fmt.Println("Successfully started transaction exit")
		return nil
	},
}

// Flags override full node information
// All necessary exit information is returned, or error is thrown
func parseExitFlags(txBytes, proof, confirmSignatures []byte) ([]byte, []byte, []byte, error) {
	if viper.GetString(txBytesF) != "" {
		txBytes = []byte(viper.GetString(txBytesF))
	}

	if viper.GetString(proofF) != "" {
		proof = []byte(viper.GetString(proofF))
	}

	if viper.GetString(sigsF) != "" {
		confirmSignatures = []byte(viper.GetString(sigsF))
	}

	// return error if information is missing
	if len(txBytes) == 0 {
		return txBytes, proof, confirmSignatures, fmt.Errorf("please provide txBytes for the given position")
	}

	if len(proof) == 0 {
		return txBytes, proof, confirmSignatures, fmt.Errorf("please provide a merkle proof of inclusion for the given position")
	}

	if len(confirmSignatures) == 0 {
		return txBytes, proof, confirmSignatures, fmt.Errorf("please provde confirmation signatures for the given position")
	}

	return txBytes, proof, confirmSignatures, nil
}