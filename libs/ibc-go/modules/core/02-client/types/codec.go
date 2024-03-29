package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	codectypes "github.com/FiboChain/fbc/libs/cosmos-sdk/codec/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	txmsg "github.com/FiboChain/fbc/libs/cosmos-sdk/types/ibc-adapter"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/msgservice"
	"github.com/FiboChain/fbc/libs/ibc-go/modules/core/exported"
	"github.com/FiboChain/fbc/x/gov/types"
	"github.com/gogo/protobuf/proto"
)

func init() {
	types.RegisterProposalTypeCodec(&ClientUpdateProposal{}, "ibc.core.client.v1.ClientUpdateProposal")
	types.RegisterProposalTypeCodec(&MsgUpgradeClient{}, "ibc.core.client.v1.MsgUpgradeClient")
}

// RegisterInterfaces registers the client interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {

	registry.RegisterInterface(
		"ibc.core.client.v1.ClientState",
		(*exported.ClientState)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.ConsensusState",
		(*exported.ConsensusState)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.Header",
		(*exported.Header)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.Height",
		(*exported.Height)(nil),
		&Height{},
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.Misbehaviour",
		(*exported.Misbehaviour)(nil),
	)
	registry.RegisterImplementations(
		(*sdk.MsgAdapter)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgUpgradeClient{},
		&MsgSubmitMisbehaviour{},
	)
	registry.RegisterImplementations(
		(*txmsg.Msg)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgUpgradeClient{},
		&MsgSubmitMisbehaviour{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&ClientUpdateProposal{}, "ibc.core.client.v1.ClientUpdateProposal", nil)
	cdc.RegisterConcrete(&MsgUpgradeClient{}, "ibc.core.client.v1.MsgUpgradeClient", nil)
}

// UnpackClientState unpacks an Any into a ClientState. It returns an error if the
// client state can't be unpacked into a ClientState.
func UnpackClientState(any *codectypes.Any) (exported.ClientState, error) {
	if any == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	clientState, ok := any.GetCachedValue().(exported.ClientState)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack Any into ClientState %T", any)
	}

	return clientState, nil
}

// PackClientState constructs a new Any packed with the given client state value. It returns
// an error if the client state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackClientState(clientState exported.ClientState) (*codectypes.Any, error) {
	msg, ok := clientState.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", clientState)
	}

	anyClientState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, err.Error())
	}

	return anyClientState, nil
}

// PackConsensusState constructs a new Any packed with the given consensus state value. It returns
// an error if the consensus state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackConsensusState(consensusState exported.ConsensusState) (*codectypes.Any, error) {
	msg, ok := consensusState.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", consensusState)
	}

	anyConsensusState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, err.Error())
	}

	return anyConsensusState, nil
}

// MustPackConsensusState calls PackConsensusState and panics on error.
func MustPackConsensusState(consensusState exported.ConsensusState) *codectypes.Any {
	anyConsensusState, err := PackConsensusState(consensusState)
	if err != nil {
		panic(err)
	}

	return anyConsensusState
}

// UnpackConsensusState unpacks an Any into a ConsensusState. It returns an error if the
// consensus state can't be unpacked into a ConsensusState.
func UnpackConsensusState(any *codectypes.Any) (exported.ConsensusState, error) {
	if any == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	consensusState, ok := any.GetCachedValue().(exported.ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack Any into ConsensusState %T", any)
	}

	return consensusState, nil
}

// PackHeader constructs a new Any packed with the given header value. It returns
// an error if the header can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackHeader(header exported.Header) (*codectypes.Any, error) {
	msg, ok := header.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", header)
	}

	anyHeader, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, err.Error())
	}

	return anyHeader, nil
}

// UnpackHeader unpacks an Any into a Header. It returns an error if the
// consensus state can't be unpacked into a Header.
func UnpackHeader(any *codectypes.Any) (exported.Header, error) {
	if any == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	header, ok := any.GetCachedValue().(exported.Header)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack Any into Header %T", any)
	}

	return header, nil
}

// PackMisbehaviour constructs a new Any packed with the given misbehaviour value. It returns
// an error if the misbehaviour can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackMisbehaviour(misbehaviour exported.Misbehaviour) (*codectypes.Any, error) {
	msg, ok := misbehaviour.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", misbehaviour)
	}

	anyMisbhaviour, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, err.Error())
	}

	return anyMisbhaviour, nil
}

// UnpackMisbehaviour unpacks an Any into a Misbehaviour. It returns an error if the
// Any can't be unpacked into a Misbehaviour.
func UnpackMisbehaviour(any *codectypes.Any) (exported.Misbehaviour, error) {
	if any == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	misbehaviour, ok := any.GetCachedValue().(exported.Misbehaviour)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack Any into Misbehaviour %T", any)
	}

	return misbehaviour, nil
}
