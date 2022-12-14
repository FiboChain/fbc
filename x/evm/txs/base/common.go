package base

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/FiboChain/fbc/app/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	tmtypes "github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/evm/types"
)

func msg2st(ctx *sdk.Context, k *Keeper, msg *types.MsgEthereumTx) (st types.StateTransition, err error) {
	var chainIDEpoch *big.Int
	chainIDEpoch, err = ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return
	}

	var sender common.Address
	// Verify signature and retrieve sender address
	sender, err = getSender(ctx, chainIDEpoch, msg)
	if err != nil {
		return
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash(ctx.BlockHeight())
	ethHash := common.BytesToHash(txHash)

	st = types.StateTransition{
		AccountNonce: msg.Data.AccountNonce,
		Price:        msg.Data.Price,
		GasLimit:     msg.Data.GasLimit,
		Recipient:    msg.Data.Recipient,
		Amount:       msg.Data.Amount,
		Payload:      msg.Data.Payload,
		Csdb:         types.CreateEmptyCommitStateDB(k.GenerateCSDBParams(), *ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       sender,
		Simulate:     ctx.IsCheckTx(),
		TraceTx:      ctx.IsTraceTx(),
		TraceTxLog:   ctx.IsTraceTxLog(),
	}

	return
}

func getSender(ctx *sdk.Context, chainIDEpoch *big.Int, msg *types.MsgEthereumTx) (sender common.Address, err error) {
	if ctx.IsCheckTx() {
		if from := ctx.From(); len(from) > 0 {
			sender = common.HexToAddress(from)
			return
		}
	}
	err = msg.VerifySig(chainIDEpoch, ctx.BlockHeight())
	if err == nil {
		sender = common.HexToAddress(msg.GetFrom())
	}

	return
}
