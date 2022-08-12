package ethereum_watcher

import (
	"context"
	"ethereum-watcher/plugin"
	"ethereum-watcher/structs"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"testing"
)

//var api = "https://bsc-testnet.nodereal.io/v1/f62bd255a11145dfbc560565c1ad47c9"
var api = "https://bsc-mainnet.nodereal.io/v1/100eacb4d89e4df3a73cae315b777652"

// todo why some tx index in block is zero?
func TestTxReceiptPlugin(t *testing.T) {
	log.SetLevel(log.DEBUG)

	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(txAndReceipt *structs.RemovableTxAndReceipt) {
		if txAndReceipt.IsRemoved {
			fmt.Println("Removed >>", txAndReceipt.Tx.GetHash(), txAndReceipt.Receipt.GetTxIndex())
		} else {
			fmt.Println("Adding >>", txAndReceipt.Tx.GetHash(), txAndReceipt.Receipt.GetTxIndex())
		}
	}))

	_ = w.RunTillExit()
}

func TestErc20TransferPlugin(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterTxReceiptPlugin(plugin.NewERC20TransferPlugin(
		func(token, from, to string, amount decimal.Decimal, isRemove bool) {

			logrus.Infof("New ERC20 Transfer >> token(%s), %s -> %s, amount: %s, isRemoved: %t",
				token, from, to, amount, isRemove)

		},
	))

	_ = w.RunTillExit()
}

func TestFilterPlugin(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		fmt.Println("tx:", txAndReceipt.Tx.GetHash())
	}

	// only accept txs which end with: f
	filterFunc := func(tx types.Transaction) bool {
		txHash := tx.GetHash()

		return txHash[len(txHash)-1:] == "f"
	}

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPluginWithFilter(callback, filterFunc))

	err := w.RunTillExitFromBlock(7840000)
	if err != nil {
		fmt.Println("RunTillExit with err:", err)
	}
}

func TestFilterPluginForDyDxApprove(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		receipt := txAndReceipt.Receipt

		for _, receiptLog := range receipt.GetLogs() {
			topics := receiptLog.GetTopics()
			if len(topics) == 3 &&
				topics[0] == "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925" &&
				topics[2] == "0x0000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e" {
				fmt.Printf(">> approving to dydx, tx: %s\n", txAndReceipt.Tx.GetHash())
			}
		}
	}

	// only accept txs which send to DAI
	filterFunc := func(tx types.Transaction) bool {
		return tx.GetTo() == "0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359"
	}

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPluginWithFilter(callback, filterFunc))

	err := w.RunTillExitFromBlock(7844853)
	if err != nil {
		fmt.Println("RunTillExit with err:", err)
	}
}

func TestGetTxByHash(t *testing.T) {
	client, err := ethclient.Dial(api)
	if err != nil {
		panic(err)
	}
	number, _ := client.BlockNumber(context.Background())
	logrus.Infof("blocknumber: %d", number)
	transactionByHash, err := client.TransactionReceipt(context.Background(), common.HexToHash("0xe16122e6ca4ab8312a651dd1ff225b1d9b3391b15f9374687e1845e8f360fb9a"))
	if err != nil {
		panic(err)
	}
	//logrus.Infof("tx hash is pending: %v", isPending)
	fmt.Println(transactionByHash)
}
