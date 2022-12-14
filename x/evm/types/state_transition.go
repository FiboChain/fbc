package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/innertx"
	"github.com/FiboChain/fbc/x/common/analyzer"
)

// StateTransition defines data to transitionDB in evm
type StateTransition struct {
	// TxData fields
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    *common.Address
	Amount       *big.Int
	Payload      []byte

	ChainID    *big.Int
	Csdb       *CommitStateDB
	TxHash     *common.Hash
	Sender     common.Address
	Simulate   bool // i.e CheckTx execution
	TraceTx    bool // reexcute tx or its predesessors
	TraceTxLog bool // trace tx for its evm logs (predesessors are set to false)
}

// GasInfo returns the gas limit, gas consumed and gas refunded from the EVM transition
// execution
type GasInfo struct {
	GasLimit    uint64
	GasConsumed uint64
	GasRefunded uint64
}

// ExecutionResult represents what's returned from a transition
type ExecutionResult struct {
	Logs      []*ethtypes.Log
	Bloom     *big.Int
	Result    *sdk.Result
	GasInfo   GasInfo
	TraceLogs []byte
}

// GetHashFn implements vm.GetHashFunc for Ethermint. It handles 3 cases:
//  1. The requested height matches the current height (and thus same epoch number)
//  2. The requested height is from an previous height from the same chain epoch
//  3. The requested height is from a height greater than the latest one
func GetHashFn(ctx sdk.Context, csdb *CommitStateDB) vm.GetHashFunc {
	return func(height uint64) common.Hash {
		switch {
		case ctx.BlockHeight() == int64(height):
			// Case 1: The requested height matches the one from the context so we can retrieve the header
			// hash directly from the context.
			return csdb.bhash

		case ctx.BlockHeight() > int64(height):
			// Case 2: if the chain is not the current height we need to retrieve the hash from the store for the
			// current chain epoch. This only applies if the current height is greater than the requested height.
			return csdb.WithContext(ctx).GetHeightHash(height)

		default:
			// Case 3: heights greater than the current one returns an empty hash.
			return common.Hash{}
		}
	}
}

func (st *StateTransition) newEVM(
	ctx sdk.Context,
	csdb *CommitStateDB,
	gasLimit uint64,
	gasPrice *big.Int,
	config *ChainConfig,
	vmConfig vm.Config,
) *vm.EVM {
	// Create context for evm
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     GetHashFn(ctx, csdb),
		Coinbase:    common.BytesToAddress(ctx.BlockProposerAddress()),
		BlockNumber: big.NewInt(ctx.BlockHeight()),
		Time:        big.NewInt(ctx.BlockTime().Unix()),
		Difficulty:  big.NewInt(0), // unused. Only required in PoW context
		GasLimit:    gasLimit,
	}

	txCtx := vm.TxContext{
		Origin:   st.Sender,
		GasPrice: gasPrice,
	}

	return vm.NewEVM(blockCtx, txCtx, csdb, config.EthereumConfig(st.ChainID), vmConfig)
}

// TransitionDb will transition the state by applying the current transaction and
// returning the evm execution result.
// NOTE: State transition checks are run during AnteHandler execution.
func (st StateTransition) TransitionDb(ctx sdk.Context, config ChainConfig) (exeRes *ExecutionResult, resData *ResultData, err error, innerTxs, erc20Contracts interface{}) {
	defer func() {
		if e := recover(); e != nil {
			// if the msg recovered can be asserted into type 'ErrContractBlockedVerify', it must be captured by the panics of blocked
			// contract calling
			switch rType := e.(type) {
			case ErrContractBlockedVerify:
				err = ErrCallBlockedContract(rType.Descriptor)
			default:
				panic(e)
			}
		}
	}()

	contractCreation := st.Recipient == nil

	cost, err := core.IntrinsicGas(st.Payload, []ethtypes.AccessTuple{}, contractCreation, config.IsHomestead(), config.IsIstanbul())
	if err != nil {
		return exeRes, resData, sdkerrors.Wrap(err, "invalid intrinsic gas for transaction"), innerTxs, erc20Contracts
	}

	consumedGas := ctx.GasMeter().GasConsumed()
	if consumedGas < cost {
		// If Cosmos standard tx ante handler cost is less than EVM intrinsic cost
		// gas must be consumed to match to accurately simulate an Ethereum transaction
		ctx.GasMeter().ConsumeGas(cost-consumedGas, "Intrinsic gas match")
	}

	// This gas limit the the transaction gas limit with intrinsic gas subtracted
	gasLimit := st.GasLimit - ctx.GasMeter().GasConsumed()

	// This gas meter is set up to consume gas from gaskv during evm execution and be ignored
	currentGasMeter := ctx.GasMeter()
	evmGasMeter := sdk.NewInfiniteGasMeter()
	ctx.SetGasMeter(evmGasMeter)
	csdb := st.Csdb.WithContext(ctx)

	StartTxLog := func(tag string) {
		if !ctx.IsCheckTx() {
			analyzer.StartTxLog(tag)
		}
	}
	StopTxLog := func(tag string) {
		if !ctx.IsCheckTx() {
			analyzer.StopTxLog(tag)
		}
	}

	params := csdb.GetParams()

	var senderStr = EthAddressToString(&st.Sender)

	to := ""
	var recipientStr string
	if st.Recipient != nil {
		to = EthAddressToString(st.Recipient)
		recipientStr = to
	}
	enableDebug := checkTracesSegment(ctx.BlockHeight(), senderStr, to)

	var tracer vm.Tracer
	if st.TraceTxLog || enableDebug {
		tracer = vm.NewStructLogger(evmLogConfig)
	} else {
		tracer = NewNoOpTracer()
	}

	vmConfig := vm.Config{
		ExtraEips:        params.ExtraEIPs,
		Debug:            st.TraceTxLog || enableDebug,
		Tracer:           tracer,
		ContractVerifier: NewContractVerifier(params),
	}

	evm := st.newEVM(ctx, csdb, gasLimit, st.Price, &config, vmConfig)

	var (
		ret             []byte
		leftOverGas     uint64
		contractAddress common.Address
		recipientLog    string
		senderRef       = vm.AccountRef(st.Sender)
	)

	// Get nonce of account outside of the EVM
	currentNonce := csdb.GetNonce(st.Sender)
	// Set nonce of sender account before evm state transition for usage in generating Create address
	csdb.SetNonce(st.Sender, st.AccountNonce)

	//add InnerTx
	callTx := innertx.AddDefaultInnerTx(evm, innertx.CosmosDepth, senderStr, "", "", "", st.Amount, nil)

	// create contract or execute call
	switch contractCreation {
	case true:
		if !params.EnableCreate {
			return exeRes, resData, ErrCreateDisabled, innerTxs, erc20Contracts
		}

		// check whether the deployer address is in the whitelist if the whitelist is enabled
		senderAccAddr := st.Sender.Bytes()
		if params.EnableContractDeploymentWhitelist && !csdb.IsDeployerInWhitelist(senderAccAddr) {
			return exeRes, resData, ErrUnauthorizedAccount(senderAccAddr), innerTxs, erc20Contracts
		}

		StartTxLog(analyzer.EVMCORE)
		defer StopTxLog(analyzer.EVMCORE)
		ret, contractAddress, leftOverGas, err = evm.Create(senderRef, st.Payload, gasLimit, st.Amount)

		contractAddressStr := EthAddressToString(&contractAddress)
		recipientLog = strings.Join([]string{"contract address ", contractAddressStr}, "")

		innertx.UpdateDefaultInnerTx(callTx, contractAddressStr, innertx.CosmosCallType, innertx.EvmCreateName, gasLimit-leftOverGas)
	default:
		if !params.EnableCall {
			return exeRes, resData, ErrCallDisabled, innerTxs, erc20Contracts
		}

		// Increment the nonce for the next transaction	(just for evm state transition)
		csdb.SetNonce(st.Sender, csdb.GetNonce(st.Sender)+1)
		StartTxLog(analyzer.EVMCORE)
		defer StopTxLog(analyzer.EVMCORE)
		ret, leftOverGas, err = evm.Call(senderRef, *st.Recipient, st.Payload, gasLimit, st.Amount)

		if recipientStr == "" {
			recipientStr = EthAddressToString(st.Recipient)
		}

		recipientLog = strings.Join([]string{"recipient address ", recipientStr}, "")

		innertx.UpdateDefaultInnerTx(callTx, recipientStr, innertx.CosmosCallType, innertx.EvmCallName, gasLimit-leftOverGas)
	}

	gasConsumed := gasLimit - leftOverGas

	innerTxs, erc20Contracts = innertx.ParseInnerTxAndContract(evm, err != nil)

	defer func() {
		// Consume gas from evm execution
		// Out of gas check does not need to be done here since it is done within the EVM execution
		currentGasMeter.ConsumeGas(gasConsumed, "EVM execution consumption")
	}()

	defer func() {
		if !st.Simulate && enableDebug && !st.TraceTx {
			result := &core.ExecutionResult{
				UsedGas:    gasConsumed,
				Err:        err,
				ReturnData: ret,
			}
			saveTraceResult(ctx, tracer, result)
		}
	}()
	// return trace log if tracetxlog no matter err = nil  or not nil
	defer func() {
		var traceLogs []byte
		if st.TraceTxLog {
			result := &core.ExecutionResult{
				UsedGas:    gasConsumed,
				Err:        err,
				ReturnData: ret,
			}
			traceLogs, err = GetTracerResult(tracer, result)
			if err != nil {
				traceLogs = []byte(err.Error())
			}
			if exeRes == nil {
				exeRes = &ExecutionResult{
					Result: &sdk.Result{},
				}
			}
			exeRes.TraceLogs = traceLogs
		}
	}()
	if err != nil {
		// Consume gas before returning
		return exeRes, resData, newRevertError(ret, err), innerTxs, erc20Contracts
	}

	// Resets nonce to value pre state transition
	csdb.SetNonce(st.Sender, currentNonce)

	// Generate bloom filter to be saved in tx receipt data
	bloomInt := big.NewInt(0)

	var (
		bloomFilter ethtypes.Bloom
		logs        []*ethtypes.Log
	)

	if st.TxHash != nil && !st.Simulate {
		logs = csdb.GetLogs()

		bloomInt = big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs))
		bloomFilter = ethtypes.BytesToBloom(bloomInt.Bytes())
	}

	if !st.Simulate || st.TraceTx {
		// Finalise state if not a simulated transaction or a trace tx
		// TODO: change to depend on config
		if err = csdb.Finalise(true); err != nil {
			return
		}

		if _, err = csdb.Commit(true); err != nil {
			return
		}
	}

	// Encode all necessary data into slice of bytes to return in sdk result
	resData = &ResultData{
		Bloom:  bloomFilter,
		Logs:   logs,
		Ret:    ret,
		TxHash: *st.TxHash,
	}

	if contractCreation {
		resData.ContractAddress = contractAddress
	}

	resBz, err := EncodeResultData(resData)
	if err != nil {
		return
	}

	resultLog := strings.Join([]string{"executed EVM state transition; sender address ", senderStr, "; ", recipientLog}, "")
	exeRes = &ExecutionResult{
		Logs:  logs,
		Bloom: bloomInt,
		Result: &sdk.Result{
			Data: resBz,
			Log:  resultLog,
		},
		GasInfo: GasInfo{
			GasConsumed: gasConsumed,
			GasLimit:    gasLimit,
			GasRefunded: leftOverGas,
		},
	}
	return
}

func newRevertError(data []byte, e error) error {
	var resultError []string
	if data == nil || e.Error() != vm.ErrExecutionReverted.Error() {
		return e
	}
	resultError = append(resultError, e.Error())
	reason, errUnpack := abi.UnpackRevert(data)
	if errUnpack == nil {
		resultError = append(resultError, vm.ErrExecutionReverted.Error()+":"+reason)
	} else {
		resultError = append(resultError, hexutil.Encode(data))
	}
	resultError = append(resultError, ErrorHexData)
	resultError = append(resultError, hexutil.Encode(data))
	ret, error := json.Marshal(resultError)

	//failed to marshal, return original data in error
	if error != nil {
		return fmt.Errorf(e.Error()+"[%v]", hexutil.Encode(data))
	}
	return errors.New(string(ret))
}
