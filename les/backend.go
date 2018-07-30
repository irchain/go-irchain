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

// Package les implements the Light HappyUC Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/happyuc-project/happyuc-go/accounts"
	"github.com/happyuc-project/happyuc-go/common"
	"github.com/happyuc-project/happyuc-go/common/hexutil"
	"github.com/happyuc-project/happyuc-go/consensus"
	"github.com/happyuc-project/happyuc-go/core"
	"github.com/happyuc-project/happyuc-go/core/bloombits"
	"github.com/happyuc-project/happyuc-go/core/rawdb"
	"github.com/happyuc-project/happyuc-go/core/types"
	"github.com/happyuc-project/happyuc-go/event"
	"github.com/happyuc-project/happyuc-go/huc"
	"github.com/happyuc-project/happyuc-go/huc/downloader"
	"github.com/happyuc-project/happyuc-go/huc/filters"
	"github.com/happyuc-project/happyuc-go/huc/gasprice"
	"github.com/happyuc-project/happyuc-go/hucdb"
	"github.com/happyuc-project/happyuc-go/internal/hucapi"
	"github.com/happyuc-project/happyuc-go/light"
	"github.com/happyuc-project/happyuc-go/log"
	"github.com/happyuc-project/happyuc-go/node"
	"github.com/happyuc-project/happyuc-go/p2p"
	"github.com/happyuc-project/happyuc-go/p2p/discv5"
	"github.com/happyuc-project/happyuc-go/params"
	"github.com/happyuc-project/happyuc-go/rpc"
	_ "github.com/happyuc-project/happyuc-go/swarm/api"
)

type LightHappyUC struct {
	config *huc.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb hucdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *hucapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *huc.Config) (*LightHappyUC, error) {
	chainDb, err := huc.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lhuc := &LightHappyUC{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           huc.CreateConsensusEngine(ctx, &config.Huchash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     huc.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	lhuc.relay = NewLesTxRelay(peers, lhuc.reqDist)
	lhuc.serverPool = newServerPool(chainDb, quitSync, &lhuc.wg)
	lhuc.retriever = newRetrieveManager(peers, lhuc.reqDist, lhuc.serverPool)
	lhuc.odr = NewLesOdr(chainDb, lhuc.chtIndexer, lhuc.bloomTrieIndexer, lhuc.bloomIndexer, lhuc.retriever)
	if lhuc.blockchain, err = light.NewLightChain(lhuc.odr, lhuc.chainConfig, lhuc.engine); err != nil {
		return nil, err
	}
	lhuc.bloomIndexer.Start(lhuc.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lhuc.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lhuc.txPool = light.NewTxPool(lhuc.chainConfig, lhuc.blockchain, lhuc.relay)
	if lhuc.protocolManager, err = NewProtocolManager(lhuc.chainConfig, true, ClientProtocolVersions, config.NetworkId, lhuc.eventMux, lhuc.engine, lhuc.peers, lhuc.blockchain, nil, chainDb, lhuc.odr, lhuc.relay, quitSync, &lhuc.wg); err != nil {
		return nil, err
	}
	lhuc.ApiBackend = &LesApiBackend{lhuc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lhuc.ApiBackend.gpo = gasprice.NewOracle(lhuc.ApiBackend, gpoParams)
	return lhuc, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Coinbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the happyuc package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (huc *LightHappyUC) APIs() []rpc.API {
	return append(hucapi.GetAPIs(huc.ApiBackend), []rpc.API{
		{
			Namespace: "huc",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "huc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(huc.protocolManager.downloader, huc.eventMux),
			Public:    true,
		}, {
			Namespace: "huc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(huc.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   huc.netRPCService,
			Public:    true,
		},
	}...)
}

func (huc *LightHappyUC) ResetWithGenesisBlock(gb *types.Block) {
	huc.blockchain.ResetWithGenesisBlock(gb)
}

func (huc *LightHappyUC) BlockChain() *light.LightChain      { return huc.blockchain }
func (huc *LightHappyUC) TxPool() *light.TxPool              { return huc.txPool }
func (huc *LightHappyUC) Engine() consensus.Engine           { return huc.engine }
func (huc *LightHappyUC) LesVersion() int                    { return int(huc.protocolManager.SubProtocols[0].Version) }
func (huc *LightHappyUC) Downloader() *downloader.Downloader { return huc.protocolManager.downloader }
func (huc *LightHappyUC) EventMux() *event.TypeMux           { return huc.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (huc *LightHappyUC) Protocols() []p2p.Protocol {
	return huc.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// HappyUC protocol implementation.
func (huc *LightHappyUC) Start(srvr *p2p.Server) error {
	huc.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	huc.netRPCService = hucapi.NewPublicNetAPI(srvr, huc.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	huc.serverPool.start(srvr, lesTopic(huc.blockchain.Genesis().Hash(), protocolVersion))
	huc.protocolManager.Start(huc.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// HappyUC protocol.
func (huc *LightHappyUC) Stop() error {
	huc.odr.Stop()
	if huc.bloomIndexer != nil {
		huc.bloomIndexer.Close()
	}
	if huc.chtIndexer != nil {
		huc.chtIndexer.Close()
	}
	if huc.bloomTrieIndexer != nil {
		huc.bloomTrieIndexer.Close()
	}
	huc.blockchain.Stop()
	huc.protocolManager.Stop()
	huc.txPool.Stop()

	huc.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	huc.chainDb.Close()
	close(huc.shutdownChan)

	return nil
}
