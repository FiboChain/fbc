package codec

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec/types"
	"github.com/gogo/protobuf/proto"
)

type (
	// Marshaler defines the interface module codecs must implement in order to support
	// backwards compatibility with Amino while allowing custom Protobuf-based
	// serialization. Note, Amino can still be used without any dependency on
	// Protobuf. There are two typical implementations that fulfill this contract:
	//
	// 1. AminoCodec: Provides full Amino serialization compatibility.
	// 2. ProtoCodec: Provides full Protobuf serialization compatibility.
	Marshaler interface {
		BinaryMarshaler
		JSONMarshaler
	}

	BinaryMarshaler interface {
		MarshalBinaryBare(o ProtoMarshaler) ([]byte, error)
		MustMarshalBinaryBare(o ProtoMarshaler) []byte

		MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error)
		MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte

		UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error
		MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler)

		UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error
		MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler)

		MarshalInterface(i proto.Message) ([]byte, error)
		UnmarshalInterface(bz []byte, ptr interface{}) error

		types.AnyUnpacker
	}

	JSONMarshaler interface {
		MarshalJSON(o proto.Message) ([]byte, error)
		MustMarshalJSON(o proto.Message) []byte
		MarshalInterfaceJSON(i proto.Message) ([]byte, error)
		UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error

		UnmarshalJSON(bz []byte, ptr proto.Message) error
		MustUnmarshalJSON(bz []byte, ptr proto.Message)
	}

	// ProtoMarshaler defines an interface a type must implement as protocol buffer
	// defined message.
	ProtoMarshaler interface {
		proto.Message // for JSON serialization

		Marshal() ([]byte, error)
		MarshalTo(data []byte) (n int, err error)
		MarshalToSizedBuffer(dAtA []byte) (int, error)
		Size() int
		Unmarshal(data []byte) error
	}

	// AminoMarshaler defines an interface where Amino marshalling can be
	// overridden by custom marshalling.
	AminoMarshaler interface {
		MarshalAmino() ([]byte, error)
		UnmarshalAmino([]byte) error
		MarshalAminoJSON() ([]byte, error)
		UnmarshalAminoJSON([]byte) error
	}
	//
	IbcCodec interface {
		BinaryCodec
		JSONCodec
	}

	BinaryCodec interface {
		// Marshal returns binary encoding of v.
		Marshal(o ProtoMarshaler) ([]byte, error)
		// MustMarshal calls Marshal and panics if error is returned.
		MustMarshal(o ProtoMarshaler) []byte

		// MarshalLengthPrefixed returns binary encoding of v with bytes length prefix.
		MarshalLengthPrefixed(o ProtoMarshaler) ([]byte, error)
		// MustMarshalLengthPrefixed calls MarshalLengthPrefixed and panics if
		// error is returned.
		MustMarshalLengthPrefixed(o ProtoMarshaler) []byte

		// Unmarshal parses the data encoded with Marshal method and stores the result
		// in the value pointed to by v.
		Unmarshal(bz []byte, ptr ProtoMarshaler) error
		// MustUnmarshal calls Unmarshal and panics if error is returned.
		MustUnmarshal(bz []byte, ptr ProtoMarshaler)

		// Unmarshal parses the data encoded with UnmarshalLengthPrefixed method and stores
		// the result in the value pointed to by v.
		UnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler) error
		// MustUnmarshalLengthPrefixed calls UnmarshalLengthPrefixed and panics if error
		// is returned.
		MustUnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler)

		// MarshalInterface is a helper method which will wrap `i` into `Any` for correct
		// binary interface (de)serialization.
		MarshalInterface(i proto.Message) ([]byte, error)
		// UnmarshalInterface is a helper method which will parse binary enoded data
		// into `Any` and unpack any into the `ptr`. It fails if the target interface type
		// is not registered in codec, or is not compatible with the serialized data
		UnmarshalInterface(bz []byte, ptr interface{}) error

		types.AnyUnpacker
	}

	JSONCodec interface {
		// MarshalJSON returns JSON encoding of v.
		MarshalJSON(o proto.Message) ([]byte, error)
		// MustMarshalJSON calls MarshalJSON and panics if error is returned.
		MustMarshalJSON(o proto.Message) []byte
		// MarshalInterfaceJSON is a helper method which will wrap `i` into `Any` for correct
		// JSON interface (de)serialization.
		MarshalInterfaceJSON(i proto.Message) ([]byte, error)
		// UnmarshalInterfaceJSON is a helper method which will parse JSON enoded data
		// into `Any` and unpack any into the `ptr`. It fails if the target interface type
		// is not registered in codec, or is not compatible with the serialized data
		UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error

		// UnmarshalJSON parses the data encoded with MarshalJSON method and stores the result
		// in the value pointed to by v.
		UnmarshalJSON(bz []byte, ptr proto.Message) error
		// MustUnmarshalJSON calls Unmarshal and panics if error is returned.
		MustUnmarshalJSON(bz []byte, ptr proto.Message)
	}
)

// ///////
var (
	_ CdcAbstraction = (*CodecProxy)(nil)
)

type CodecProxy struct {
	protoCodec *ProtoCodec
	cdc        *Codec
}

func (mp *CodecProxy) UnmarshalBinaryLengthPrefixedWithRegisteredUbmarshaller(bz []byte, ptr interface{}) (interface{}, error) {
	return mp.cdc.UnmarshalBinaryLengthPrefixedWithRegisteredUbmarshaller(bz, ptr)
}

func (mp *CodecProxy) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	return mp.cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func NewCodecProxy(protoCodec *ProtoCodec, cdc *Codec) *CodecProxy {
	return &CodecProxy{protoCodec: protoCodec, cdc: cdc}
}

func (mp *CodecProxy) GetCdc() *Codec {
	return mp.cdc
}

func (mp *CodecProxy) GetProtocMarshal() *ProtoCodec {
	return mp.protoCodec
}
