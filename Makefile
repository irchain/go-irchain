# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: girc android ios girc-cross swarm evm all test clean
.PHONY: girc-linux girc-linux-386 girc-linux-amd64 girc-linux-mips64 girc-linux-mips64le
.PHONY: girc-linux-arm girc-linux-arm-5 girc-linux-arm-6 girc-linux-arm-7 girc-linux-arm64
.PHONY: girc-darwin girc-darwin-386 girc-darwin-amd64
.PHONY: girc-windows girc-windows-386 girc-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

girc:
	build/env.sh go run build/ci.go install ./cmd/girc
	@echo "Done building."
	@echo "Run \"$(GOBIN)/girc\" to launch girc."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/girc.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Girc.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

girc-cross: girc-linux girc-darwin girc-windows girc-android girc-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/girc-*

girc-linux: girc-linux-386 girc-linux-amd64 girc-linux-arm girc-linux-mips64 girc-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-*

girc-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/girc
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep 386

girc-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/girc
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep amd64

girc-linux-arm: girc-linux-arm-5 girc-linux-arm-6 girc-linux-arm-7 girc-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep arm

girc-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/girc
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep arm-5

girc-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/girc
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep arm-6

girc-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/girc
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep arm-7

girc-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/girc
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep arm64

girc-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/girc
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep mips

girc-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/girc
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep mipsle

girc-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/girc
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep mips64

girc-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/girc
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/girc-linux-* | grep mips64le

girc-darwin: girc-darwin-386 girc-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/girc-darwin-*

girc-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/girc
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/girc-darwin-* | grep 386

girc-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/girc
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/girc-darwin-* | grep amd64

girc-windows: girc-windows-386 girc-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/girc-windows-*

girc-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/girc
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/girc-windows-* | grep 386

girc-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/girc
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/girc-windows-* | grep amd64
