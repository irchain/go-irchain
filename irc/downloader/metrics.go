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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/irchain/go-irchain/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("irc/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("irc/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("irc/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("irc/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("irc/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("irc/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("irc/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("irc/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("irc/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("irc/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("irc/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("irc/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("irc/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("irc/downloader/states/drop", nil)
)
