package main

import (
	"context"
	"ethereum-watcher"
	"ethereum-watcher/blockchain"
	"ethereum-watcher/plugin"
	"ethereum-watcher/rpc"
	"ethereum-watcher/utils"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

var api string
var contractAddr string
var tokenAddr string
var eventSigs []string
var blockBackoff int

func main() {
	utils.SetLogger(uint32(logrus.DebugLevel), false)

	rootCMD.AddCommand(blockNumCMD)
	rootCMD.AddCommand(tokenTransferCMD)
	rootCMD.Flags().StringVar(&api, "rpc", "https://bsc-testnet.nodereal.io/v1/f62bd255a11145dfbc560565c1ad47c9", "RPC url")
	_ = rootCMD.MarkFlagRequired("rpc")

	tokenTransferCMD.Flags().StringVar(&tokenAddr, "token", "", "token address listen")
	_ = tokenTransferCMD.MarkFlagRequired("token")

	contractEventListenerCMD.Flags().StringVarP(&contractAddr, "contract", "c", "", "contract address listen to")
	_ = contractEventListenerCMD.MarkFlagRequired("contract")
	contractEventListenerCMD.Flags().StringArrayVarP(&eventSigs, "events", "e", []string{}, "signatures of events we are interested in")
	_ = contractEventListenerCMD.MarkFlagRequired("events")
	contractEventListenerCMD.Flags().IntVar(&blockBackoff, "block-backoff", 0, "how many blocks we go back")
	rootCMD.AddCommand(contractEventListenerCMD)

	if err := rootCMD.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCMD = &cobra.Command{
	Use:   "ethereum-watcher",
	Short: "ethereum-watcher makes getting updates from Ethereum easier",
}

var blockNumCMD = &cobra.Command{
	Use:   "new-block-number",
	Short: "Print number of new block",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		w := ethereum_watcher.NewHttpBasedEthWatcher(ctx, api)

		logrus.Println("waiting for new block...")
		w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
			logrus.Printf(">> found new block: %d, is removed: %t", i, b)
		}))

		go func() {
			<-c
			cancel()
		}()

		err := w.RunTillExit()
		if err != nil {
			logrus.Printf("exit with err: %s", err)
		} else {
			logrus.Infoln("exit")
		}
	},
}

var tokenTransferCMD = &cobra.Command{
	Use:   "token-transfer",
	Short: "Show Transfer Event of Token",
	Run: func(cmd *cobra.Command, args []string) {
		// Transfer
		topicsInterestedIn := []string{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"}

		handler := func(from, to int, receiptLogs []blockchain.IReceiptLog, isUpToHighestBlock bool) error {

			if from != to {
				logrus.Infof("See new USDT Transfer at blockRange: %d -> %d, count: %2d", from, to, len(receiptLogs))
			} else {
				logrus.Infof("See new USDT Transfer at block: %d, count: %2d", from, len(receiptLogs))
			}

			for _, log := range receiptLogs {
				logrus.Infof("  >> tx: https://etherscan.io/tx/%s", log.GetTransactionHash())
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
  listen to Transfer & Approve events from Multi-Collateral-DAI
  
  /bin/ethereum-watcher contract-event-listener \
    --block-backoff 100
    --contract 0x6b175474e89094c44da98b954eedeac495271d0f \
    --events 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef`,
	Run: func(cmd *cobra.Command, args []string) {

		handler := func(from, to int, receiptLogs []blockchain.IReceiptLog, isUpToHighestBlock bool) error {

			if from != to {
				logrus.Infof("# of interested events at block(%d->%d): %d", from, to, len(receiptLogs))
			} else {
				logrus.Infof("# of interested events at block(%d): %d", from, len(receiptLogs))
			}

			for _, log := range receiptLogs {
				logrus.Infof("  >> tx: https://etherscan.io/tx/%s", log.GetTransactionHash())
			}

			fmt.Println("  ")

			return nil
		}

		startBlockNum := -1
		if blockBackoff > 0 {
			rpc := rpc.NewEthRPCWithRetry(api, 3)
			curBlockNum, err := rpc.GetCurrentBlockNum()
			if err == nil {
				startBlockNum = int(curBlockNum) - blockBackoff

				if startBlockNum > 0 {
					logrus.Infof("--block-backoff activated, we start from block: %d (= %d - %d)",
						startBlockNum, curBlockNum, blockBackoff)
				}
			}
		}

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
