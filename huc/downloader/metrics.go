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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/happyuc-project/happyuc-go/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("huc/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("huc/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("huc/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("huc/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("huc/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("huc/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("huc/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("huc/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("huc/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("huc/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("huc/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("huc/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("huc/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("huc/downloader/states/drop", nil)
)
