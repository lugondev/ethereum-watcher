package ethereum_watcher

import (
	"context"
	"ethereum-watcher/plugin"
	"ethereum-watcher/structs"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestTxHashPlugin(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxPlugin(plugin.NewTxHashPlugin(func(txHash string, isRemoved bool) {
		fmt.Println(">>", txHash, isRemoved)
	}))

	_ = w.RunTillExit()
}

func TestTxPlugin(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxPlugin(plugin.NewTxPlugin(func(tx structs.RemovableTx) {
		logrus.Printf(">> block: %d, txHash: %s", tx.GetBlockNumber(), tx.GetHash())
	}))

	_ = w.RunTillExit()
}
