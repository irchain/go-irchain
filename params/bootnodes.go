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

// MainnetBootnodes are the hnode URLs of the P2P bootstrap nodes running on
// the main HappyUC network.
var MainnetBootnodes = []string{
	"hnode://07fcc60a717613f1502e1925d0e1c74d8532917b1b26159e2aa294dd1f12b9863bcbbf6e6ae64681f8f6d228223b87a14ed1032411b6fa6214828a332aaadf31@72.11.140.162:50505",
	"hnode://149e66442d2d9bb1f07479513778dd6e9a9cd90f89e0297863aff3ad1c9a8c91ce4f45476b38c428712bdd7c891c7515650073baf47d99041676f655af0293d1@47.94.56.101:50505",
	"hnode://47f572691118b98904959cb20321c9df26c1ab70d5507aad40edc58b94095c15c3d2057519750f4b32e3963ebff2c82e0cbe9e1383ebb157b01915722cd084bc@193.112.32.158:50505",
	"hnode://890be21099778151f3130c69dc676b70125fc8ec109a852177fd85ef68aa701dc395baa3d07caa9b370e4bc950228721e140d0b8878c7ef10be3fd427562299c@218.253.193.226:50505",
	"hnode://bbcd7ae3055e43ebe3848a011bf59b5911391e96efdd8bebc431c267ff381b131d686a06b0d05b44cd2ef2dd00e405850438afe9c4586d8e576fb93cfc3b495a@59.111.94.209:50505",
	"hnode://dd50fbb46efab23487d4eb7fea4e7d6e14ac8a6daabbae0a82446f88a5d1d9aade7f0a8b90eaccc343cb8f9386bf90020c724d854a6d61da4fa0fbcb7167a157@111.230.27.226:50505",
	"hnode://df6f3688406de2e6b43106ec011a01d8e85f7d25afdb8fd3d213bee9ff99ed8790fc410b72b03100772d98f11bc126f29e39e05b6d08626817be48cce3b517b9@112.74.96.198:50505",
}

// TestnetBootnodes are the hnode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"hnode://0aaab2e176b97fdd7900c963a0cbe7a6497bcefca8ce82ff4535132be6114b868577093db600957fbccacdf05267f227017a2abeddefc825d5282cb0350d3ab8@112.74.96.198:40404",
}

// RinkebyBootnodes are the hnode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes []string

// DiscoveryV5Bootnodes are the hnode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes []string
