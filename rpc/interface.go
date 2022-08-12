package rpc

import "github.com/ethereum/go-ethereum/core/types"

type IBlockChainRPC interface {
	GetCurrentBlockNum() (uint64, error)

	GetBlockByNum(uint64) (types.Block, error)
	GetTransactionReceipt(txHash string) (types.Receipt, error)

	GetLogs(from, to uint64, address string, topics []string) ([]types.Log, error)
}
