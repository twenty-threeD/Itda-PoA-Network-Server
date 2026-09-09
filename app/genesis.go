package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	authmodule "github.com/cosmos/cosmos-sdk/x/auth"
	bankmodule "github.com/cosmos/cosmos-sdk/x/bank"
	consensusmodule "github.com/cosmos/cosmos-sdk/x/consensus"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa"
)

type GenesisState map[string]json.RawMessage

var ModuleBasics = module.NewBasicManager(
	authmodule.AppModuleBasic{},
	bankmodule.AppModuleBasic{},
	consensusmodule.AppModuleBasic{},
	poa.AppModule{},
	payment.AppModule{},
)

func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}
