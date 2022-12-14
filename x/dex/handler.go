package dex

import (
	"fmt"
	"strconv"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	types2 "github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/common"
	"github.com/FiboChain/fbc/x/common/perf"
	"github.com/FiboChain/fbc/x/dex/types"
)

// NewHandler handles all "dex" type messages.
func NewHandler(k IKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		// disable dex tx handler
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Dex messages are not allowd.")

		ctx = ctx.WithEventManager(sdk.NewEventManager())
		logger := ctx.Logger().With("module", ModuleName)

		var handlerFun func() (*sdk.Result, error)
		var name string
		switch msg := msg.(type) {
		case MsgList:
			name = "handleMsgList"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgList(ctx, k, msg, logger)
			}
		case MsgDeposit:
			name = "handleMsgDeposit"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgDeposit(ctx, k, msg, logger)
			}
		case MsgWithdraw:
			name = "handleMsgWithDraw"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgWithDraw(ctx, k, msg, logger)
			}
		case MsgTransferOwnership:
			name = "handleMsgTransferOwnership"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgTransferOwnership(ctx, k, msg, logger)
			}
		case MsgConfirmOwnership:
			name = "handleMsgConfirmOwnership"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgConfirmOwnership(ctx, k, msg, logger)
			}
		case MsgCreateOperator:
			name = "handleMsgCreateOperator"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgCreateOperator(ctx, k, msg, logger)
			}
		case MsgUpdateOperator:
			name = "handleMsgUpdateOperator"
			handlerFun = func() (*sdk.Result, error) {
				return handleMsgUpdateOperator(ctx, k, msg, logger)
			}
		default:
			return types.ErrDexUnknownMsgType(msg.Type()).Result()
		}

		seq := perf.GetPerf().OnDeliverTxEnter(ctx, ModuleName, name)
		defer perf.GetPerf().OnDeliverTxExit(ctx, ModuleName, name, seq)

		res, err := handlerFun()
		common.SanityCheckHandler(res, err)
		return res, err
	}
}

func handleMsgList(ctx sdk.Context, keeper IKeeper, msg MsgList, logger log.Logger) (*sdk.Result, error) {

	if !keeper.GetTokenKeeper().TokenExist(ctx, msg.ListAsset) ||
		!keeper.GetTokenKeeper().TokenExist(ctx, msg.QuoteAsset) {
		return types.ErrTokenOfPairNotExist(msg.ListAsset, msg.QuoteAsset).Result()
	}

	if _, exists := keeper.GetOperator(ctx, msg.Owner); !exists {
		return types.ErrUnknownOperator(msg.Owner).Result()
	}

	tokenPair := &TokenPair{
		BaseAssetSymbol:  msg.ListAsset,
		QuoteAssetSymbol: msg.QuoteAsset,
		InitPrice:        msg.InitPrice,
		MaxPriceDigit:    int64(DefaultMaxPriceDigitSize),
		MaxQuantityDigit: int64(DefaultMaxQuantityDigitSize),
		MinQuantity:      sdk.MustNewDecFromStr("0.00000001"),
		Owner:            msg.Owner,
		Delisting:        false,
		Deposits:         DefaultTokenPairDeposit,
		BlockHeight:      ctx.BlockHeight(),
	}

	// check whether a specific token pair exists with the symbols of base asset and quote asset
	// Note: aaa_bbb and bbb_aaa are actually one token pair
	if keeper.GetTokenPair(ctx, fmt.Sprintf("%s_%s", tokenPair.BaseAssetSymbol, tokenPair.QuoteAssetSymbol)) != nil ||
		keeper.GetTokenPair(ctx, fmt.Sprintf("%s_%s", tokenPair.QuoteAssetSymbol, tokenPair.BaseAssetSymbol)) != nil {
		return types.ErrTokenPairExisted(tokenPair.BaseAssetSymbol, tokenPair.QuoteAssetSymbol).Result()
	}

	// deduction fee
	feeCoins := keeper.GetParams(ctx).ListFee.ToCoins()
	err := keeper.GetSupplyKeeper().SendCoinsFromAccountToModule(ctx, msg.Owner, keeper.GetFeeCollector(), feeCoins)
	if err != nil {
		return types.ErrInsufficientFeeCoins(feeCoins).Result()
	}

	err2 := keeper.SaveTokenPair(ctx, tokenPair)
	if err2 != nil {
		return types.ErrTokenPairSaveFailed(err2.Error()).Result()
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgList: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("list-asset", tokenPair.BaseAssetSymbol),
			sdk.NewAttribute("quote-asset", tokenPair.QuoteAssetSymbol),
			sdk.NewAttribute("init-price", tokenPair.InitPrice.String()),
			sdk.NewAttribute("max-price-digit", strconv.FormatInt(tokenPair.MaxPriceDigit, 10)),
			sdk.NewAttribute("max-size-digit", strconv.FormatInt(tokenPair.MaxQuantityDigit, 10)),
			sdk.NewAttribute("min-trade-size", tokenPair.MinQuantity.String()),
			sdk.NewAttribute("delisting", fmt.Sprintf("%t", tokenPair.Delisting)),
			sdk.NewAttribute(sdk.AttributeKeyFee, feeCoins.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgDeposit(ctx sdk.Context, keeper IKeeper, msg MsgDeposit, logger log.Logger) (*sdk.Result, error) {
	confirmOwnership, exist := keeper.GetConfirmOwnership(ctx, msg.Product)
	if exist && !ctx.BlockTime().After(confirmOwnership.Expire) {
		return types.ErrIsTransferringOwner(msg.Product).Result()
	}
	if sdkErr := keeper.Deposit(ctx, msg.Product, msg.Depositor, msg.Amount); sdkErr != nil {
		return types.ErrDepositFailed(sdkErr.Error()).Result()
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgDeposit: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil

}

func handleMsgWithDraw(ctx sdk.Context, keeper IKeeper, msg MsgWithdraw, logger log.Logger) (*sdk.Result, error) {
	if sdkErr := keeper.Withdraw(ctx, msg.Product, msg.Depositor, msg.Amount); sdkErr != nil {
		return nil, sdkErr
	}

	logger.Debug(fmt.Sprintf("successfully handleMsgWithDraw: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgTransferOwnership(ctx sdk.Context, keeper IKeeper, msg MsgTransferOwnership,
	logger log.Logger) (*sdk.Result, error) {
	// validate
	tokenPair := keeper.GetTokenPair(ctx, msg.Product)
	if tokenPair == nil {
		return types.ErrTokenPairNotFound(msg.Product).Result()
	}
	if !tokenPair.Owner.Equals(msg.FromAddress) {
		return types.ErrUnauthorized(msg.FromAddress.String(), msg.Product).Result()
	}
	if _, exist := keeper.GetOperator(ctx, msg.ToAddress); !exist {
		return types.ErrUnknownOperator(msg.ToAddress).Result()
	}
	confirmOwnership, exist := keeper.GetConfirmOwnership(ctx, msg.Product)
	if exist && !ctx.BlockTime().After(confirmOwnership.Expire) {
		return types.ErrRepeatedTransferOwner(msg.Product).Result()
	}

	// withdraw
	if tokenPair.Deposits.IsPositive() {
		if err := keeper.Withdraw(ctx, msg.Product, msg.FromAddress, tokenPair.Deposits); err != nil {
			return types.ErrWithdrawFailed(err.Error()).Result()
		}
	}

	// deduction fee
	feeCoins := keeper.GetParams(ctx).TransferOwnershipFee.ToCoins()
	err := keeper.GetSupplyKeeper().SendCoinsFromAccountToModule(ctx, msg.FromAddress, keeper.GetFeeCollector(), feeCoins)
	if err != nil {
		return types.ErrInsufficientFeeCoins(feeCoins).Result()
	}

	// set ConfirmOwnership
	expireTime := ctx.BlockTime().Add(keeper.GetParams(ctx).OwnershipConfirmWindow)
	confirmOwnership = &types.ConfirmOwnership{
		Product:     msg.Product,
		FromAddress: msg.FromAddress,
		ToAddress:   msg.ToAddress,
		Expire:      expireTime,
	}
	keeper.SetConfirmOwnership(ctx, confirmOwnership)

	logger.Debug(fmt.Sprintf("successfully handleMsgTransferOwnership: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyFee, feeCoins.String()),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgConfirmOwnership(ctx sdk.Context, keeper IKeeper, msg MsgConfirmOwnership, logger log.Logger) (*sdk.Result, error) {
	confirmOwnership, exist := keeper.GetConfirmOwnership(ctx, msg.Product)
	if !exist {
		return types.ErrGetConfirmOwnershipNotExist(msg.Address.String()).Result()
	}
	if ctx.BlockTime().After(confirmOwnership.Expire) {
		// delete ownership confirming information
		keeper.DeleteConfirmOwnership(ctx, confirmOwnership.Product)
		return types.ErrIsTransferOwnerExpired(confirmOwnership.Expire.String()).Result()
	}
	if !confirmOwnership.ToAddress.Equals(msg.Address) {
		return types.ErrUnauthorized(confirmOwnership.ToAddress.String(), msg.Product).Result()
	}

	tokenPair := keeper.GetTokenPair(ctx, msg.Product)
	if tokenPair == nil {
		return types.ErrTokenPairNotFound(msg.Product).Result()
	}
	// transfer ownership
	tokenPair.Owner = msg.Address
	keeper.UpdateTokenPair(ctx, msg.Product, tokenPair)
	keeper.UpdateUserTokenPair(ctx, msg.Product, confirmOwnership.FromAddress, msg.Address)
	// delete ownership confirming information
	keeper.DeleteConfirmOwnership(ctx, confirmOwnership.Product)

	logger.Debug(fmt.Sprintf("successfully handleMsgConfirmOwnership: "+
		"BlockHeight: %d, Msg: %+v", ctx.BlockHeight(), msg))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		),
	)
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgCreateOperator(ctx sdk.Context, keeper IKeeper, msg MsgCreateOperator, logger log.Logger) (*sdk.Result, error) {

	logger.Debug(fmt.Sprintf("handleMsgCreateOperator msg: %+v", msg))

	if _, isExist := keeper.GetOperator(ctx, msg.Owner); isExist {
		return types.ErrExistOperator(msg.Owner).Result()
	}
	operator := types.DEXOperator{
		Address:            msg.Owner,
		HandlingFeeAddress: msg.HandlingFeeAddress,
		Website:            msg.Website,
		InitHeight:         ctx.BlockHeight(),
		TxHash:             fmt.Sprintf("%X", types2.Tx(ctx.TxBytes()).Hash(ctx.BlockHeight())),
	}
	keeper.SetOperator(ctx, operator)

	// deduction fee
	feeCoins := keeper.GetParams(ctx).RegisterOperatorFee.ToCoins()
	err := keeper.GetSupplyKeeper().SendCoinsFromAccountToModule(ctx, msg.Owner, keeper.GetFeeCollector(), feeCoins)
	if err != nil {
		return common.ErrInsufficientCoins(DefaultParamspace, err.Error()).Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyFee, feeCoins.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgUpdateOperator(ctx sdk.Context, keeper IKeeper, msg MsgUpdateOperator, logger log.Logger) (*sdk.Result, error) {

	logger.Debug(fmt.Sprintf("handleMsgUpdateOperator msg: %+v", msg))

	operator, isExist := keeper.GetOperator(ctx, msg.Owner)
	if !isExist {
		return types.ErrUnknownOperator(msg.Owner).Result()
	}
	if !operator.Address.Equals(msg.Owner) {
		return types.ErrUnauthorizedOperator(operator.Address.String(), msg.Owner.String()).Result()
	}

	operator.HandlingFeeAddress = msg.HandlingFeeAddress
	operator.Website = msg.Website

	keeper.SetOperator(ctx, operator)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
