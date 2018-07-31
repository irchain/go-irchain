// Copyright 2014 The go-irchain Authors
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

// Package irc implements the IrChain protocol.
package irc

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/irchain/go-irchain/accounts"
	"github.com/irchain/go-irchain/common"
	"github.com/irchain/go-irchain/common/hexutil"
	"github.com/irchain/go-irchain/consensus"
	"github.com/irchain/go-irchain/consensus/clique"
	"github.com/irchain/go-irchain/consensus/irchash"
	"github.com/irchain/go-irchain/core"
	"github.com/irchain/go-irchain/core/bloombits"
	"github.com/irchain/go-irchain/core/rawdb"
	"github.com/irchain/go-irchain/core/types"
	"github.com/irchain/go-irchain/core/vm"
	"github.com/irchain/go-irchain/event"
	"github.com/irchain/go-irchain/irc/downloader"
	"github.com/irchain/go-irchain/irc/filters"
	"github.com/irchain/go-irchain/irc/gasprice"
	"github.com/irchain/go-irchain/ircdb"
	"github.com/irchain/go-irchain/internal/ircapi"
	"github.com/irchain/go-irchain/log"
	"github.com/irchain/go-irchain/miner"
	"github.com/irchain/go-irchain/node"
	"github.com/irchain/go-irchain/p2p"
	"github.com/irchain/go-irchain/params"
	"github.com/irchain/go-irchain/rlp"
	"github.com/irchain/go-irchain/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// IrChain implements the IrChain full node service.
type IrChain struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the irchain

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb ircdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *IrcApiBackend

	miner    *miner.Miner
	gasPrice *big.Int
	coinbase common.Address

	networkId     uint64
	netRPCService *ircapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and coinbase)
}

func (irc *IrChain) AddLesServer(ls LesServer) {
	irc.lesServer = ls
	ls.SetBloomBitsIndexer(irc.bloomIndexer)
}

// New creates a new IrChain object (including the
// initialisation of the common IrChain object)
func New(ctx *node.ServiceContext, config *Config) (*IrChain, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run irc.IrChain in light sync mode, use les.LightIrChain")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	irc := &IrChain{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Irchash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		coinbase:       config.Coinbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising IrChain protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run geth upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	irc.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, irc.chainConfig, irc.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		irc.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	irc.bloomIndexer.Start(irc.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	irc.txPool = core.NewTxPool(config.TxPool, irc.chainConfig, irc.blockchain)

	if irc.protocolManager, err = NewProtocolManager(irc.chainConfig, config.SyncMode, config.NetworkId, irc.eventMux, irc.txPool, irc.engine, irc.blockchain, chainDb); err != nil {
		return nil, err
	}
	irc.miner = miner.New(irc, irc.chainConfig, irc.EventMux(), irc.engine)
	irc.miner.SetExtra(makeExtraData(config.ExtraData))

	irc.ApiBackend = &IrcApiBackend{irc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	irc.ApiBackend.gpo = gasprice.NewOracle(irc.ApiBackend, gpoParams)

	return irc, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"girc",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (ircdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*ircdb.LDBDatabase); ok {
		db.Meter("irc/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an IrChain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *irchash.Config, chainConfig *params.ChainConfig, db ircdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == irchash.ModeFake:
		log.Warn("Irchash used in fake mode")
		return irchash.NewFaker()
	case config.PowMode == irchash.ModeTest:
		log.Warn("Irchash used in test mode")
		return irchash.NewTester()
	case config.PowMode == irchash.ModeShared:
		log.Warn("Irchash used in shared mode")
		return irchash.NewShared()
	default:
		engine := irchash.New(irchash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the irchain package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (irc *IrChain) APIs() []rpc.API {
	apis := ircapi.GetAPIs(irc.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, irc.engine.APIs(irc.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "irc",
			Version:   "1.0",
			Service:   NewPublicIrChainAPI(irc),
			Public:    true,
		}, {
			Namespace: "irc",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(irc),
			Public:    true,
		}, {
			Namespace: "irc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(irc.protocolManager.downloader, irc.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(irc),
			Public:    false,
		}, {
			Namespace: "irc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(irc.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(irc),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(irc),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(irc.chainConfig, irc),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   irc.netRPCService,
			Public:    true,
		},
	}...)
}

func (irc *IrChain) ResetWithGenesisBlock(gb *types.Block) {
	irc.blockchain.ResetWithGenesisBlock(gb)
}

func (irc *IrChain) Coinbase() (eb common.Address, err error) {
	irc.lock.RLock()
	coinbase := irc.coinbase
	irc.lock.RUnlock()

	if coinbase != (common.Address{}) {
		return coinbase, nil
	}
	if wallets := irc.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			coinbase := accounts[0].Address

			irc.lock.Lock()
			irc.coinbase = coinbase
			irc.lock.Unlock()

			log.Info("Coinbase automatically configured", "address", coinbase)
			return coinbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("coinbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (irc *IrChain) SetCoinbase(coinbase common.Address) {
	irc.lock.Lock()
	irc.coinbase = coinbase
	irc.lock.Unlock()

	irc.miner.SetCoinbase(coinbase)
}

func (irc *IrChain) StartMining(local bool) error {
	eb, err := irc.Coinbase()
	if err != nil {
		log.Error("Cannot start mining without coinbase", "err", err)
		return fmt.Errorf("coinbase missing: %v", err)
	}
	if clique, ok := irc.engine.(*clique.Clique); ok {
		wallet, err := irc.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Coinbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&irc.protocolManager.acceptTxs, 1)
	}
	go irc.miner.Start(eb)
	return nil
}

func (irc *IrChain) StopMining()         { irc.miner.Stop() }
func (irc *IrChain) IsMining() bool      { return irc.miner.Mining() }
func (irc *IrChain) Miner() *miner.Miner { return irc.miner }

func (irc *IrChain) AccountManager() *accounts.Manager  { return irc.accountManager }
func (irc *IrChain) BlockChain() *core.BlockChain       { return irc.blockchain }
func (irc *IrChain) TxPool() *core.TxPool               { return irc.txPool }
func (irc *IrChain) EventMux() *event.TypeMux           { return irc.eventMux }
func (irc *IrChain) Engine() consensus.Engine           { return irc.engine }
func (irc *IrChain) ChainDb() ircdb.Database            { return irc.chainDb }
func (irc *IrChain) IsListening() bool                  { return true } // Always listening
func (irc *IrChain) IrcVersion() int                    { return int(irc.protocolManager.SubProtocols[0].Version) }
func (irc *IrChain) NetVersion() uint64                 { return irc.networkId }
func (irc *IrChain) Downloader() *downloader.Downloader { return irc.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (irc *IrChain) Protocols() []p2p.Protocol {
	if irc.lesServer == nil {
		return irc.protocolManager.SubProtocols
	}
	return append(irc.protocolManager.SubProtocols, irc.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// IrChain protocol implementation.
func (irc *IrChain) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	irc.startBloomHandlers()

	// Start the RPC service
	irc.netRPCService = ircapi.NewPublicNetAPI(srvr, irc.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if irc.config.LightServ > 0 {
		if irc.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", irc.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= irc.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	irc.protocolManager.Start(maxPeers)
	if irc.lesServer != nil {
		irc.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// IrChain protocol.
func (irc *IrChain) Stop() error {
	irc.bloomIndexer.Close()
	irc.blockchain.Stop()
	irc.protocolManager.Stop()
	if irc.lesServer != nil {
		irc.lesServer.Stop()
	}
	irc.txPool.Stop()
	irc.miner.Stop()
	irc.eventMux.Stop()

	irc.chainDb.Close()
	close(irc.shutdownChan)

	return nil
}
