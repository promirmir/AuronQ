.PHONY: all test vet build smoke integration clean

all: test vet build

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o bin/auronq ./cmd/auronq

smoke: build
	./scripts/smoke-test.sh ./bin/auronq

integration: build
	./scripts/two-node-test.sh ./bin/auronq

clean:
	rm -rf bin .smoke .two-node
