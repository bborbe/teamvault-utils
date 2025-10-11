
all: precommit install

install:
	go build -o $(GOPATH)/bin/teamvault-config-dir-generator cmd/teamvault-config-dir-generator/*
	go build -o $(GOPATH)/bin/teamvault-config-parser cmd/teamvault-config-parser/*
	go build -o $(GOPATH)/bin/teamvault-password cmd/teamvault-password/*
	go build -o $(GOPATH)/bin/teamvault-url cmd/teamvault-url/*
	go build -o $(GOPATH)/bin/teamvault-username cmd/teamvault-username/*
	go build -o $(GOPATH)/bin/teamvault-file cmd/teamvault-file/*

default: precommit

precommit: ensure format generate test check addlicense
	@echo "ready to commit"

ensure:
	go mod tidy
	go mod verify
	rm -rf vendor

format:
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	go run -mod=mod github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/teamvault-utils -format -excludes vendor ./...
	find . -type d -name vendor -prune -o -type f -name '*.go' -print0 | xargs -0 -n 10 go run -mod=mod github.com/segmentio/golines --max-len=100 -w

generate:
	rm -rf mocks avro
	go generate -mod=mod ./...

.PHONY: test
test:
	# -race
	go test -mod=mod -p=$${GO_TEST_PARALLEL:-1} -cover $(shell go list -mod=mod ./... | grep -v /vendor/)

# TODO: enable lint
# check: lint vet errcheck vulncheck osv-scanner gosec trivy
check: vet errcheck vulncheck osv-scanner gosec trivy

vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)

errcheck:
	go run -mod=mod github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=mod ./... | grep -v /vendor/ | grep -v k8s/client)

vulncheck:
	go run -mod=mod golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=mod ./... | grep -v /vendor/)

osv-scanner:
	go run -mod=mod github.com/google/osv-scanner/cmd/osv-scanner -r .

gosec:
	go run -mod=mod github.com/securego/gosec/v2/cmd/gosec -exclude=G104 ./...

trivy:
	trivy fs --scanners vuln,secret --quiet --no-progress --disable-telemetry --exit-code 1 .

lint:
	go run -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint run --config .golangci.yml ./...

addlicense:
	go run -mod=mod github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')
