
default: precommit

install:
	go build -o $(GOPATH)/bin/teamvault-config-dir-generator cmd/teamvault-config-dir-generator/*
	go build -o $(GOPATH)/bin/teamvault-config-parser cmd/teamvault-config-parser/*
	go build -o $(GOPATH)/bin/teamvault-password cmd/teamvault-password/*
	go build -o $(GOPATH)/bin/teamvault-url cmd/teamvault-url/*
	go build -o $(GOPATH)/bin/teamvault-username cmd/teamvault-username/*
	go build -o $(GOPATH)/bin/teamvault-file cmd/teamvault-file/*

precommit: ensure format generate test check
	@echo "ready to commit"

ensure:
	go mod tidy
	go mod verify
	rm -rf vendor

format:
	go run -mod=mod github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/teamvault-utils -format -excludes vendor ./...

generate:
	rm -rf mocks avro
	go generate -mod=mod ./...

test:
	go test -mod=mod -p=1 -cover -race $(shell go list -mod=mod ./... | grep -v /vendor/)

check: vet errcheck vulncheck

vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)

errcheck:
	go run -mod=mod github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=mod ./... | grep -v /vendor/)

addlicense:
	go run -mod=mod github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

vulncheck:
	go run -mod=mod golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=mod ./... | grep -v /vendor/)
