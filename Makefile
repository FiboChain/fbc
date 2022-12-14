maDEP := $(shell command -v dep 2> /dev/null)
SUM := $(shell which shasum)

COMMIT := $(shell git rev-parse HEAD)
CAT := $(if $(filter $(OS),Windows_NT),type,cat)
export GO111MODULE=on

GithubTop=github.com



Version=v1.2.2
CosmosSDK=v0.39.2
Tendermint=v0.33.9
Iavl=v0.14.3
Name=fbchain
ServerName=fbchaind
ClientName=fbchaincli
# the height of the 1st block is GenesisHeight+1
GenesisHeight=0
MercuryHeight=1
VenusHeight=1

# process linker flags
ifeq ($(VERSION),)
    VERSION = $(COMMIT)
endif

build_tags = netgo

ifeq ($(WITH_ROCKSDB),true)
  CGO_ENABLED=1
  build_tags += rocksdb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))


ifeq ($(MAKECMDGOALS),mainnet)
   GenesisHeight=2322600
   MercuryHeight=5150000
   VenusHeight=8200000
else ifeq ($(MAKECMDGOALS),testnet)
   GenesisHeight=1121818
   MercuryHeight=5300000
   VenusHeight=8510000
endif

ldflags = -X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.Version=$(Version) \
	-X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.Name=$(Name) \
  -X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.ServerName=$(ServerName) \
  -X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.ClientName=$(ClientName) \
  -X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.Commit=$(COMMIT) \
  -X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.CosmosSDK=$(CosmosSDK) \
  -X $(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.Tendermint=$(Tendermint) \
  -X "$(GithubTop)/FiboChain/fbc/libs/cosmos-sdk/version.BuildTags=$(build_tags)" \
  -X $(GithubTop)/FiboChain/fbc/libs/tendermint/types.MILESTONE_GENESIS_HEIGHT=$(GenesisHeight) \
  -X $(GithubTop)/FiboChain/fbc/libs/tendermint/types.MILESTONE_MERCURY_HEIGHT=$(MercuryHeight) \
  -X $(GithubTop)/FiboChain/fbc/libs/tendermint/types.MILESTONE_VENUS_HEIGHT=$(VenusHeight)

ifeq ($(WITH_ROCKSDB),true)
  ldflags += -X github.com/FiboChain/fbc/libs/cosmos-sdk/types.DBBackend=rocksdb
endif

BUILD_FLAGS := -ldflags '$(ldflags)'

ifeq ($(DEBUG),true)
	BUILD_FLAGS += -gcflags "all=-N -l"
endif

all: install

install: fbchain

fbchain:
	go install -v $(BUILD_FLAGS) -tags "$(build_tags)" ./cmd/fbchaind
	go install -v $(BUILD_FLAGS) -tags "$(build_tags)" ./cmd/fbchaincli

mainnet: fbchain

testnet: fbchain

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./app/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/backend/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/common/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/dex/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/distribution/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/genutil/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/gov/...
#	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/order/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/params/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/staking/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/token/...
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./x/upgrade/...

get_vendor_deps:
	@echo "--> Generating vendor directory via dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -vendor-only

update_vendor_deps:
	@echo "--> Running dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -update

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
.PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

cli:
	go install -v $(BUILD_FLAGS) -tags "$(build_tags)" ./cmd/fbchaincli

server:
	go install -v $(BUILD_FLAGS) -tags "$(build_tags)" ./cmd/fbchaind

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s

build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -tags "$(build_tags)" -o build/fbchaind.exe ./cmd/fbchaind
	go build $(BUILD_FLAGS) -tags "$(build_tags)" -o build/fbchaincli.exe ./cmd/fbchaincli
else
	go build $(BUILD_FLAGS) -tags "$(build_tags)" -o build/fbchaind ./cmd/fbchaind
	go build $(BUILD_FLAGS) -tags "$(build_tags)" -o build/fbchaincli ./cmd/fbchaincli
endif

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

rocksdb:
	@echo "Installing rocksdb..."
	@bash ./libs/rocksdb/install.sh
.PHONY: rocksdb

.PHONY: build
