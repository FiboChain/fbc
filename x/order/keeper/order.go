package keeper

import (
	"fmt"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"

	"github.com/FiboChain/fbc/x/order/types"
	token "github.com/FiboChain/fbc/x/token/types"
)

// TryPlaceOrder tries to charge fee & lock coins for a new order
func (k Keeper) TryPlaceOrder(ctx sdk.Context, order *types.Order) (fee sdk.SysCoins, err error) {
	logger := ctx.Logger().With("module", "order")
	// Trying to lock coins
	needLockCoins := order.NeedLockCoins()
	err = k.LockCoins(ctx, order.Sender, needLockCoins, token.LockCoinsTypeQuantity)
	if err != nil {
		logger.Info(fmt.Sprintf("place order failed: %v, %v", err, order))
		return fee, err
	}

	// charge fee for placing a new order
	// Note: collected fees stored in cache, make sure handler will be succeed before updating cache
	// Currently, after lock coins successfully, placeOrder will succeed if charging succeed
	fee = GetOrderNewFee(order)

	if err := k.LockCoins(ctx, order.Sender, fee, token.LockCoinsTypeFee); err != nil {
		return fee, err
	}

	return fee, err
}

// PlaceOrder updates BlockOrderNum, DepthBook, execute TryPlaceOrder, and set the specified order to keeper
func (k Keeper) PlaceOrder(ctx sdk.Context, order *types.Order) error {
	fee, err := k.TryPlaceOrder(ctx, order)
	if err != nil {
		return err
	}
	order.RecordOrderNewFee(fee)
	k.AddFeeDetail(ctx, order.Sender, fee, types.FeeTypeOrderNew)

	blockHeight := ctx.BlockHeight()
	orderNum := k.GetBlockOrderNum(ctx, blockHeight)
	order.OrderID = types.FormatOrderID(blockHeight, orderNum+1)

	k.SetBlockOrderNum(ctx, blockHeight, orderNum+1)
	k.SetOrder(ctx, order.OrderID, order)

	// update depth book and orderIDsMap in cache
	k.InsertOrderIntoDepthBook(order)
	return nil
}

// ExpireOrder quits the specified order with the expired state
func (k Keeper) ExpireOrder(ctx sdk.Context, order *types.Order, logger log.Logger) {
	k.quitOrder(ctx, order, types.FeeTypeOrderExpire, logger)
}

// CancelOrder quits the specified order with the canceled state
func (k Keeper) CancelOrder(ctx sdk.Context, order *types.Order, logger log.Logger) sdk.SysCoins {
	return k.quitOrder(ctx, order, types.FeeTypeOrderCancel, logger)
}

// quitOrder unlocks & charges fee, unlocks coins, updates order, and updates DepthBook
func (k Keeper) quitOrder(ctx sdk.Context, order *types.Order, feeType string, logger log.Logger) (fee sdk.SysCoins) {
	switch feeType {
	case types.FeeTypeOrderCancel:
		order.Cancel()
	case types.FeeTypeOrderExpire:
		order.Expire()
	default:
		return
	}

	// unlock coins in this order & charge fee
	needUnlockCoins := order.NeedUnlockCoins()
	k.UnlockCoins(ctx, order.Sender, needUnlockCoins, token.LockCoinsTypeQuantity)

	lockedFee := GetOrderNewFee(order)
	fee = GetOrderCostFee(order, ctx)
	receiveFee := lockedFee.Sub(fee)

	k.UnlockCoins(ctx, order.Sender, lockedFee, token.LockCoinsTypeFee)
	k.AddFeeDetail(ctx, order.Sender, receiveFee, types.FeeTypeOrderReceive)
	order.RecordOrderReceiveFee(receiveFee)

	err := k.AddCollectedFees(ctx, fee, order.Sender, feeType, false)

	if err != nil {
		logger.Error(fmt.Sprintf("failed to charge order(%s) %s fee: %v", feeType, order.OrderID, err))
	}

	order.Unlock()
	k.SetOrder(ctx, order.OrderID, order)

	// remove order from depth book cache
	k.RemoveOrderFromDepthBook(order, feeType)
	return fee
}

func (k Keeper) DropExpiredOrdersByBlockHeight(ctx sdk.Context, expiredBlockHeight int64) {
	logger := ctx.Logger().With("module", "order")
	store := ctx.KVStore(k.orderStoreKey)
	iter := sdk.KVStorePrefixIterator(store, types.GetOrderKey(types.FormatOrderIDPrefix(expiredBlockHeight)))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var order types.Order
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &order)
		if order.Status == types.OrderStatusOpen && !k.IsProductLocked(ctx, order.Product) {
			k.ExpireOrder(ctx, &order, logger)
			logger.Info(fmt.Sprintf("order (%s) expired", order.OrderID))
		}
	}
}
