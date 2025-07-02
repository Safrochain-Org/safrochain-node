package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Params{}, "cw-hooks/Params", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cw-hooks/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterStaking{}, "cw-hooks/MsgRegisterStaking")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterGovernance{}, "cw-hooks/MsgRegisterGovernance")
	legacy.RegisterAminoMsg(cdc, &MsgUnregisterGovernance{}, "cw-hooks/MsgUnregisterGovernance")
	legacy.RegisterAminoMsg(cdc, &MsgUnregisterStaking{}, "cw-hooks/MsgUnregisterStaking")
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgRegisterGovernance{},
		&MsgRegisterStaking{},
		&MsgUnregisterGovernance{},
		&MsgUnregisterStaking{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
