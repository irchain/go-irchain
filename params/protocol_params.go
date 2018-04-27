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

package params

import "math/big"

var (
	TargetGasLimit uint64 = GenesisGasLimit // The artificial target
)

const (
	GasLimitBoundDivisor  uint64 = 1024    // The bound divisor of the gas limit, used in update calculations.
	QuadCoeffDiv          uint64 = 512     // Divisor for the quadratic particle of the memory cost equation.
	MinGasLimit           uint64 = 0       // Minimum the gas limit may ever be.
	GenesisGasLimit       uint64 = 4712388 // Gas limit of the Genesis block.
	ExpByteGas            uint64 = 0       // Times ceil(log256(exponent)) for the EXP instruction.
	SloadGas              uint64 = 0       // Multiplied by the number of 32-byte words that are copied (round up) for any *COPY operation and added.
	CallValueTransferGas  uint64 = 0       // Paid for CALL when the value transfer is non-zero.
	CallNewAccountGas     uint64 = 0       // Paid for CALL when the destination address didn't exist prior.
	TxGas                 uint64 = 21000   // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxGasContractCreation uint64 = 53000   // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	SstoreSetGas          uint64 = 0       // Once per SLOAD operation.
	LogDataGas            uint64 = 0       // Per byte in a LOG* operation's data.
	CallStipend           uint64 = 2300    // Free gas given at beginning of call.
	Sha3Gas               uint64 = 0       // Once per SHA3 operation.
	Sha3WordGas           uint64 = 0       // Once per word of the SHA3 operation's data.
	SstoreResetGas        uint64 = 0       // Once per SSTORE operation if the zeroness changes from zero.
	SstoreClearGas        uint64 = 0       // Once per SSTORE operation if the zeroness doesn't change.
	SstoreRefundGas       uint64 = 0       // Once per SSTORE operation if the zeroness changes to zero.
	JumpdestGas           uint64 = 0       // Refunded gas, once per SSTORE operation if the zeroness changes to zero.
	EpochDuration         uint64 = 30000   // Duration between proof-of-work epochs.
	CallGas               uint64 = 0       // Once per CALL operation & message call transaction.
	CreateDataGas         uint64 = 0       //
	ExpGas                uint64 = 0       // Once per EXP instruction
	LogGas                uint64 = 0       // Per LOG* operation.
	CopyGas               uint64 = 0       //
	TierStepGas           uint64 = 0       // Once per operation, for a selection of them.
	LogTopicGas           uint64 = 0       // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicGas, LOG4 incurs 4 * c_txLogTopicGas.
	CreateGas             uint64 = 0       // Once per CREATE operation & contract-creation transaction.
	SuicideRefundGas      uint64 = 0       // Refunded following a suicide operation.
	MemoryGas             uint64 = 0       // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.
	TxDataZeroGas         uint64 = 4       // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	TxDataNonZeroGas      uint64 = 68      // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.
	MaximumExtraDataSize  uint64 = 32      // Maximum size extra data may be after Genesis.
	CallCreateDepth       uint64 = 1024    // Maximum depth of call/create stack.
	StackLimit            uint64 = 1024    // Maximum size of VM stack allowed.
	MaxCodeSize                  = 24576   // Maximum bytecode to permit for a contract

	// Precompiled contract gas prices
	EcrecoverGas            uint64 = 0  // Elliptic curve sender recovery gas price
	Sha256BaseGas           uint64 = 0  // Base price for a SHA256 operation
	Sha256PerWordGas        uint64 = 0  // Per-word price for a SHA256 operation
	Ripemd160BaseGas        uint64 = 0  // Base price for a RIPEMD160 operation
	Ripemd160PerWordGas     uint64 = 0  // Per-word price for a RIPEMD160 operation
	IdentityBaseGas         uint64 = 0  // Base price for a data copy operation
	IdentityPerWordGas      uint64 = 0  // Per-work price for a data copy operation
	ModExpQuadCoeffDiv      uint64 = 20 // Divisor for the quadratic particle of the big int modular exponentiation
	Bn256AddGas             uint64 = 0  // Gas needed for an elliptic curve addition
	Bn256ScalarMulGas       uint64 = 0  // Gas needed for an elliptic curve scalar multiplication
	Bn256PairingBaseGas     uint64 = 0  // Base price for an elliptic curve pairing check
	Bn256PairingPerPointGas uint64 = 0  // Per-point price for an elliptic curve pairing check
)

var (
	DifficultyBoundDivisor = big.NewInt(2048)   // The bound divisor of the difficulty, used in the update calculations.
	GenesisDifficulty      = big.NewInt(131072) // Difficulty of the Genesis block.
	MinimumDifficulty      = big.NewInt(131072) // The minimum that the difficulty may ever be.
	DurationLimit          = big.NewInt(13)     // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)
