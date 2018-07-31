// Copyright 2016 The go-irchain Authors
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

// Package les implements the Light IrChain Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/irchain/go-irchain/accounts"
	"github.com/irchain/go-irchain/common"
	"github.com/irchain/go-irchain/common/hexutil"
	"github.com/irchain/go-irchain/consensus"
	"github.com/irchain/go-irchain/core"
	"github.com/irchain/go-irchain/core/bloombits"
	"github.com/irchain/go-irchain/core/rawdb"
	"github.com/irchain/go-irchain/core/types"
	"github.com/irchain/go-irchain/event"
	"github.com/irchain/go-irchain/irc"
	"github.com/irchain/go-irchain/irc/downloader"
	"github.com/irchain/go-irchain/irc/filters"
	"github.com/irchain/go-irchain/irc/gasprice"
	"github.com/irchain/go-irchain/ircdb"
	"github.com/irchain/go-irchain/internal/ircapi"
	"github.com/irchain/go-irchain/light"
	"github.com/irchain/go-irchain/log"
	"github.com/irchain/go-irchain/node"
	"github.com/irchain/go-irchain/p2p"
	"github.com/irchain/go-irchain/p2p/discv5"
	"github.com/irchain/go-irchain/params"
	"github.com/irchain/go-irchain/rpc"
	_ "github.com/irchain/go-irchain/swarm/api"
)

type LightIrChain struct {
	config *irc.Config

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
	chainDb ircdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *ircapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *irc.Config) (*LightIrChain, error) {
	chainDb, err := irc.CreateDB(ctx, config, "lightchaindata")
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

	lirc := &LightIrChain{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           irc.CreateConsensusEngine(ctx, &config.Irchash, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     irc.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	lirc.relay = NewLesTxRelay(peers, lirc.reqDist)
	lirc.serverPool = newServerPool(chainDb, quitSync, &lirc.wg)
	lirc.retriever = newRetrieveManager(peers, lirc.reqDist, lirc.serverPool)
	lirc.odr = NewLesOdr(chainDb, lirc.chtIndexer, lirc.bloomTrieIndexer, lirc.bloomIndexer, lirc.retriever)
	if lirc.blockchain, err = light.NewLightChain(lirc.odr, lirc.chainConfig, lirc.engine); err != nil {
		return nil, err
	}
	lirc.bloomIndexer.Start(lirc.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lirc.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lirc.txPool = light.NewTxPool(lirc.chainConfig, lirc.blockchain, lirc.relay)
	if lirc.protocolManager, err = NewProtocolManager(lirc.chainConfig, true, ClientProtocolVersions, config.NetworkId, lirc.eventMux, lirc.engine, lirc.peers, lirc.blockchain, nil, chainDb, lirc.odr, lirc.relay, quitSync, &lirc.wg); err != nil {
		return nil, err
	}
	lirc.ApiBackend = &LesApiBackend{lirc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lirc.ApiBackend.gpo = gasprice.NewOracle(lirc.ApiBackend, gpoParams)
	return lirc, nil
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

// APIs returns the collection of RPC services the irchain package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (irc *LightIrChain) APIs() []rpc.API {
	return append(ircapi.GetAPIs(irc.ApiBackend), []rpc.API{
		{
			Namespace: "irc",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "irc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(irc.protocolManager.downloader, irc.eventMux),
			Public:    true,
		}, {
			Namespace: "irc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(irc.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   irc.netRPCService,
			Public:    true,
		},
	}...)
}

func (irc *LightIrChain) ResetWithGenesisBlock(gb *types.Block) {
	irc.blockchain.ResetWithGenesisBlock(gb)
}

func (irc *LightIrChain) BlockChain() *light.LightChain      { return irc.blockchain }
func (irc *LightIrChain) TxPool() *light.TxPool              { return irc.txPool }
func (irc *LightIrChain) Engine() consensus.Engine           { return irc.engine }
func (irc *LightIrChain) LesVersion() int                    { return int(irc.protocolManager.SubProtocols[0].Version) }
func (irc *LightIrChain) Downloader() *downloader.Downloader { return irc.protocolManager.downloader }
func (irc *LightIrChain) EventMux() *event.TypeMux           { return irc.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (irc *LightIrChain) Protocols() []p2p.Protocol {
	return irc.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// IrChain protocol implementation.
func (irc *LightIrChain) Start(srvr *p2p.Server) error {
	irc.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	irc.netRPCService = ircapi.NewPublicNetAPI(srvr, irc.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	irc.serverPool.start(srvr, lesTopic(irc.blockchain.Genesis().Hash(), protocolVersion))
	irc.protocolManager.Start(irc.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// IrChain protocol.
func (irc *LightIrChain) Stop() error {
	irc.odr.Stop()
	if irc.bloomIndexer != nil {
		irc.bloomIndexer.Close()
	}
	if irc.chtIndexer != nil {
		irc.chtIndexer.Close()
	}
	if irc.bloomTrieIndexer != nil {
		irc.bloomTrieIndexer.Close()
	}
	irc.blockchain.Stop()
	irc.protocolManager.Stop()
	irc.txPool.Stop()

	irc.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	irc.chainDb.Close()
	close(irc.shutdownChan)

	return nil
}
