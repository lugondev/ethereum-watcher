package rpc

import (
	"github.com/ethereum/go-ethereum/core/types"
	"time"
)

type EthBlockChainRPCWithRetry struct {
	*EthBlockChainRPC
	maxRetryTimes int
}

func NewEthRPCWithRetry(api string, maxRetryCount int) *EthBlockChainRPCWithRetry {
	rpc := NewEthRPC(api)

	return &EthBlockChainRPCWithRetry{rpc, maxRetryCount}
}

func (rpc EthBlockChainRPCWithRetry) GetBlockByNum(num uint64) (rst *types.Block, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetBlockByNum(num)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}

func (rpc EthBlockChainRPCWithRetry) GetTransactionReceipt(txHash string) (rst *types.Receipt, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetTransactionReceipt(txHash)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}

func (rpc EthBlockChainRPCWithRetry) EthGetTransactionByHash(txHash string) (rst *types.Transaction, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetTransactionByHash(txHash)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}

func (rpc EthBlockChainRPCWithRetry) GetCurrentBlockNum() (rst uint64, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetCurrentBlockNum()
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}
func (rpc EthBlockChainRPCWithRetry) GetLogs(
	fromBlockNum, toBlockNum uint64,
	address string,
	topics []string,
) (rst []*types.Log, err error) {
	for i := 0; i <= rpc.maxRetryTimes; i++ {
		rst, err = rpc.EthBlockChainRPC.GetLogs(fromBlockNum, toBlockNum, address, topics)
		if err == nil {
			break
		} else {
			time.Sleep(time.Duration(500*(i+1)) * time.Millisecond)
		}
	}

	return
}
