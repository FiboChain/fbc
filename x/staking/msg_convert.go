package staking

import (
	"encoding/json"
	"errors"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/baseapp"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	tmtypes "github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/common"
	"github.com/FiboChain/fbc/x/staking/types"
)

var (
	ErrCheckSignerFail = errors.New("check signer fail")
)

func init() {
	RegisterConvert()
}

func RegisterConvert() {
	enableHeight := tmtypes.GetVenus3Height()
	baseapp.RegisterCmHandle("fibochain/staking/MsgDeposit", baseapp.NewCMHandle(ConvertDepositMsg, enableHeight))
	baseapp.RegisterCmHandle("fibochain/staking/MsgWithdraw", baseapp.NewCMHandle(ConvertWithdrawMsg, enableHeight))
	baseapp.RegisterCmHandle("fibochain/staking/MsgAddShares", baseapp.NewCMHandle(ConvertAddSharesMsg, enableHeight))
}

func ConvertDepositMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgDeposit{}
	err := json.Unmarshal(data, &newMsg)
	if err != nil {
		return nil, err
	}
	err = newMsg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	if ok := common.CheckSignerAddress(signers, newMsg.GetSigners()); !ok {
		return nil, ErrCheckSignerFail
	}
	return newMsg, nil
}

func ConvertWithdrawMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgWithdraw{}
	err := json.Unmarshal(data, &newMsg)
	if err != nil {
		return nil, err
	}
	err = newMsg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	if ok := common.CheckSignerAddress(signers, newMsg.GetSigners()); !ok {
		return nil, ErrCheckSignerFail
	}
	return newMsg, nil
}

func ConvertAddSharesMsg(data []byte, signers []sdk.AccAddress) (sdk.Msg, error) {
	newMsg := types.MsgAddShares{}
	err := json.Unmarshal(data, &newMsg)
	if err != nil {
		return nil, err
	}
	err = newMsg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	if ok := common.CheckSignerAddress(signers, newMsg.GetSigners()); !ok {
		return nil, ErrCheckSignerFail
	}
	return newMsg, nil
}
