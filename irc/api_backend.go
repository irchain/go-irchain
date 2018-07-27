// Copyright 2015 The go-irchain Authors
// This file is part of the go-irchain library.
//
// The go-irchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-irchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-irchain library. If not, see <http://www.gnu.org/licenses/>.

package irc

import (
	"context"
	"math/big"

	"github.com/irchain/go-irchain/accounts"
	"github.com/irchain/go-irchain/common"
	"github.com/irchain/go-irchain/common/math"
	"github.com/irchain/go-irchain/core"
	"github.com/irchain/go-irchain/core/bloombits"
	"github.com/irchain/go-irchain/core/rawdb"
	"github.com/irchain/go-irchain/core/state"
	"github.com/irchain/go-irchain/core/types"
	"github.com/irchain/go-irchain/core/vm"
	"github.com/irchain/go-irchain/event"
	"github.com/irchain/go-irchain/irc/downloader"
	"github.com/irchain/go-irchain/irc/gasprice"
	"github.com/irchain/go-irchain/ircdb"
	"github.com/irchain/go-irchain/params"
	"github.com/irchain/go-irchain/rpc"
)

// IrcApiBackend implements ircapi.Backend for full nodes
type IrcApiBackend struct {
	irc *IrChain
	gpo *gasprice.Oracle
}

func (b *IrcApiBackend) ChainConfig() *params.ChainConfig {
	return b.irc.chainConfig
}

func (b *IrcApiBackend) CurrentBlock() *types.Block {
	return b.irc.blockchain.CurrentBlock()
}

func (b *IrcApiBackend) SetHead(number uint64) {
	b.irc.protocolManager.downloader.Cancel()
	b.irc.blockchain.SetHead(number)
}

func (b *IrcApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.irc.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.irc.blockchain.CurrentBlock().Header(), nil
	}
	return b.irc.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *IrcApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.irc.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.irc.blockchain.CurrentBlock(), nil
	}
	return b.irc.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *IrcApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.irc.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.irc.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *IrcApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.irc.blockchain.GetBlockByHash(blockHash), nil
}

func (b *IrcApiBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.irc.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.irc.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *IrcApiBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.irc.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.irc.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *IrcApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.irc.blockchain.GetTdByHash(blockHash)
}

func (b *IrcApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(*msg.To(), math.MaxBig256)
	vmError := func() error { return nil }
	context := core.NewEVMContext(msg, header, b.irc.BlockChain(), nil)
	return vm.NewEVM(context, state, b.irc.chainConfig, vmCfg), vmError, nil
}

func (b *IrcApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.irc.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *IrcApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.irc.BlockChain().SubscribeChainEvent(ch)
}

func (b *IrcApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.irc.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *IrcApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.irc.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *IrcApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.irc.BlockChain().SubscribeLogsEvent(ch)
}

func (b *IrcApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.irc.txPool.AddLocal(signedTx)
}

func (b *IrcApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.irc.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *IrcApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.irc.txPool.Get(hash)
}

func (b *IrcApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.irc.txPool.State().GetNonce(addr), nil
}

func (b *IrcApiBackend) Stats() (pending int, queued int) {
	return b.irc.txPool.Stats()
}

func (b *IrcApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.irc.TxPool().Content()
}

func (b *IrcApiBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.irc.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *IrcApiBackend) Downloader() *downloader.Downloader {
	return b.irc.Downloader()
}

func (b *IrcApiBackend) ProtocolVersion() int {
	return b.irc.IrcVersion()
}

func (b *IrcApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *IrcApiBackend) ChainDb() ircdb.Database {
	return b.irc.ChainDb()
}

func (b *IrcApiBackend) EventMux() *event.TypeMux {
	return b.irc.EventMux()
}

func (b *IrcApiBackend) AccountManager() *accounts.Manager {
	return b.irc.AccountManager()
}

func (b *IrcApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.irc.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *IrcApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.irc.bloomRequests)
	}
}
