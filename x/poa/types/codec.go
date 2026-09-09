package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddValidator{}, "itda/x/poa/MsgAddValidator", nil)
	cdc.RegisterConcrete(&MsgRemoveValidator{}, "itda/x/poa/MsgRemoveValidator", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "itda/x/poa/MsgUpdateParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddValidator{},
		&MsgRemoveValidator{},
		&MsgUpdateParams{},
	)

	cryptocodec.RegisterInterfaces(registry)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
