package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/privval"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	paymenttypes "github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
	poatypes "github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

const (
	flagMoniker = "moniker"
	flagPower   = "power"
)

func AddGenesisValidatorCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-validator [operator-address]",
		Short: "Add a validator to the poa genesis validator set using this node's consensus key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			operatorAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid operator address: %w", err)
			}

			moniker, err := cmd.Flags().GetString(flagMoniker)
			if err != nil {
				return err
			}

			power, err := cmd.Flags().GetInt64(flagPower)
			if err != nil {
				return err
			}

			pubKey, err := loadConsensusPubKey(config)
			if err != nil {
				return err
			}

			validator, err := poatypes.NewValidator(operatorAddress.String(), pubKey, moniker, power)
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()

			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			var poaGenesis poatypes.GenesisState
			if err := clientCtx.Codec.UnmarshalJSON(appState[poatypes.ModuleName], &poaGenesis); err != nil {
				return err
			}

			for _, existing := range poaGenesis.Validators {
				if existing.OperatorAddress == validator.OperatorAddress {
					return fmt.Errorf("validator %s is already in genesis", validator.OperatorAddress)
				}
			}

			poaGenesis.Validators = append(poaGenesis.Validators, validator)

			poaGenesisBz, err := clientCtx.Codec.MarshalJSON(&poaGenesis)
			if err != nil {
				return err
			}

			appState[poatypes.ModuleName] = poaGenesisBz

			appStateJSON, err := jsonMarshal(appState)
			if err != nil {
				return err
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flagMoniker, "validator", "Human readable label for the validator")
	cmd.Flags().Int64(flagPower, 1, "Voting power to assign to the validator")

	return cmd
}

func SetGenesisAuthorityCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-authority [address]",
		Short: "Set the poa and payment module authority in genesis",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			authority, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid authority address: %w", err)
			}

			genFile := config.GenesisFile()

			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			var poaGenesis poatypes.GenesisState
			if err := clientCtx.Codec.UnmarshalJSON(appState[poatypes.ModuleName], &poaGenesis); err != nil {
				return err
			}

			poaGenesis.Authority = authority.String()

			poaGenesisBz, err := clientCtx.Codec.MarshalJSON(&poaGenesis)
			if err != nil {
				return err
			}

			appState[poatypes.ModuleName] = poaGenesisBz

			var paymentGenesis paymenttypes.GenesisState
			if err := clientCtx.Codec.UnmarshalJSON(appState[paymenttypes.ModuleName], &paymentGenesis); err != nil {
				return err
			}

			paymentGenesis.Authority = authority.String()

			paymentGenesisBz, err := clientCtx.Codec.MarshalJSON(&paymentGenesis)
			if err != nil {
				return err
			}

			appState[paymenttypes.ModuleName] = paymentGenesisBz

			appStateJSON, err := jsonMarshal(appState)
			if err != nil {
				return err
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")

	return cmd
}

func loadConsensusPubKey(config *cmtconfig.Config) (cryptotypes.PubKey, error) {
	privValidator := privval.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())

	cmtPubKey, err := privValidator.GetPubKey()
	if err != nil {
		return nil, err
	}

	return cryptocodec.FromCmtPubKeyInterface(cmtPubKey)
}

func jsonMarshal(appState map[string]json.RawMessage) (json.RawMessage, error) {
	return json.MarshalIndent(appState, "", "  ")
}
