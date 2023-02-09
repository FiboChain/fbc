package common

import (
	"errors"
	"fmt"

	"github.com/FiboChain/fbc/libs/tendermint/types"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

// IBC port sentinel errors
var (
	ErrDisableProxyBeforeHeight = sdkerrors.Register(ModuleProxy, 1, "this feature is disable")
)

func MsgNotSupportBeforeHeight(msg proto.Message, h int64) error {
	if types.HigherThanVenus4(h) {
		return nil
	}
	return errors.New(fmt.Sprintf("msg:%s not support before height:%d", sdk.MsgTypeURL(msg), types.GetVenus4Height()))
}
