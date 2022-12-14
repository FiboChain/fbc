package keeper

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/common"
	"github.com/FiboChain/fbc/x/order/types"
)

// ===============================================

// SetBlockOrderNum sets BlockOrderNum to keeper
// BlockOrderNum means OrderNum at a given BlockHeight
func (k Keeper) SetBlockOrderNum(ctx sdk.Context, blockHeight int64, orderNum int64) {
	store := ctx.KVStore(k.orderStoreKey)
	key := types.GetOrderNumPerBlockKey(blockHeight)
	store.Set(key, common.Int64ToBytes(orderNum))
}

// DropBlockOrderNum deletes OrderNum from keeper
func (k Keeper) DropBlockOrderNum(ctx sdk.Context, blockHeight int64) {
	store := ctx.KVStore(k.orderStoreKey)
	key := types.GetOrderNumPerBlockKey(blockHeight)
	store.Delete(key)
}

// ===============================================

// SetExpireBlockHeight sets ExpireBlockHeight to keeper
// ExpireBlockHeight means a slice of expired height that need to be solved
func (k Keeper) SetExpireBlockHeight(ctx sdk.Context, blockHeight int64, expireBlockHeight []int64) {
	store := ctx.KVStore(k.orderStoreKey)
	key := types.GetExpireBlockHeightKey(blockHeight)
	store.Set(key, k.cdc.MustMarshalBinaryBare(expireBlockHeight))
}

// DropExpireBlockHeight deletes ExpireBlockHeight from keeper
func (k Keeper) DropExpireBlockHeight(ctx sdk.Context, blockHeight int64) {
	store := ctx.KVStore(k.orderStoreKey)
	key := types.GetExpireBlockHeightKey(blockHeight)
	store.Delete(key)
}

// ===============================================
// nolint
func (k Keeper) SetOrder(ctx sdk.Context, orderID string, order *types.Order) {
	store := ctx.KVStore(k.orderStoreKey)
	store.Set(types.GetOrderKey(orderID), k.cdc.MustMarshalBinaryBare(order))
}

// nolint
func (k Keeper) DropOrder(ctx sdk.Context, orderID string) {
	store := ctx.KVStore(k.orderStoreKey)
	store.Delete(types.GetOrderKey(orderID))
}

// ===============================================
// nolint
func (k Keeper) StoreDepthBook(ctx sdk.Context, product string, depthBook *types.DepthBook) {
	store := ctx.KVStore(k.orderStoreKey)
	if depthBook == nil || len(depthBook.Items) == 0 {
		store.Delete(types.GetDepthBookKey(product))
	} else {
		store.Set(types.GetDepthBookKey(product), k.cdc.MustMarshalBinaryBare(depthBook)) //
	}
}

// ===============================================
// LastPrice means the latest transaction price of a given Product
// nolint
func (k Keeper) SetLastPrice(ctx sdk.Context, product string, price sdk.Dec) {
	store := ctx.KVStore(k.orderStoreKey)
	store.Set(types.GetPriceKey(product), k.cdc.MustMarshalBinaryBare(price))
	k.diskCache.setLastPrice(product, price)
}

// ===============================================
// nolint
func (k Keeper) StoreOrderIDsMap(ctx sdk.Context, key string, orderIDs []string) {
	store := ctx.KVStore(k.orderStoreKey)
	if len(orderIDs) == 0 {
		store.Delete(types.GetOrderIDsKey(key))
	} else {
		store.Set(types.GetOrderIDsKey(key), k.cdc.MustMarshalJSON(orderIDs)) //StoreOrderIDsMap
	}
}

// ===============================================
// 7.

//func (k Keeper) updateStoreProductLockMap(ctx sdk.Context, lockMap *types.ProductLockMap) {
//	store := ctx.KVStore(k.otherStoreKey)
//	bz, _ := json.Marshal(lockMap)
//	store.Set([]byte("productLockMap"), bz)
//}
//
//

// ===============================================
// LastExpiredBlockHeight means that the block height of his expired height
// list has been processed by expired recently
// nolint
func (k Keeper) SetLastExpiredBlockHeight(ctx sdk.Context, expiredBlockHeight int64) {
	store := ctx.KVStore(k.orderStoreKey)
	store.Set(types.LastExpiredBlockHeightKey, common.Int64ToBytes(expiredBlockHeight)) //lastExpiredBlockHeight
}

// ===============================================
// OpenOrderNum means the number of orders currently in the open state
// nolint
func (k Keeper) setOpenOrderNum(ctx sdk.Context, orderNum int64) {
	store := ctx.KVStore(k.orderStoreKey)
	store.Set(types.OpenOrderNumKey, common.Int64ToBytes(orderNum)) //openOrderNum
}

// ===============================================
// StoreOrderNum means the number of orders currently stored
// nolint
func (k Keeper) setStoreOrderNum(ctx sdk.Context, orderNum int64) {
	store := ctx.KVStore(k.orderStoreKey)
	store.Set(types.StoreOrderNumKey, common.Int64ToBytes(orderNum)) //StoreOrderNum
}

// ===============================================

// SetLastClosedOrderIDs sets closed order ids in this block
func (k Keeper) SetLastClosedOrderIDs(ctx sdk.Context, orderIDs []string) {
	store := ctx.KVStore(k.orderStoreKey)
	if len(orderIDs) == 0 {
		store.Delete(types.RecentlyClosedOrderIDsKey)
	}
	store.Set(types.RecentlyClosedOrderIDsKey, k.cdc.MustMarshalJSON(orderIDs)) //recentlyClosedOrderIDs
}

// SetOrderIDs sets OrderIDs to diskCache
func (k Keeper) SetOrderIDs(key string, orderIDs []string) {
	k.diskCache.setOrderIDs(key, orderIDs)
}

// GetProductsFromDepthBookMap gets products from DepthBookMap in diskCache
func (k Keeper) GetProductsFromDepthBookMap() []string {
	return k.diskCache.getProductsFromDepthBookMap()
}
