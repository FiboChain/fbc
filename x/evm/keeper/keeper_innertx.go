package keeper

func initInnerDB() error {
	return nil
}

type BlockInnerData = interface{}

func defaultBlockInnerData() BlockInnerData {
	return nil
}

// InitInnerBlock init inner block data
func (k *Keeper) InitInnerBlock(...interface{}) {}

func (k *Keeper) UpdateInnerBlockData(...interface{}) {}

// AddInnerTx add inner tx
func (k *Keeper) AddInnerTx(...interface{}) {}

// AddContract add erc20 contract
func (k *Keeper) AddContract(...interface{}) {}

func (k *Keeper) UpdateInnerTx(...interface{}) {
}

// DeleteInnerTx delete inner tx
func (k *Keeper) DeleteInnerTx(...interface{}) {}

func (k *Keeper) UpdateWasmInnerTx(...interface{}) {

}
