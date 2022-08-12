package blockchain

import (
	"context"
	"errors"
	"ethereum-watcher/utils"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
)

type BlockChain interface {
	GetTokenBalance(tokenAddress, address string) decimal.Decimal
	GetTokenAllowance(tokenAddress, proxyAddress, address string) decimal.Decimal

	GetBlockNumber() (uint64, error)
	GetBlockByNumber(blockNumber uint64) (types.Block, error)

	GetTransaction(ID string) (types.Transaction, error)
	GetTransactionReceipt(ID string) (types.Receipt, error)
	GetTransactionAndReceipt(ID string) (types.Transaction, types.Receipt, error)
}

type IReceiptLog interface {
	GetRemoved() bool
	GetLogIndex() int
	GetTransactionIndex() int
	GetTransactionHash() string
	GetBlockNum() int
	GetBlockHash() string
	GetAddress() string
	GetData() string
	GetTopics() []string
}

type EthereumBlock struct {
	*types.Block
}

func (block *EthereumBlock) GetTransactions() []types.Transaction {
	txs := make([]types.Transaction, 0, 20)

	for i := range block.Transactions() {
		tx := block.Transactions()[i]
		txs = append(txs, *tx)
	}

	return txs
}

type EthereumTransaction struct {
	*types.Transaction
}

type EthereumTransactionReceipt struct {
	*types.Receipt
}

func (r *EthereumTransactionReceipt) GetLogs() (rst []*types.Log) {
	for i := range r.Logs {
		rst = append(rst, r.Logs[i])
	}

	return
}

type ReceiptLog struct {
	*types.Log
}

type Ethereum struct {
	client *ethclient.Client
}

func (e *Ethereum) GetBlockByNumber(number uint64) (*types.Block, error) {

	block, err := e.client.BlockByNumber(context.Background(), big.NewInt(int64(number)))

	if err != nil {
		log.Errorf("get Block by Number failed %+v", err)
		return nil, err
	}

	if block == nil {
		log.Errorf("get Block by Number returns nil block for num: %d", number)
		return nil, errors.New("get Block by Number returns nil block for num: " + strconv.Itoa(int(number)))
	}

	return block, nil
}

func (e *Ethereum) GetBlockNumber() (uint64, error) {
	number, err := e.client.BlockNumber(context.Background())

	if err != nil {
		log.Errorf("GetBlockNumber failed, %v", err)
		return 0, err
	}

	return number, nil
}

func (e *Ethereum) GetTransaction(ID string) (*types.Transaction, error) {
	tx, _, err := e.client.TransactionByHash(context.Background(), common.HexToHash(ID))

	if err != nil {
		log.Errorf("GetTransaction failed, %v", err)
		return nil, err
	}

	return tx, nil
}

func (e *Ethereum) GetTransactionReceipt(ID string) (*types.Receipt, error) {
	txReceipt, err := e.client.TransactionReceipt(context.Background(), common.HexToHash(ID))

	if err != nil {
		log.Errorf("GetTransactionReceipt failed, %v", err)
		return nil, err
	}

	return txReceipt, nil
}

func (e *Ethereum) GetTransactionAndReceipt(ID string) (*types.Transaction, *types.Receipt, error) {
	txReceiptChannel := make(chan *types.Receipt)

	go func() {
		rec, _ := e.GetTransactionReceipt(ID)
		txReceiptChannel <- rec
	}()

	txInfoChannel := make(chan *types.Transaction)
	go func() {
		tx, _ := e.GetTransaction(ID)
		txInfoChannel <- tx
	}()

	return <-txInfoChannel, <-txReceiptChannel, nil
}

func (e *Ethereum) GetTokenBalance(tokenAddress, address string) decimal.Decimal {
	toContract := common.HexToAddress(tokenAddress)
	res, err := e.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &toContract,
		From: common.HexToAddress(address),
		Data: []byte(fmt.Sprintf("0x70a08231000000000000000000000000%s", without0xPrefix(address))),
	}, nil)

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(string(res))
}

func without0xPrefix(address string) string {
	if address[:2] == "0x" {
		address = address[2:]
	}

	return address
}

func (e *Ethereum) GetTokenAllowance(tokenAddress, proxyAddress, address string) decimal.Decimal {
	toContract := common.HexToAddress(tokenAddress)
	res, err := e.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &toContract,
		From: common.HexToAddress(address),
		Data: []byte(fmt.Sprintf("0xdd62ed3e000000000000000000000000%s000000000000000000000000%s", without0xPrefix(address), without0xPrefix(proxyAddress))),
	}, nil)

	if err != nil {
		panic(err)
	}

	return utils.StringToDecimal(string(res))
}

func (e *Ethereum) GetNonce(address string) (int, error) {
	nonceAt, err := e.client.NonceAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return 0, err
	}

	return int(nonceAt), err
}
