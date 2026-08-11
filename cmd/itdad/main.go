package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/app"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/cmd/itdad/cmd"
)

func main() {
	app.SetBech32Prefixes()

	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "ITDA", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
