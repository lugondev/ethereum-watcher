package rpc

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"math/big"
)

type EthBlockChainRPC struct {
	rpcImpl *ethclient.Client
}

func NewEthRPC(api string) *EthBlockChainRPC {
	client, err := ethclient.Dial(api)
	if err != nil {
		panic(err)
	}

	return &EthBlockChainRPC{client}
}

func (rpc EthBlockChainRPC) GetBlockByNum(num uint64) (*types.Block, error) {
	block, err := rpc.rpcImpl.BlockByNumber(context.Background(), big.NewInt(int64(num)))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, errors.New("nil block")
	}

	return block, err
}

func (rpc EthBlockChainRPC) GetTransactionReceipt(txHash string) (*types.Receipt, error) {
	receipt, err := rpc.rpcImpl.TransactionReceipt(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}
	if receipt == nil {
		return nil, errors.New("nil receipt")
	}

	return receipt, err
}

func (rpc EthBlockChainRPC) GetTransactionByHash(txHash string) (*types.Transaction, error) {
	transaction, _, err := rpc.rpcImpl.TransactionByHash(context.Background(), common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, errors.New("nil transaction")
	}
	return transaction, err
}

func (rpc EthBlockChainRPC) GetCurrentBlockNum() (uint64, error) {
	num, err := rpc.rpcImpl.BlockNumber(context.Background())
	return num, err
}

func (rpc EthBlockChainRPC) GetLogs(
	fromBlockNum, toBlockNum uint64,
	address string,
	topics []string,
) ([]*types.Log, error) {

	filterParam := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlockNum)),
		ToBlock:   big.NewInt(int64(toBlockNum)),
		Addresses: []common.Address{common.HexToAddress(address)},
		Topics: [][]common.Hash{
			funk.Map(topics, func(topic string) common.Hash {
				return common.HexToHash(topic)
			}).([]common.Hash),
		},
	}

	logs, err := rpc.rpcImpl.FilterLogs(context.Background(), filterParam)
	if err != nil {
		logrus.Warnf("EthGetLogs err: %s, params: %+v", err, filterParam)
		return nil, err
	}

	logrus.Debugf("EthGetLogs logs count at block(%d - %d): %d", fromBlockNum, toBlockNum, len(logs))

	var result []*types.Log
	for i := 0; i < len(logs); i++ {
		l := logs[i]

		logrus.Debugf("EthGetLogs receipt log: %+v", l)

		result = append(result, &l)
	}

	return result, err
}
