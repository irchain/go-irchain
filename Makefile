# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: ghuc android ios ghuc-cross swarm evm all test clean
.PHONY: ghuc-linux ghuc-linux-386 ghuc-linux-amd64 ghuc-linux-mips64 ghuc-linux-mips64le
.PHONY: ghuc-linux-arm ghuc-linux-arm-5 ghuc-linux-arm-6 ghuc-linux-arm-7 ghuc-linux-arm64
.PHONY: ghuc-darwin ghuc-darwin-386 ghuc-darwin-amd64
.PHONY: ghuc-windows ghuc-windows-386 ghuc-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

ghuc:
	build/env.sh go run build/ci.go install ./cmd/ghuc
	@echo "Done building."
	@echo "Run \"$(GOBIN)/ghuc\" to launch ghuc."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/ghuc.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Ghuc.framework\" to use the library."

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

ghuc-cross: ghuc-linux ghuc-darwin ghuc-windows ghuc-android ghuc-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-*

ghuc-linux: ghuc-linux-386 ghuc-linux-amd64 ghuc-linux-arm ghuc-linux-mips64 ghuc-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-*

ghuc-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/ghuc
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep 386

ghuc-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/ghuc
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep amd64

ghuc-linux-arm: ghuc-linux-arm-5 ghuc-linux-arm-6 ghuc-linux-arm-7 ghuc-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep arm

ghuc-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/ghuc
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep arm-5

ghuc-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/ghuc
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep arm-6

ghuc-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/ghuc
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep arm-7

ghuc-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/ghuc
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep arm64

ghuc-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/ghuc
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep mips

ghuc-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/ghuc
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep mipsle

ghuc-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/ghuc
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep mips64

ghuc-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/ghuc
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-linux-* | grep mips64le

ghuc-darwin: ghuc-darwin-386 ghuc-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-darwin-*

ghuc-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/ghuc
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-darwin-* | grep 386

ghuc-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/ghuc
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-darwin-* | grep amd64

ghuc-windows: ghuc-windows-386 ghuc-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-windows-*

ghuc-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/ghuc
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-windows-* | grep 386

ghuc-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/ghuc
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/ghuc-windows-* | grep amd64
