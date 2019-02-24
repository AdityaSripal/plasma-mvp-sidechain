package keys

import (
	ks "github.com/FourthState/plasma-mvp-sidechain/client/keystore"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagName = "name"
)

func init() {
	keysCmd.AddCommand(updateCmd)
	updateCmd.Flags().String(flagName, "", "updated key name.")
	viper.BindPFlags(updateCmd.Flags())
}

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update account passphrase or key name",
	Long: `Update local encrypted private keys to be encrypted with the new passphrase.

Usage:
	plasmacli keys update <name>
	plasmacli keys update <name> --name <updatedName>
	`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		updatedName := viper.GetString(flagName)
		if err := ks.Update(name, updatedName); err != nil {
			return err
		}

		return nil
	},
}