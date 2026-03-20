
.PHONY: default
default: precommit

.PHONY: precommit
precommit: ensure format generate test check addlicense
	@echo "ready to commit"

.PHONY: ensure
ensure:
	go mod tidy
	go mod verify
	rm -rf vendor

.PHONY: format
format:
	find . -type f -name 'go.mod' -not -path './vendor/*' -exec go run -mod=mod github.com/shoenig/go-modtool -w fmt "{}" \;
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	go run -mod=mod github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/teamvault-utils -format -excludes vendor ./...
	find . -type d -name vendor -prune -o -type f -name '*.go' -print0 | xargs -0 -n 10 go run -mod=mod github.com/segmentio/golines --max-len=100 -w

.PHONY: generate
generate:
	rm -rf mocks avro
	mkdir -p mocks
	echo "package mocks" > mocks/mocks.go
	go generate -mod=mod ./...

.PHONY: test
test:
	go run -mod=mod github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --race --cover --trace

# TODO: enable lint
# check: lint vet errcheck vulncheck osv-scanner gosec trivy
.PHONY: check
check: vet errcheck vulncheck osv-scanner gosec trivy

.PHONY: lint
lint:
	go run -mod=mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config .golangci.yml ./...

.PHONY: vet
vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: errcheck
errcheck:
	go run -mod=mod github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=mod ./... | grep -v /vendor/ | grep -v k8s/client)

.PHONY: vulncheck
vulncheck:
	go run -mod=mod golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: osv-scanner
osv-scanner:
	@if [ -f .osv-scanner.toml ]; then \
		echo "Using .osv-scanner.toml"; \
		go run -mod=mod github.com/google/osv-scanner/v2/cmd/osv-scanner --config .osv-scanner.toml --recursive .; \
	else \
		echo "No config found, running default scan"; \
		go run -mod=mod github.com/google/osv-scanner/v2/cmd/osv-scanner --recursive .; \
	fi

.PHONY: gosec
gosec:
	go run -mod=mod github.com/securego/gosec/v2/cmd/gosec -exclude=G104 ./...

.PHONY: trivy
trivy:
	trivy fs \
	--db-repository ghcr.io/aquasecurity/trivy-db \
	--scanners vuln,secret \
	--quiet \
	--no-progress \
	--disable-telemetry \
	--exit-code 1 .

.PHONY: addlicense
addlicense:
	go run -mod=mod github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

.PHONY: install
install:
	go build -o $(GOPATH)/bin/teamvault-config-dir-generator cmd/teamvault-config-dir-generator/*
	go build -o $(GOPATH)/bin/teamvault-config-parser cmd/teamvault-config-parser/*
	go build -o $(GOPATH)/bin/teamvault-password cmd/teamvault-password/*
	go build -o $(GOPATH)/bin/teamvault-url cmd/teamvault-url/*
	go build -o $(GOPATH)/bin/teamvault-username cmd/teamvault-username/*
	go build -o $(GOPATH)/bin/teamvault-file cmd/teamvault-file/*
