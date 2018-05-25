// Copyright 2015 The happyuc-go Authors
// This file is part of the happyuc-go library.
//
// The happyuc-go library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The happyuc-go library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the happyuc-go library. If not, see <http://www.gnu.org/licenses/>.

package huc

import (
	"context"
	"math/big"

	"github.com/happyuc-project/happyuc-go/accounts"
	"github.com/happyuc-project/happyuc-go/common"
	"github.com/happyuc-project/happyuc-go/common/math"
	"github.com/happyuc-project/happyuc-go/core"
	"github.com/happyuc-project/happyuc-go/core/bloombits"
	"github.com/happyuc-project/happyuc-go/core/state"
	"github.com/happyuc-project/happyuc-go/core/types"
	"github.com/happyuc-project/happyuc-go/core/vm"
	"github.com/happyuc-project/happyuc-go/event"
	"github.com/happyuc-project/happyuc-go/huc/downloader"
	"github.com/happyuc-project/happyuc-go/huc/gasprice"
	"github.com/happyuc-project/happyuc-go/hucdb"
	"github.com/happyuc-project/happyuc-go/params"
	"github.com/happyuc-project/happyuc-go/rpc"
)

// HucApiBackend implements hucapi.Backend for full nodes
type HucApiBackend struct {
	huc *HappyUC
	gpo *gasprice.Oracle
}

func (b *HucApiBackend) ChainConfig() *params.ChainConfig {
	return b.huc.chainConfig
}

func (b *HucApiBackend) CurrentBlock() *types.Block {
	return b.huc.blockchain.CurrentBlock()
}

func (b *HucApiBackend) SetHead(number uint64) {
	b.huc.protocolManager.downloader.Cancel()
	b.huc.blockchain.SetHead(number)
}

func (b *HucApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.huc.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.huc.blockchain.CurrentBlock().Header(), nil
	}
	return b.huc.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *HucApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.huc.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.huc.blockchain.CurrentBlock(), nil
	}
	return b.huc.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *HucApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.huc.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.huc.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *HucApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.huc.blockchain.GetBlockByHash(blockHash), nil
}

func (b *HucApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.huc.chainDb, blockHash, core.GetBlockNumber(b.huc.chainDb, blockHash)), nil
}

func (b *HucApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.huc.chainDb, blockHash, core.GetBlockNumber(b.huc.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *HucApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.huc.blockchain.GetTdByHash(blockHash)
}

func (b *HucApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.huc.BlockChain(), nil)
	return vm.NewEVM(context, state, b.huc.chainConfig, vmCfg), vmError, nil
}

func (b *HucApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.huc.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *HucApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.huc.BlockChain().SubscribeChainEvent(ch)
}

func (b *HucApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.huc.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *HucApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.huc.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *HucApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.huc.BlockChain().SubscribeLogsEvent(ch)
}

func (b *HucApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.huc.txPool.AddLocal(signedTx)
}

func (b *HucApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.huc.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *HucApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.huc.txPool.Get(hash)
}

func (b *HucApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.huc.txPool.State().GetNonce(addr), nil
}

func (b *HucApiBackend) Stats() (pending int, queued int) {
	return b.huc.txPool.Stats()
}

func (b *HucApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.huc.TxPool().Content()
}

func (b *HucApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.huc.TxPool().SubscribeTxPreEvent(ch)
}

func (b *HucApiBackend) Downloader() *downloader.Downloader {
	return b.huc.Downloader()
}

func (b *HucApiBackend) ProtocolVersion() int {
	return b.huc.HucVersion()
}

func (b *HucApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *HucApiBackend) ChainDb() hucdb.Database {
	return b.huc.ChainDb()
}

func (b *HucApiBackend) EventMux() *event.TypeMux {
	return b.huc.EventMux()
}

func (b *HucApiBackend) AccountManager() *accounts.Manager {
	return b.huc.AccountManager()
}

func (b *HucApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.huc.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *HucApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.huc.bloomRequests)
	}
}
