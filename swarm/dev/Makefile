.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/happyuc-project/happyuc-go/cmd/swarm
	go build -o bin/geth     github.com/happyuc-project/happyuc-go/cmd/geth
	go build -o bin/bootnode github.com/happyuc-project/happyuc-go/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/happyuc-project/happyuc-go/swarm/...
