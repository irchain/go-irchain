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

// package webuext contains ghuc specific webu.js extensions.
package webuext

var Modules = map[string]string{
	"admin":      Admin_JS,
	"chequebook": Chequebook_JS,
	"clique":     Clique_JS,
	"debug":      Debug_JS,
	"huc":        Huc_JS,
	"miner":      Miner_JS,
	"net":        Net_JS,
	"personal":   Personal_JS,
	"rpc":        RPC_JS,
	"shh":        Shh_JS,
	"swarmfs":    SWARMFS_JS,
	"txpool":     TxPool_JS,
}

const Chequebook_JS = `
webu._extend({
	property: 'chequebook',
	methods: [
		new webu._extend.Method({
			name: 'deposit',
			call: 'chequebook_deposit',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Property({
			name: 'balance',
			getter: 'chequebook_balance',
			outputFormatter: webu._extend.utils.toDecimal
		}),
		new webu._extend.Method({
			name: 'cash',
			call: 'chequebook_cash',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Method({
			name: 'issue',
			call: 'chequebook_issue',
			params: 2,
			inputFormatter: [null, null]
		}),
	]
});
`

const Clique_JS = `
webu._extend({
	property: 'clique',
	methods: [
		new webu._extend.Method({
			name: 'getSnapshot',
			call: 'clique_getSnapshot',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Method({
			name: 'getSnapshotAtHash',
			call: 'clique_getSnapshotAtHash',
			params: 1
		}),
		new webu._extend.Method({
			name: 'getSigners',
			call: 'clique_getSigners',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Method({
			name: 'getSignersAtHash',
			call: 'clique_getSignersAtHash',
			params: 1
		}),
		new webu._extend.Method({
			name: 'propose',
			call: 'clique_propose',
			params: 2
		}),
		new webu._extend.Method({
			name: 'discard',
			call: 'clique_discard',
			params: 1
		}),
	],
	properties: [
		new webu._extend.Property({
			name: 'proposals',
			getter: 'clique_proposals'
		}),
	]
});
`

const Admin_JS = `
webu._extend({
	property: 'admin',
	methods: [
		new webu._extend.Method({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new webu._extend.Method({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new webu._extend.Method({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Method({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new webu._extend.Method({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new webu._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new webu._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new webu._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new webu._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
	],
	properties: [
		new webu._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new webu._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new webu._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
	]
});
`

const Debug_JS = `
webu._extend({
	property: 'debug',
	methods: [
		new webu._extend.Method({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1
		}),
		new webu._extend.Method({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new webu._extend.Method({
			name: 'setHead',
			call: 'debug_setHead',
			params: 1
		}),
		new webu._extend.Method({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new webu._extend.Method({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1
		}),
		new webu._extend.Method({
			name: 'chaindbProperty',
			call: 'debug_chaindbProperty',
			params: 1,
			outputFormatter: console.log
		}),
		new webu._extend.Method({
			name: 'chaindbCompact',
			call: 'debug_chaindbCompact',
		}),
		new webu._extend.Method({
			name: 'metrics',
			call: 'debug_metrics',
			params: 1
		}),
		new webu._extend.Method({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new webu._extend.Method({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new webu._extend.Method({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new webu._extend.Method({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new webu._extend.Method({
			name: 'freeOSMemory',
			call: 'debug_freeOSMemory',
			params: 0,
		}),
		new webu._extend.Method({
			name: 'setGCPercent',
			call: 'debug_setGCPercent',
			params: 1,
		}),
		new webu._extend.Method({
			name: 'memStats',
			call: 'debug_memStats',
			params: 0,
		}),
		new webu._extend.Method({
			name: 'gcStats',
			call: 'debug_gcStats',
			params: 0,
		}),
		new webu._extend.Method({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new webu._extend.Method({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new webu._extend.Method({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new webu._extend.Method({
			name: 'goTrace',
			call: 'debug_goTrace',
			params: 2
		}),
		new webu._extend.Method({
			name: 'startGoTrace',
			call: 'debug_startGoTrace',
			params: 1
		}),
		new webu._extend.Method({
			name: 'stopGoTrace',
			call: 'debug_stopGoTrace',
			params: 0
		}),
		new webu._extend.Method({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new webu._extend.Method({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new webu._extend.Method({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new webu._extend.Method({
			name: 'mutexProfile',
			call: 'debug_mutexProfile',
			params: 2
		}),
		new webu._extend.Method({
			name: 'setMutexProfileRate',
			call: 'debug_setMutexProfileRate',
			params: 1
		}),
		new webu._extend.Method({
			name: 'writeMutexProfile',
			call: 'debug_writeMutexProfile',
			params: 1
		}),
		new webu._extend.Method({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new webu._extend.Method({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2,
			inputFormatter: [null, null]
		}),
		new webu._extend.Method({
			name: 'traceBlockFromFile',
			call: 'debug_traceBlockFromFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new webu._extend.Method({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2,
			inputFormatter: [null, null]
		}),
		new webu._extend.Method({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2,
			inputFormatter: [null, null]
		}),
		new webu._extend.Method({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2,
			inputFormatter: [null, null]
		}),
		new webu._extend.Method({
			name: 'preimage',
			call: 'debug_preimage',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Method({
			name: 'getBadBlocks',
			call: 'debug_getBadBlocks',
			params: 0,
		}),
		new webu._extend.Method({
			name: 'storageRangeAt',
			call: 'debug_storageRangeAt',
			params: 5,
		}),
		new webu._extend.Method({
			name: 'getModifiedAccountsByNumber',
			call: 'debug_getModifiedAccountsByNumber',
			params: 2,
			inputFormatter: [null, null],
		}),
		new webu._extend.Method({
			name: 'getModifiedAccountsByHash',
			call: 'debug_getModifiedAccountsByHash',
			params: 2,
			inputFormatter:[null, null],
		}),
	],
	properties: []
});
`

const Huc_JS = `
webu._extend({
	property: 'huc',
	methods: [
		new webu._extend.Method({
			name: 'sign',
			call: 'huc_sign',
			params: 2,
			inputFormatter: [webu._extend.formatters.inputAddressFormatter, null]
		}),
		new webu._extend.Method({
			name: 'resend',
			call: 'huc_resend',
			params: 3,
			inputFormatter: [webu._extend.formatters.inputTransactionFormatter, webu._extend.utils.fromDecimal, webu._extend.utils.fromDecimal]
		}),
		new webu._extend.Method({
			name: 'signTransaction',
			call: 'huc_signTransaction',
			params: 1,
			inputFormatter: [webu._extend.formatters.inputTransactionFormatter]
		}),
		new webu._extend.Method({
			name: 'submitTransaction',
			call: 'huc_submitTransaction',
			params: 1,
			inputFormatter: [webu._extend.formatters.inputTransactionFormatter]
		}),
		new webu._extend.Method({
			name: 'getRawTransaction',
			call: 'huc_getRawTransactionByHash',
			params: 1
		}),
		new webu._extend.Method({
			name: 'getRawTransactionFromBlock',
			call: function(args) {
				return (webu._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'huc_getRawTransactionByBlockHashAndIndex' : 'huc_getRawTransactionByBlockNumberAndIndex';
			},
			params: 2,
			inputFormatter: [webu._extend.formatters.inputBlockNumberFormatter, webu._extend.utils.toHex]
		}),
	],
	properties: [
		new webu._extend.Property({
			name: 'pendingTransactions',
			getter: 'huc_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(webu._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
	]
});
`

const Miner_JS = `
webu._extend({
	property: 'miner',
	methods: [
		new webu._extend.Method({
			name: 'start',
			call: 'miner_start',
			params: 1,
			inputFormatter: [null]
		}),
		new webu._extend.Method({
			name: 'stop',
			call: 'miner_stop'
		}),
		new webu._extend.Method({
			name: 'setCoinbase',
			call: 'miner_setCoinbase',
			params: 1,
			inputFormatter: [webu._extend.formatters.inputAddressFormatter]
		}),
		new webu._extend.Method({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new webu._extend.Method({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [webu._extend.utils.fromDecimal]
		}),
		new webu._extend.Method({
			name: 'getHashrate',
			call: 'miner_getHashrate'
		}),
	],
	properties: []
});
`

const Net_JS = `
webu._extend({
	property: 'net',
	methods: [],
	properties: [
		new webu._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
	]
});
`

const Personal_JS = `
webu._extend({
	property: 'personal',
	methods: [
		new webu._extend.Method({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
		new webu._extend.Method({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, webu._extend.formatters.inputAddressFormatter, null]
		}),
		new webu._extend.Method({
			name: 'ecRecover',
			call: 'personal_ecRecover',
			params: 2
		}),
		new webu._extend.Method({
			name: 'openWallet',
			call: 'personal_openWallet',
			params: 2
		}),
		new webu._extend.Method({
			name: 'deriveAccount',
			call: 'personal_deriveAccount',
			params: 3
		}),
		new webu._extend.Method({
			name: 'signTransaction',
			call: 'personal_signTransaction',
			params: 2,
			inputFormatter: [webu._extend.formatters.inputTransactionFormatter, null]
		}),
	],
	properties: [
		new webu._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`

const RPC_JS = `
webu._extend({
	property: 'rpc',
	methods: [],
	properties: [
		new webu._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const Shh_JS = `
webu._extend({
	property: 'shh',
	methods: [
	],
	properties:
	[
		new webu._extend.Property({
			name: 'version',
			getter: 'shh_version',
			outputFormatter: webu._extend.utils.toDecimal
		}),
		new webu._extend.Property({
			name: 'info',
			getter: 'shh_info'
		}),
	]
});
`

const SWARMFS_JS = `
webu._extend({
	property: 'swarmfs',
	methods:
	[
		new webu._extend.Method({
			name: 'mount',
			call: 'swarmfs_mount',
			params: 2
		}),
		new webu._extend.Method({
			name: 'unmount',
			call: 'swarmfs_unmount',
			params: 1
		}),
		new webu._extend.Method({
			name: 'listmounts',
			call: 'swarmfs_listmounts',
			params: 0
		}),
	]
});
`

const TxPool_JS = `
webu._extend({
	property: 'txpool',
	methods: [],
	properties:
	[
		new webu._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new webu._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new webu._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = webu._extend.utils.toDecimal(status.pending);
				status.queued = webu._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
	]
});
`
