package ethereum_watcher

import (
	"context"
	"ethereum-watcher/plugin"
	"ethereum-watcher/structs"
	"ethereum-watcher/utils"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestReceiptLogsPlugin(t *testing.T) {
	utils.SetLogger(uint32(logrus.DebugLevel), false)

	w := NewHttpBasedEthWatcher(context.Background(), api)

	contract := "0x63bB8a255a8c045122EFf28B3093Cc225B711F6D"
	// Match
	topics := []string{"0x6bf96fcc2cec9e08b082506ebbc10114578a497ff1ea436628ba8996b750677c"}

	w.RegisterReceiptLogPlugin(plugin.NewReceiptLogPlugin(contract, topics, func(receipt *structs.RemovableReceiptLog) {
		if receipt.IsRemoved {
			logrus.Infof("Removed >> %+v", receipt)
		} else {
			logrus.Infof("Adding >> %+v, tx: %s, logIdx: %d", receipt, receipt.Log.TxHash.String(), receipt.Log.Index)
		}
	}))

	//startBlock := 12304546
	startBlock := 12101723
	err := w.RunTillExitFromBlock(uint64(startBlock))

	fmt.Println("err:", err)
}
