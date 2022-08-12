package ethereum_watcher

import (
	"context"
	"ethereum-watcher/plugin"
	"ethereum-watcher/structs"
	"ethereum-watcher/utils"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
			fmt.Println("Removed >>", txAndReceipt.Tx.Hash(), txAndReceipt.Receipt.TransactionIndex)
		} else {
			fmt.Println("Adding >>", txAndReceipt.Tx.Hash(), txAndReceipt.Receipt.TransactionIndex)
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

func TestFilterPluginForDyDxApprove(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	callback := func(txAndReceipt *structs.RemovableTxAndReceipt) {
		receipt := txAndReceipt.Receipt

		for _, receiptLog := range receipt.Logs {
			topics := receiptLog.Topics
			if len(topics) == 3 &&
				topics[0] == common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925") &&
				topics[2] == common.HexToHash("0x0000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e") {
				fmt.Printf(">> approving to dydx, tx: %s\n", txAndReceipt.Tx.Hash())
			}
		}
	}

	// only accept txs which send to DAI
	filterFunc := func(tx *types.Transaction) bool {
		addr := common.HexToAddress("0x89d24a6b4ccb1b6faa2625fe562bdd9a23260359")
		return tx.To() == &addr
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
	utils.Infof("tx hash is status: %v", transactionByHash.Status)
	fmt.Println(transactionByHash)
}
