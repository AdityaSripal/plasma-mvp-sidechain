package cmd

import (
	"errors"
	"fmt"

	"github.com/AdityaSripal/plasma-mvp-sidechain/client"
	"github.com/AdityaSripal/plasma-mvp-sidechain/client/context"
	"github.com/AdityaSripal/plasma-mvp-sidechain/utils"
	"github.com/ethereum/go-ethereum/common"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagTo        = "to"
	flagPositions = "position"
	flagAmounts   = "amounts"
)

func init() {
	rootCmd.AddCommand(sendTxCmd)
	sendTxCmd.Flags().String(flagTo, "", "Addresses sending to (separated by commas)")
	// Format for positions can be adjusted
	sendTxCmd.Flags().String(flagPositions, "", "UTXO Positions to be spent, format: blknum0.txindex0.oindex0.depositnonce0::blknum1.txindex1.oindex1.depositnonce1")

	sendTxCmd.Flags().String(flagAmounts, "", "Amounts to be spent, format: amount1, amount2, fee")

	sendTxCmd.Flags().String(client.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	sendTxCmd.Flags().String(client.FlagAddress, "", "Address to sign with")
	viper.BindPFlags(sendTxCmd.Flags())
}

var sendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "Build, Sign, and Send transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.NewClientContextFromViper()

		// get the directory for our keystore
		dir := viper.GetString(FlagHomeDir)

		// get the from/to address
		from, err := ctx.GetInputAddresses(dir)
		if err != nil {
			return err
		}
		toStr := viper.GetString(flagTo)
		if toStr == "" {
			return errors.New("must provide an address to send to")
		}

		toAddrs := strings.Split(toStr, ",")
		if len(toAddrs) > 2 {
			return errors.New("incorrect amount of addresses provided")
		}

		var addr1, addr2 common.Address
		addr1, err = client.StrToAddress(toAddrs[0])
		if err != nil {
			return err
		}
		if len(toAddrs) > 1 && toAddrs[1] != "" {
			addr2, err = client.StrToAddress(toAddrs[1])
			if err != nil {
				return err
			}
		}

		// Get positions for transaction inputs
		posStr := viper.GetString(flagPositions)
		position, err := client.ParsePositions(posStr)
		if err != nil {
			return err
		}

		// Get amounts
		amtStr := viper.GetString(flagAmounts)
		amounts, err := client.ParseAmounts(amtStr)
		if utils.ZeroAddress(addr2) && amounts[1] != 0 {
			return fmt.Errorf("You are trying to send %d amount to the nil address. Please input the zero address if you would like to burn your amount", amounts[1])
		}
		msg := client.BuildMsg(from[0], from[1], addr1, addr2, position[0], position[1], amounts[0], amounts[1])
		res, err := ctx.SignBuildBroadcast(from, msg, dir)
		if err != nil {
			return err
		}
		fmt.Printf("Committed at block %d. Hash %s\n", res.Height, res.Hash.String())
		return nil
	},
}
