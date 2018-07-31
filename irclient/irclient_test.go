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

package irclient

import "github.com/happyuc-project/happyuc-go"

// Verify that Client implements the happyuc interfaces.
var (
	_ = happyuc.ChainReader(&Client{})
	_ = happyuc.TransactionReader(&Client{})
	_ = happyuc.ChainStateReader(&Client{})
	_ = happyuc.ChainSyncReader(&Client{})
	_ = happyuc.ContractCaller(&Client{})
	_ = happyuc.GasEstimator(&Client{})
	_ = happyuc.GasPricer(&Client{})
	_ = happyuc.LogFilterer(&Client{})
	_ = happyuc.PendingStateReader(&Client{})
	// _ = happyuc.PendingStateEventer(&Client{})
	_ = happyuc.PendingContractCaller(&Client{})
)
