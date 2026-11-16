BINARY  := bin/repora
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

PLATFORMS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64
SHA256    ?= $(shell command -v sha256sum >/dev/null 2>&1 && echo sha256sum || echo shasum -a 256)

.PHONY: build run test race vet fmt fmt-check lint gate cover dist clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/repora

run:
	go run ./cmd/repora

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

fmt-check:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "unformatted files:"; echo "$$out"; exit 1; fi

lint: fmt-check vet

# Everything CI runs. Keep this and .github/workflows/ci.yml in sync.
gate: fmt-check vet test race

cover:
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out

dist:
	@rm -rf dist
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os="$${platform%/*}"; arch="$${platform#*/}"; \
		out="dist/repora_$${os}_$${arch}"; \
		echo "building $$out"; \
		GOOS="$$os" GOARCH="$$arch" go build -trimpath -ldflags "$(LDFLAGS)" -o "$$out" ./cmd/repora || exit 1; \
	done
	@cd dist && $(SHA256) repora_* > SHA256SUMS && cat SHA256SUMS
	@echo "version: $(VERSION)"

clean:
	rm -rf bin dist cover.out
