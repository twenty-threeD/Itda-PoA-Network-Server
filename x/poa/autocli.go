package poa

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "itda.poa.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the module parameters",
				},
				{
					RpcMethod: "Validators",
					Use:       "validators",
					Short:     "List the active validator set",
				},
				{
					RpcMethod:      "Validator",
					Use:            "validator [operator-address]",
					Short:          "Query a single validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "operator_address"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "itda.poa.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "RemoveValidator",
					Use:            "remove-validator [operator-address]",
					Short:          "Remove a validator from the active set",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "operator_address"}},
				},
				{
					RpcMethod: "AddValidator",
					Skip:      true,
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
		},
	}
}
