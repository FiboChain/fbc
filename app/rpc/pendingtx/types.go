package pendingtx

import (
	"github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/evm/watcher"
)

type PendingMsg struct {
	Topic  string      `json:"topic"`
	Source interface{} `json:"source"`
	// not use interface for fast json
	Data *watcher.Transaction `json:"data"`
}

type RmPendingMsg struct {
	Topic  string      `json:"topic"`
	Source interface{} `json:"source"`
	// not use interface for fast json
	Data *RmPendingTx `json:"data"`
}

type RmPendingTx struct {
	From   string                  `json:"from"`
	Hash   string                  `json:"hash"`
	Nonce  string                  `json:"nonce"`
	Delete bool                    `json:"delete"`
	Reason types.RmPendingTxReason `json:"reason"`
}
