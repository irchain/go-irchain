// Copyright 2016 The happyuc-go Authors
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

package les

import (
	"context"
	"math/big"

	"github.com/happyuc-project/happyuc-go/accounts"
	"github.com/happyuc-project/happyuc-go/common"
	"github.com/happyuc-project/happyuc-go/common/math"
	"github.com/happyuc-project/happyuc-go/core"
	"github.com/happyuc-project/happyuc-go/core/bloombits"
	"github.com/happyuc-project/happyuc-go/core/rawdb"
	"github.com/happyuc-project/happyuc-go/core/state"
	"github.com/happyuc-project/happyuc-go/core/types"
	"github.com/happyuc-project/happyuc-go/core/vm"
	"github.com/happyuc-project/happyuc-go/event"
	"github.com/happyuc-project/happyuc-go/huc/downloader"
	"github.com/happyuc-project/happyuc-go/huc/gasprice"
	"github.com/happyuc-project/happyuc-go/hucdb"
	"github.com/happyuc-project/happyuc-go/light"
	"github.com/happyuc-project/happyuc-go/params"
	"github.com/happyuc-project/happyuc-go/rpc"
)

type LesApiBackend struct {
	huc *LightHappyUC
	gpo *gasprice.Oracle
}

func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
	return b.huc.chainConfig
}

func (b *LesApiBackend) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(b.huc.BlockChain().CurrentHeader())
}

func (b *LesApiBackend) SetHead(number uint64) {
	b.huc.protocolManager.downloader.Cancel()
	b.huc.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
		return b.huc.blockchain.CurrentHeader(), nil
	}

	return b.huc.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, err
	}
	return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	return light.NewState(ctx, header, b.huc.odr), header, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.huc.blockchain.GetBlockByHash(ctx, blockHash)
}

func (b *LesApiBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.huc.chainDb, hash); number != nil {
		return light.GetBlockReceipts(ctx, b.huc.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	if number := rawdb.ReadHeaderNumber(b.huc.chainDb, hash); number != nil {
		return light.GetBlockLogs(ctx, b.huc.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetTd(hash common.Hash) *big.Int {
	return b.huc.blockchain.GetTdByHash(hash)
}

func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(*msg.To(), math.MaxBig256)
	context := core.NewEVMContext(msg, header, b.huc.blockchain, nil)
	return vm.NewEVM(context, state, b.huc.chainConfig, vmCfg), state.Error, nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.huc.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.huc.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (types.Transactions, error) {
	return b.huc.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return b.huc.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.huc.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.huc.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.huc.txPool.Content()
}

func (b *LesApiBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.huc.txPool.SubscribeNewTxsEvent(ch)
}

func (b *LesApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.huc.blockchain.SubscribeChainEvent(ch)
}

func (b *LesApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.huc.blockchain.SubscribeChainHeadEvent(ch)
}

func (b *LesApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.huc.blockchain.SubscribeChainSideEvent(ch)
}

func (b *LesApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.huc.blockchain.SubscribeLogsEvent(ch)
}

func (b *LesApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.huc.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return b.huc.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.huc.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LesApiBackend) ChainDb() hucdb.Database {
	return b.huc.chainDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return b.huc.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.huc.accountManager
}

func (b *LesApiBackend) BloomStatus() (uint64, uint64) {
	if b.huc.bloomIndexer == nil {
		return 0, 0
	}
	sections, _, _ := b.huc.bloomIndexer.Sections()
	return light.BloomTrieFrequency, sections
}

func (b *LesApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.huc.bloomRequests)
	}
}
