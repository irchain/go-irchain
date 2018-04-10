// Copyright 2014 The happyuc-go Authors
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

// Package huc implements the HappyUC protocol.
package huc

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/happyuc-project/happyuc-go/accounts"
	"github.com/happyuc-project/happyuc-go/common"
	"github.com/happyuc-project/happyuc-go/common/hexutil"
	"github.com/happyuc-project/happyuc-go/consensus"
	"github.com/happyuc-project/happyuc-go/consensus/clique"
	"github.com/happyuc-project/happyuc-go/consensus/huchash"
	"github.com/happyuc-project/happyuc-go/core"
	"github.com/happyuc-project/happyuc-go/core/bloombits"
	"github.com/happyuc-project/happyuc-go/core/types"
	"github.com/happyuc-project/happyuc-go/core/vm"
	"github.com/happyuc-project/happyuc-go/huc/downloader"
	"github.com/happyuc-project/happyuc-go/huc/filters"
	"github.com/happyuc-project/happyuc-go/huc/gasprice"
	"github.com/happyuc-project/happyuc-go/hucdb"
	"github.com/happyuc-project/happyuc-go/event"
	"github.com/happyuc-project/happyuc-go/internal/hucapi"
	"github.com/happyuc-project/happyuc-go/log"
	"github.com/happyuc-project/happyuc-go/miner"
	"github.com/happyuc-project/happyuc-go/node"
	"github.com/happyuc-project/happyuc-go/p2p"
	"github.com/happyuc-project/happyuc-go/params"
	"github.com/happyuc-project/happyuc-go/rlp"
	"github.com/happyuc-project/happyuc-go/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// HappyUC implements the HappyUC full node service.
type HappyUC struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the happyuc
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb hucdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *EthApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	coinbase common.Address

	networkId     uint64
	netRPCService *hucapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and coinbase)
}

func (huc *HappyUC) AddLesServer(ls LesServer) {
	huc.lesServer = ls
	ls.SetBloomBitsIndexer(huc.bloomIndexer)
}

// New creates a new HappyUC object (including the
// initialisation of the common HappyUC object)
func New(ctx *node.ServiceContext, config *Config) (*HappyUC, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run huc.HappyUC in light sync mode, use les.LightHappyUC")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	huc := &HappyUC{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Huchash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		coinbase:       config.Coinbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising HappyUC protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run ghuc upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	huc.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, huc.chainConfig, huc.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		huc.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	huc.bloomIndexer.Start(huc.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	huc.txPool = core.NewTxPool(config.TxPool, huc.chainConfig, huc.blockchain)

	if huc.protocolManager, err = NewProtocolManager(huc.chainConfig, config.SyncMode, config.NetworkId, huc.eventMux, huc.txPool, huc.engine, huc.blockchain, chainDb); err != nil {
		return nil, err
	}
	huc.miner = miner.New(huc, huc.chainConfig, huc.EventMux(), huc.engine)
	huc.miner.SetExtra(makeExtraData(config.ExtraData))

	huc.ApiBackend = &EthApiBackend{huc, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	huc.ApiBackend.gpo = gasprice.NewOracle(huc.ApiBackend, gpoParams)

	return huc, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"ghuc",
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
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (hucdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*hucdb.LDBDatabase); ok {
		db.Meter("huc/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an HappyUC service
func CreateConsensusEngine(ctx *node.ServiceContext, config *huchash.Config, chainConfig *params.ChainConfig, db hucdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == huchash.ModeFake:
		log.Warn("Huchash used in fake mode")
		return huchash.NewFaker()
	case config.PowMode == huchash.ModeTest:
		log.Warn("Huchash used in test mode")
		return huchash.NewTester()
	case config.PowMode == huchash.ModeShared:
		log.Warn("Huchash used in shared mode")
		return huchash.NewShared()
	default:
		engine := huchash.New(huchash.Config{
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

// APIs returns the collection of RPC services the happyuc package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (huc *HappyUC) APIs() []rpc.API {
	apis := hucapi.GetAPIs(huc.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, huc.engine.APIs(huc.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "huc",
			Version:   "1.0",
			Service:   NewPublicHappyUCAPI(huc),
			Public:    true,
		}, {
			Namespace: "huc",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(huc),
			Public:    true,
		}, {
			Namespace: "huc",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(huc.protocolManager.downloader, huc.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(huc),
			Public:    false,
		}, {
			Namespace: "huc",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(huc.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(huc),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(huc),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(huc.chainConfig, huc),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   huc.netRPCService,
			Public:    true,
		},
	}...)
}

func (huc *HappyUC) ResetWithGenesisBlock(gb *types.Block) {
	huc.blockchain.ResetWithGenesisBlock(gb)
}

func (huc *HappyUC) Coinbase() (eb common.Address, err error) {
	huc.lock.RLock()
	coinbase := huc.coinbase
	huc.lock.RUnlock()

	if coinbase != (common.Address{}) {
		return coinbase, nil
	}
	if wallets := huc.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			coinbase := accounts[0].Address

			huc.lock.Lock()
			huc.coinbase = coinbase
			huc.lock.Unlock()

			log.Info("Coinbase automatically configured", "address", coinbase)
			return coinbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("coinbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (huc *HappyUC) SetCoinbase(coinbase common.Address) {
	huc.lock.Lock()
	huc.coinbase = coinbase
	huc.lock.Unlock()

	huc.miner.SetCoinbase(coinbase)
}

func (huc *HappyUC) StartMining(local bool) error {
	eb, err := huc.Coinbase()
	if err != nil {
		log.Error("Cannot start mining without coinbase", "err", err)
		return fmt.Errorf("coinbase missing: %v", err)
	}
	if clique, ok := huc.engine.(*clique.Clique); ok {
		wallet, err := huc.accountManager.Find(accounts.Account{Address: eb})
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
		atomic.StoreUint32(&huc.protocolManager.acceptTxs, 1)
	}
	go huc.miner.Start(eb)
	return nil
}

func (huc *HappyUC) StopMining()         { huc.miner.Stop() }
func (huc *HappyUC) IsMining() bool      { return huc.miner.Mining() }
func (huc *HappyUC) Miner() *miner.Miner { return huc.miner }

func (huc *HappyUC) AccountManager() *accounts.Manager  { return huc.accountManager }
func (huc *HappyUC) BlockChain() *core.BlockChain       { return huc.blockchain }
func (huc *HappyUC) TxPool() *core.TxPool               { return huc.txPool }
func (huc *HappyUC) EventMux() *event.TypeMux           { return huc.eventMux }
func (huc *HappyUC) Engine() consensus.Engine           { return huc.engine }
func (huc *HappyUC) ChainDb() hucdb.Database            { return huc.chainDb }
func (huc *HappyUC) IsListening() bool                  { return true } // Always listening
func (huc *HappyUC) EthVersion() int                    { return int(huc.protocolManager.SubProtocols[0].Version) }
func (huc *HappyUC) NetVersion() uint64                 { return huc.networkId }
func (huc *HappyUC) Downloader() *downloader.Downloader { return huc.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (huc *HappyUC) Protocols() []p2p.Protocol {
	if huc.lesServer == nil {
		return huc.protocolManager.SubProtocols
	}
	return append(huc.protocolManager.SubProtocols, huc.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// HappyUC protocol implementation.
func (huc *HappyUC) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	huc.startBloomHandlers()

	// Start the RPC service
	huc.netRPCService = hucapi.NewPublicNetAPI(srvr, huc.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if huc.config.LightServ > 0 {
		if huc.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", huc.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= huc.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	huc.protocolManager.Start(maxPeers)
	if huc.lesServer != nil {
		huc.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// HappyUC protocol.
func (huc *HappyUC) Stop() error {
	if huc.stopDbUpgrade != nil {
		huc.stopDbUpgrade()
	}
	huc.bloomIndexer.Close()
	huc.blockchain.Stop()
	huc.protocolManager.Stop()
	if huc.lesServer != nil {
		huc.lesServer.Stop()
	}
	huc.txPool.Stop()
	huc.miner.Stop()
	huc.eventMux.Stop()

	huc.chainDb.Close()
	close(huc.shutdownChan)

	return nil
}
