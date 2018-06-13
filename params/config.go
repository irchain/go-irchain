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

package params

import (
	"fmt"
	"math/big"

	"github.com/happyuc-project/happyuc-go/common"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0xf29c3da3e1710517cbb3a555ab20981ec2c9abacbbcb914ab91e8c23edfbf4d0")
	TestnetGenesisHash = common.HexToHash("0x389d168191585e7a14b01a654c02058053abf3ca3d167efb69a51dec86d9cfbc")
)

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Huchash:             new(HuchashConfig),
	}

	// TestnetChainConfig contains the chain parameters to run a node on the Ropsten test network.
	TestnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(3),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Huchash:             new(HuchashConfig),
	}

	// RinkebyChainConfig contains the chain parameters to run a node on the Rinkeby test network.
	RinkebyChainConfig = &ChainConfig{
		ChainID:             big.NewInt(4),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}

	// AllHuchashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the HappyUC core developers into the Huchash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllHuchashProtocolChanges = &ChainConfig{
		ChainID:             big.NewInt(1337),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Huchash:             new(HuchashConfig),
		Clique:              nil,
	}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the HappyUC core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{
		ChainID:             big.NewInt(1337),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Huchash:             nil,
		Clique: &CliqueConfig{
			Period: 0,
			Epoch:  30000,
		},
	}

	TestChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Huchash:             new(HuchashConfig),
		Clique:              nil,
	}
	// TestRules       = TestChainConfig.Rules(new(big.Int))
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID             *big.Int `json:"chainId"`                       // chainId identifies the current chain and is used for replay protection
	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated )

	// Various consensus engines
	Huchash *HuchashConfig `json:"huchash,omitempty"`
	Clique  *CliqueConfig  `json:"clique,omitempty"`
}

// HuchashConfig is the consensus engine configs for proof-of-work based sealing.
type HuchashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *HuchashConfig) String() string {
	return "huchash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Huchash != nil:
		engine = c.Huchash
	case c.Clique != nil:
		engine = c.Clique
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v Homestead: %v DAO: %v DAOSupport: %v EIP150: %v EIP155: %v EIP158: %v Byzantium: %v Constantinople: %v Engine: %v}",
		c.ChainID,
		c.ByzantiumBlock,
		c.ConstantinopleBlock,
		engine,
	)
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
// func (c *ChainConfig) IsHomestead(num *big.Int) bool {
// 	return isForked(c.HomesteadBlock, num)
// }

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
// func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
// 	return isForked(c.DAOForkBlock, num)
// }

// func (c *ChainConfig) IsEIP150(num *big.Int) bool {
// 	return isForked(c.EIP150Block, num)
// }

// func (c *ChainConfig) IsEIP155(num *big.Int) bool {
// 	return isForked(c.EIP155Block, num)
// }

// func (c *ChainConfig) IsEIP158(num *big.Int) bool {
// 	return isForked(c.EIP158Block, num)
// }

// func (c *ChainConfig) IsByzantium(num *big.Int) bool {
// 	return isForked(c.ByzantiumBlock, num)
// }

// func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
// 	return isForked(c.ConstantinopleBlock, num)
// }

// GasTable returns the gas table corresponding to the current phase (homestead or homestead reprice).
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) GasTable {
	// TODO Return Serenity GasTable
	return GasTableEIP158
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	// TODO Check Compatible
	// if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
	// 	return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	// }
	// if isForkIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
	// 	return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	// }
	// if c.IsDAOFork(head) && c.DAOForkSupport != newcfg.DAOForkSupport {
	// 	return newCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	// }
	// if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
	// 	return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	// }
	// if isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
	// 	return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	// }
	// if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
	// 	return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	// }
	// if c.IsEIP158(head) && !configNumEqual(c.ChainID, newcfg.ChainID) {
	// 	return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	// }
	// if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
	// 	return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	// }
	// if isForkIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
	// 	return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	// }
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntatic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
// type Rules struct {
// 	ChainId     *big.Int
// 	IsHomestead bool
// 	IsEIP150    bool
// 	IsEIP155    bool
// 	IsEIP158    bool
// 	IsByzantium bool
// }

// func (c *ChainConfig) Rules(num *big.Int) Rules {
// 	chainId := c.ChainId
// 	if chainId == nil {
// 		chainId = new(big.Int)
// 	}
// 	return Rules{
// 		ChainId:     new(big.Int).Set(chainId),
// 		IsHomestead: c.IsHomestead(num),
// 		IsEIP150:    c.IsEIP150(num),
// 		IsEIP155:    c.IsEIP155(num),
// 		IsEIP158:    c.IsEIP158(num),
// 		IsByzantium: c.IsByzantium(num),
// 	}
// }
