package ethereum_watcher

import (
	"context"
	"ethereum-watcher/plugin"
	"ethereum-watcher/structs"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNewBlockNumPlugin(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)

	w := NewHttpBasedEthWatcher(context.Background(), api)

	logrus.Println("waiting for new block...")
	w.RegisterBlockPlugin(plugin.NewBlockNumPlugin(func(i uint64, b bool) {
		logrus.Printf(">> found new block: %d, is removed: %t", i, b)
	}))

	_ = w.RunTillExit()
}

func TestSimpleBlockPlugin(t *testing.T) {
	w := NewHttpBasedEthWatcher(context.Background(), api)

	w.RegisterBlockPlugin(plugin.NewSimpleBlockPlugin(func(block *structs.RemovableBlock) {
		logrus.Infof(">> %+v", block.Block)
	}))

	_ = w.RunTillExit()
}
