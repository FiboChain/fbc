package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	"github.com/gogo/protobuf/proto"
)

const (
	IBCROUTER = "ibc"
)

type MsgProtoAdapter interface {
	Msg
	codec.ProtoMarshaler
}
type MsgAdapter interface {
	Msg
	proto.Message
}

// MsgTypeURL returns the TypeURL of a `sdk.Msg`.
func MsgTypeURL(msg proto.Message) string {
	return "/" + proto.MessageName(msg)
}
