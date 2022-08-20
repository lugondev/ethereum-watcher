package main

import (
	"context"
	"ethereum-watcher"
	"ethereum-watcher/rpc"
	"ethereum-watcher/utils"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
	"math/big"
	"os"
	"os/signal"
)

var api string
var verbosity uint32
var jsonLogFormat bool
var contractAddr string
var tokenAddr string
var txHash string
var eventSigs []string
var blockBackoff int

func main() {
	rootCMD.AddCommand(blockNumCMD)
	rootCMD.PersistentFlags().Uint32Var(&verbosity, "verbosity", 4, "Logging verbosity: 0=panic, 1=fatal, 2=error, 3=warning, 4=info, 5=debug, 5=trace")
	rootCMD.PersistentFlags().BoolVar(&jsonLogFormat, "json-log", false, "Format logs with JSON")
	rootCMD.PersistentFlags().StringVarP(&api, "rpc", "r", "https://bsc-testnet.nodereal.io/v1/f62bd255a11145dfbc560565c1ad47c9", "RPC url")
	_ = rootCMD.MarkPersistentFlagRequired("rpc")

	checkTxCMD.Flags().StringVar(&txHash, "hash", "", "Hash of transaction")
	_ = checkTxCMD.MarkFlagRequired("hash")

	tokenTransferCMD.Flags().StringVar(&tokenAddr, "token", "", "token address listen")
	_ = tokenTransferCMD.MarkFlagRequired("token")

	contractEventListenerCMD.Flags().StringVarP(&contractAddr, "contract", "c", "", "contract address listen to")
	_ = contractEventListenerCMD.MarkFlagRequired("contract")
	contractEventListenerCMD.Flags().StringArrayVarP(&eventSigs, "events", "e", []string{}, "signatures of events we are interested in")
	_ = contractEventListenerCMD.MarkFlagRequired("events")
	contractEventListenerCMD.Flags().IntVar(&blockBackoff, "block-backoff", 0, "how many blocks we go back")

	rootCMD.AddCommand(tokenTransferCMD)
	rootCMD.AddCommand(contractEventListenerCMD)
	rootCMD.AddCommand(checkTxCMD)

	if err := rootCMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

var rootCMD = &cobra.Command{
	Use:   "ethereum-watcher",
	Short: "ethereum-watcher makes getting updates from Ethereum easier",
}

var checkTxCMD = &cobra.Command{
	Use:   "check-tx",
	Short: "Check data tx by hash",
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLogger(verbosity, jsonLogFormat)

		rpcWithRetry := rpc.NewEthRPCWithRetry(api, 3)
		transactionReceipt, err := rpcWithRetry.GetTransactionReceipt(txHash)
		transactionByHash, err := rpcWithRetry.GetTransactionByHash(txHash)

		if err != nil {
			panic(err)
		}
		if transactionReceipt.Status == 0 {
			utils.Infof("transaction status: fail")
		} else {
			utils.Infof("transaction status: success")
		}
		msg, err := transactionByHash.AsMessage(types.NewEIP155Signer(transactionByHash.ChainId()), nil)
		if err != nil {
			utils.Errorln("Cannot get From address")
		}
		utils.Infof("from: %v to %v", msg.From().String(), transactionByHash.To().String())
		utils.Infof("data: %v", common.Bytes2Hex(transactionByHash.Data()))

		fee := big.NewInt(0).Mul(transactionByHash.GasPrice(), big.NewInt(int64(transactionReceipt.GasUsed)))
		utils.Infof("value: %v with fee %v", transactionByHash.Value(), fee.Uint64())
		utils.Infof("gas limit: %v gas used %d", transactionByHash.Gas(), transactionReceipt.GasUsed)
	},
}

var blockNumCMD = &cobra.Command{
	Use:   "new-block-number",
	Short: "Print number of new block",
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLogger(verbosity, jsonLogFormat)

		ctx, cancel := context.WithCancel(context.Background())

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		w := ethereum_watcher.NewHttpBasedEthWatcher(ctx, api)

		go func() {
			<-c
			cancel()
		}()

		err := w.RunTillExit()
		if err != nil {
			utils.Printf("exit with err: %s", err)
		} else {
			utils.Infoln("exit")
		}
	},
}

var tokenTransferCMD = &cobra.Command{
	Use:   "token-transfer",
	Short: "Show Transfer Event of Token",
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLogger(verbosity, jsonLogFormat)
		// Transfer
		topicsInterestedIn := []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}

		handler := func(from, to int, receiptLogs []*types.Log, isUpToHighestBlock bool) error {

			if from != to {
				utils.Infof("See new Transfer at blockRange: %d -> %d, count: %2d", from, to, len(receiptLogs))
			} else {
				utils.Infof("See new Transfer at block: %d, count: %2d", from, len(receiptLogs))
			}

			for _, log := range receiptLogs {
				utils.Infof("  >> tx: %s", log.TxHash.String())
			}

			fmt.Println("  ")

			return nil
		}

		receiptLogWatcher := ethereum_watcher.NewReceiptLogWatcher(
			context.TODO(),
			api,
			-1,
			tokenAddr,
			topicsInterestedIn,
			handler,
			ethereum_watcher.ReceiptLogWatcherConfig{
				StepSizeForBigLag:               5,
				IntervalForPollingNewBlockInSec: 5,
				RPCMaxRetry:                     3,
				ReturnForBlockWithNoReceiptLog:  true,
			},
		)

		err := receiptLogWatcher.Run()
		if err != nil {
			panic(err)
		}
	},
}

var contractEventListenerCMD = &cobra.Command{
	Use:   "contract-event-listener",
	Short: "listen and print events from contract",
	Example: `
	listen to Transfer & Approve events from Multi-Collateral-DAI in Ethereum
	
	./bin/ethereum-watcher contract-event-listener \
	--rpc {eth} \
	--block-backoff 100 \
	--contract 0x6b175474e89094c44da98b954eedeac495271d0f \
	--events 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.SetLogger(verbosity, jsonLogFormat)

		handler := func(from, to int, receiptLogs []*types.Log, isUpToHighestBlock bool) error {

			if from != to {
				utils.Infof("# of interested events at block(%d->%d): %d", from, to, len(receiptLogs))
			} else {
				utils.Infof("# of interested events at block(%d): %d", from, len(receiptLogs))
			}

			for _, log := range receiptLogs {
				utils.Infof("  >> tx: https://etherscan.io/tx/%s", log.TxHash.String())
			}

			fmt.Println("  ")

			return nil
		}

		startBlockNum := -1
		if blockBackoff > 0 {
			rpcWithRetry := rpc.NewEthRPCWithRetry(api, 3)
			curBlockNum, err := rpcWithRetry.GetCurrentBlockNum()
			if err == nil {
				startBlockNum = int(curBlockNum) - blockBackoff

				if startBlockNum > 0 {
					utils.Infof("--block-backoff activated, we start from block: %d (= %d - %d)",
						startBlockNum, curBlockNum, blockBackoff)
				}
			}
		}

		fmt.Println("eventSigs:", eventSigs)
		receiptLogWatcher := ethereum_watcher.NewReceiptLogWatcher(
			context.TODO(),
			api,
			startBlockNum,
			contractAddr,
			eventSigs,
			handler,
			ethereum_watcher.ReceiptLogWatcherConfig{
				StepSizeForBigLag:               5,
				IntervalForPollingNewBlockInSec: 5,
				RPCMaxRetry:                     3,
				ReturnForBlockWithNoReceiptLog:  true,
			},
		)

		err := receiptLogWatcher.Run()
		if err != nil {
			panic(err)
		}
	},
}
